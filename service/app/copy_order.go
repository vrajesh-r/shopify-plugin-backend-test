package app

import (
	"strconv"
	"strings"

	"github.com/getbread/breadkit/desmond"
	"github.com/getbread/shopify_plugin_backend/service/bread"
	"github.com/getbread/shopify_plugin_backend/service/shopify"
	"github.com/getbread/shopify_plugin_backend/service/types"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type copyOrderRequest struct {
	TransactionId string `json:"transactionId"`
	Referrer      string `json:"referrer"`
}

func (h *Handlers) CopyOrder(c *gin.Context, dc desmond.Context) {
	// parse body into request struct
	var req copyOrderRequest
	err := c.BindJSON(&req)
	if err != nil {
		log.WithFields(log.Fields{
			"error":       err.Error(),
			"queryString": c.Request.URL.RawQuery,
		}).Error("(CopyOrder) binding request to produced an error")
		c.String(400, err.Error())
		return
	}

	// pull data from query params
	shopUrl := c.Query("shop")
	shopName := strings.Split(shopUrl, ".")[0]
	// pull shop
	shop, err := findShopByName(shopName, h)
	if err != nil {
		log.WithFields(log.Fields{
			"error":       err.Error(),
			"queryString": c.Request.URL.RawQuery,
			"request":     req,
		}).Error("(CopyOrder) querying shop by name produced error")
		c.String(400, err.Error())
		return
	}

	// http clients
	bc := bread.NewClient(shop.GetAPIKeys())

	// pull transaction from Ostia with transaction_id & Bread credentials
	bt, err := bc.QueryTransaction(req.TransactionId, shop.BreadHost())
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err.Error(),
			"request": req,
			"shop":    shop,
		}).Error("(CopyOrder) query for Bread transaction produced error")
		c.String(400, err.Error())
		return
	}

	// Cancel order and transaction if merchant is not grandfathered into accelerated checkout
	if !shop.AcceleratedCheckoutPermitted {
		go cancelOrRefundTransaction(shop, bt)
		log.WithFields(log.Fields{
			"shop":                          shop.Shop,
			"embedScript":                   shop.ManualEmbedScript,
			"production":                    shop.Production,
			"allowCheckoutPDP":              shop.AllowCheckoutPDP,
			"allowCheckoutCart":             shop.AllowCheckoutCart,
			"acceleratedCheckoutPermission": shop.AcceleratedCheckoutPermitted,
			"transactionID":                 req.TransactionId,
		}).Error("(CopyOrder) shop does not have permission to use accelerated checkout")
		c.String(403, "Forbidden")
		return
	}

	var remainderPayDecline = false

	// auto_authorize transaction
	authorizeRequest := &bread.TransactionActionRequest{
		Type: "authorize",
	}
	_, err = bc.AuthorizeTransaction(bt.BreadTransactionId, shop.BreadHost(), authorizeRequest)
	if err != nil {
		_, err := bc.AuthorizeTransaction(bt.BreadTransactionId, shop.BreadHost(), authorizeRequest)
		if err != nil {
			remainderPayDecline = strings.Contains(err.Error(), "There's an issue with authorizing the credit card portion")

			if remainderPayDecline && shop.RemainderPayAutoCancel {
				go cancelOrRefundTransaction(shop, bt)
			}

			log.WithFields(log.Fields{
				"error":         err.Error(),
				"request":       req,
				"shop":          shop,
				"transactionID": bt.BreadTransactionId,
			}).Error("(CopyOrder) authorizing transaction produced error")
			var response string
			if remainderPayDecline {
				response = "remainderPay"
			} else {
				response = err.Error()
			}
			c.String(400, response)
			return
		}
	}

	// create customer on shopify backend
	customer, err := getShopifyCustomer(&bt.BillingContact, &bt.ShippingContact, shop)
	if err != nil {
		log.WithFields(log.Fields{
			"error":         err.Error(),
			"request":       req,
			"shop":          shop,
			"transactionID": bt.BreadTransactionId,
		}).Error("(CopyOrder) creating Shopify customer produced error")
		c.String(400, err.Error())
		return
	}

	// copy Bread transaction to Shopify backend
	so, err := createShopifyOrder(bt, customer, shop)
	if err != nil {
		log.WithFields(log.Fields{
			"error":         err.Error(),
			"request":       req,
			"shop":          shop,
			"transactionID": bt.BreadTransactionId,
		}).Error("(CopyOrder) creating Shopify order produced error")
		c.String(400, err.Error())
		return
	}

	// Update transaction with Shopify order number
	updateRequest := &bread.TransactionActionRequest{
		MerchantOrderId: strconv.Itoa(so.OrderNumber),
	}
	_, err = bc.UpdateTransaction(bt.BreadTransactionId, shop.BreadHost(), updateRequest)
	if err != nil {
		log.WithFields(log.Fields{
			"error":              err.Error(),
			"request":            req,
			"shop":               shop,
			"transactionID":      bt.BreadTransactionId,
			"shopifyOrderNumber": so.OrderNumber,
			"shopifyOrderID":     so.ID,
		}).Error("(CopyOrder) updating transaction with Shopify order ID produced error")
	}

	// add order to Milton order -> transaction lookup
	_, err = createOrder(shop, bt.BreadTransactionId, so.ID, h)
	if err != nil {
		log.WithFields(log.Fields{
			"error":          err.Error(),
			"request":        req,
			"shop":           shop,
			"transactionID":  bt.BreadTransactionId,
			"shopifyOrderID": so.ID,
		}).Error("(CopyOrder) creating Milton order produced error")
		c.String(400, err.Error())
		return
	}

	// auto_settle transaction
	if shop.AutoSettle {
		sc := shopify.NewClient(shop.Shop, shop.AccessToken)

		captureRequest := &shopify.CreateTransactionRequest{
			Transaction: shopify.Transaction{
				Kind:     "capture",
				Status:   "success",
				Amount:   types.Cents(bt.AdjustedTotal).ToString(),
				Currency: "USD",
				Gateway:  "Bread Shopify Payments",
				Test:     !shop.Production,
			},
		}

		settled := true

		var captureRes shopify.CreateTransactionResponse
		if err := sc.CreateTransaction(so.ID, captureRequest, &captureRes); err != nil {
			// log and continue with response
			log.WithFields(log.Fields{
				"error":          err.Error(),
				"request":        req,
				"shop":           shop,
				"transactionID":  bt.BreadTransactionId,
				"shopifyOrderID": so.ID,
			}).Error("(CopyOrder) creating Shopify capture transaction produced error")

			settled = false
		}

		if settled && !shop.EnableOrderWebhooks {
			// We need to push the settled transaction to Bread; it will not happen automatically.
			settleRequest := &bread.TransactionActionRequest{
				Type: "settle",
			}

			_, err := bc.SettleTransaction(bt.BreadTransactionId, shop.BreadHost(), settleRequest)
			if err != nil {
				log.WithFields(log.Fields{
					"error":          err.Error(),
					"request":        req,
					"shop":           shop,
					"transactionID":  bt.BreadTransactionId,
					"shopifyOrderID": so.ID,
				}).Error("(CopyOrder) Failed to settle a transaction via Ostia.")
			}
		}
	}

	// quickly respond to Shopify webhook
	c.JSON(200, gin.H{
		"orderId": strconv.Itoa(so.ID),
	})
}

func cancelOrRefundTransaction(shop types.Shop, bt *bread.TransactionResponse) {
	bc := bread.NewClient(shop.GetAPIKeys())
	transactionRequest := &bread.TransactionActionRequest{
		Amount: bt.AdjustedTotal,
	}

	if bt.Status == "PENDING" || bt.Status == "AUTHORIZED" {
		transactionRequest.Type = "cancel"
		_, err := bc.CancelTransaction(bt.BreadTransactionId, shop.BreadHost(), transactionRequest)
		if err != nil {
			log.WithFields(log.Fields{
				"error":         err.Error(),
				"shop":          shop.Shop,
				"transactionID": bt.BreadTransactionId,
			}).Error("(CopyOrder) canceling transaction produced error")
		}
		return
	}

	if bt.Status == "SETTLED" {
		transactionRequest.Type = "refund"
		_, err := bc.RefundTransaction(bt.BreadTransactionId, shop.BreadHost(), transactionRequest)
		if err != nil {
			log.WithFields(log.Fields{
				"error":         err.Error(),
				"shop":          shop.Shop,
				"transactionID": bt.BreadTransactionId,
			}).Error("(CopyOrder) refunding transaction produced error")
		}
		return
	}
}
