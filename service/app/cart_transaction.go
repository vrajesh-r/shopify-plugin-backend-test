package app

import (
	"errors"
	"strconv"

	"github.com/getbread/breadkit/desmond"
	"github.com/getbread/breadkit/zeus/types"
	"github.com/getbread/shopify_plugin_backend/service/bread"
	"github.com/getbread/shopify_plugin_backend/service/shopify"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type sendOrderShopifyRequest struct {
	TransactionID string `json:"transactionId"`
	OrderRef      string `json:"orderRef"`
	APIKey        string `json:"apiKey"`
}

// SendOrderShopify sends the processed order from a cart link to Shopify so
// that is appears in their records.
func (h *Handlers) SendOrderShopify(c *gin.Context, dc desmond.Context) {
	var breadTransaction *bread.TransactionResponse
	var req sendOrderShopifyRequest
	var customer *shopify.Customer
	var shopifyOrder *shopify.Order
	if err := c.BindJSON(&req); err != nil {
		log.WithFields(log.Fields{
			"error":       err.Error(),
			"queryString": c.Request.URL.RawQuery,
		}).Error("(SendOrderShopify) binding request produced an error")
		c.String(400, err.Error())
		return
	}

	//Check to see if we satisfied this request already
	order, err := findOrderByTransactionId(req.TransactionID, h)
	if order != nil && err == nil {
		//Means the order exists so there is no need to run through this again.
		c.JSON(200, gin.H{
			"orderID": strconv.Itoa(order.OrderId),
		})
		return
	}

	if len(req.TransactionID) == 0 || len(req.APIKey) == 0 {
		err := errors.New("(SendOrderShopify) invalid trasaction or api key")
		c.String(400, err.Error())
	}

	merchantShop, err := findShopByBreadApiKey(req.APIKey, h)
	if err != nil {
		log.WithFields(log.Fields{
			"error":       err.Error(),
			"queryString": c.Request.URL.RawQuery,
			"request":     req,
		}).Error("(SendOrderShopify) querying shop by api key produced error")
		c.String(400, err.Error())
		return
	}
	breadClient := bread.NewClient(merchantShop.BreadApiKey,
		merchantShop.BreadSecretKey)
	// Use merchantShop.BreadHost() to get the correct one. https://api-dev-sandbox.getbread.com
	breadTransaction, err = breadClient.QueryTransaction(req.TransactionID,
		merchantShop.BreadHost())
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err.Error(),
			"request": req,
			"shop":    merchantShop,
		}).Error("(SendOrderShopify) query for Bread transaction produced error")
		c.String(400, err.Error())
		return
	}

	//Obtain shopify customer
	customer, err = getShopifyCustomer(&breadTransaction.BillingContact,
		&breadTransaction.ShippingContact, merchantShop)
	if err != nil {
		log.WithFields(log.Fields{
			"error":         err.Error(),
			"request":       req,
			"shop":          merchantShop,
			"transactionID": breadTransaction.BreadTransactionId,
		}).Error("(SnedOrderShopify) creating Shopify customer produced an error")
		c.String(400, err.Error())
		return
	}
	// Copy transacation to backend
	shopifyOrder, err = createShopifyOrder(breadTransaction, customer, merchantShop)
	if err != nil {
		log.WithFields(log.Fields{
			"error":           err.Error(),
			"request":         req,
			"shop":            merchantShop,
			"transactionID":   breadTransaction.BreadTransactionId,
			"shopifyCustomer": customer,
		}).Error("(SendOrderShopify) creating Shopify order returned an error")
		c.String(400, err.Error())
		return
	}
	// add order to Milton
	_, err = createOrder(merchantShop, breadTransaction.BreadTransactionId,
		shopifyOrder.ID, h)
	if err != nil {
		log.WithFields(log.Fields{
			"error":           err.Error(),
			"request":         req,
			"shop":            merchantShop,
			"transactionID":   breadTransaction.BreadTransactionId,
			"shopifyCustomer": customer,
			"shopifyOrder":    shopifyOrder,
		}).Error("(SendOrderShopify) creating Milton order returned an error")
		c.String(400, err.Error())
		return
	}
	// auto_settle the transaction in necessary
	if merchantShop.AutoSettle {
		shopifyClient := shopify.NewClient(merchantShop.Shop, merchantShop.AccessToken)

		captureReq := shopify.CreateTransactionRequest{
			Transaction: shopify.Transaction{
				Kind:     "capture",
				Status:   "success",
				Amount:   types.Cents(breadTransaction.AdjustedTotal).ToString(),
				Currency: "USD",
				Gateway:  "Bread Shopify Payments",
				Test:     !merchantShop.Production,
			},
		}
		var captureRes shopify.CreateTransactionResponse
		if err = shopifyClient.CreateTransaction(shopifyOrder.ID, &captureReq,
			&captureRes); err != nil {
			log.WithFields(log.Fields{
				"error":         err.Error(),
				"request":       req,
				"shop":          merchantShop,
				"transactionID": breadTransaction.BreadTransactionId,
				"shopifyOrder":  shopifyOrder,
			}).Error("(SendOrderShopify) creating Shopify capture transaction produced and error")
		}
	}
	// Response to the Shopify webhook
	c.JSON(200, gin.H{
		"orderID": strconv.Itoa(shopifyOrder.ID),
	})
}
