package app

import (
	"strconv"

	"github.com/getbread/breadkit/desmond"
	"github.com/getbread/shopify_plugin_backend/service/bread"
	"github.com/getbread/shopify_plugin_backend/service/shopify"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type NewOrderTransactionRequest struct {
	Id      int    `json:"id"`
	OrderId int    `json:"order_id"`
	Gateway string `json:"gateway"`
	Kind    string `json:"kind"`
	Amount  string `json:"amount"`
	Status  string `json:"status"`
}

func (h *Handlers) NewOrderTransaction(c *gin.Context, dc desmond.Context) {
	c.String(200, "complete")

	var req NewOrderTransactionRequest
	if err := c.BindJSON(&req); err != nil {
		log.WithFields(log.Fields{
			"error":       err.Error(),
			"queryString": c.Request.URL.RawQuery,
		}).Error("(NewOrderTransaction) binding request to model produced error")
		return
	}

	// Check if order was created by Bread App
	if !isBreadAppOrder(req.Gateway) && !isBreadPOSOrder(req.Gateway) {
		return
	}

	// query order to get transaction_id
	order, err := findOrderByOrderId(req.OrderId, h)
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err.Error(),
			"request": req,
			"orderId": req.OrderId,
			"gateway": req.Gateway,
			"kind":    req.Kind,
		}).Info("(NewOrderTransaction) order not found")
		return
	}

	// query shop
	shop, err := findShopById(order.ShopId, h)
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err.Error(),
			"request": req,
			"order":   order,
		}).Error("(NewOrderTransaction) query for shop produced error")
		return
	}

	// instantiate clients
	sc := shopify.NewClient(shop.Shop, shop.AccessToken)
	bc := bread.NewClient(shop.GetAPIKeys())

	// query shopify transaction
	var searchTransactionRes shopify.SearchTransactionResponse
	err = sc.QueryTransaction(strconv.Itoa(req.OrderId), strconv.Itoa(req.Id), &searchTransactionRes)
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err.Error(),
			"request": req,
			"order":   order,
			"shop":    shop,
		}).Error("(NewOrderTransaction) query for Shopify transaction produced error")
		return
	}
	st := searchTransactionRes.Transaction

	// query bread transaction
	bt, err := bc.QueryTransaction(string(order.TxId), order.BreadHost())
	if err != nil {
		log.WithFields(log.Fields{
			"error":              err.Error(),
			"request":            req,
			"order":              order,
			"shop":               shop,
			"shopifyTransaction": st,
		}).Error("(NewOrderTransaction) query for Bread transaction produced error")
		return
	}

	// make Bread & Shopify amounts comparable
	bamount := float64(bt.AdjustedTotal) / 100.00
	samount, err := strconv.ParseFloat(st.Amount, 64)

	// search transaction res
	switch searchTransactionRes.Transaction.Kind {
	case "authorization":
		if bt.Status == "AUTHORIZED" {
			break
		}

		authorizeRequest := &bread.TransactionActionRequest{
			Type: "authorize",
		}
		if bamount != samount {
			authorizeAmountCentsFloat := samount * 100
			authorizeRequest.Amount = int(authorizeAmountCentsFloat)
		}
		_, err = bc.AuthorizeTransaction(bt.BreadTransactionId, order.BreadHost(), authorizeRequest)
		if err != nil {
			log.WithFields(log.Fields{
				"error":              err.Error(),
				"request":            req,
				"order":              order,
				"shop":               shop,
				"shopifyTransaction": st,
				"transactionID":      bt.BreadTransactionId,
			}).Info("(NewOrderTransaction) authorizing transaction produced an error")
			return
		}
	case "capture":
		log.Println("(NewOrderTransaction) CAPTURE")
		if bamount == samount {
			settleRequest := &bread.TransactionActionRequest{
				Type: "settle",
			}
			_, err = bc.SettleTransaction(bt.BreadTransactionId, order.BreadHost(), settleRequest)
			if err != nil {
				log.WithFields(log.Fields{
					"error":              err.Error(),
					"request":            req,
					"order":              order,
					"shop":               shop,
					"shopifyTransaction": st,
					"transactionID":      bt.BreadTransactionId,
				}).Error("(NewOrderTransaction) settling transaction produced error")
				return
			}
		} else {
			// do a partial cancel
			if samount > bamount {
				log.WithFields(log.Fields{
					"error":              "settle amount > transaction amount",
					"request":            req,
					"order":              order,
					"shop":               shop,
					"shopifyTransaction": st,
					"transactionID":      bt.BreadTransactionId,
				}).Error("(NewOrderTransaction) settle amount > transaction amount")
				return
			}
			cancelAmountCentsFloat := (bamount - samount) * 100.00
			cancelAmountCents := int(cancelAmountCentsFloat)
			cancelRequest := &bread.TransactionActionRequest{
				Type:   "cancel",
				Amount: cancelAmountCents,
			}
			_, err = bc.CancelTransaction(bt.BreadTransactionId, order.BreadHost(), cancelRequest)
			if err != nil {
				log.WithFields(log.Fields{
					"error":              err.Error(),
					"request":            req,
					"order":              order,
					"shop":               shop,
					"shopifyTransaction": st,
					"transactionID":      bt.BreadTransactionId,
				}).Error("(NewOrderTransaction) partial cancel before partial settle produced error")
				return
			}

			// then do a full settle
			settleRequest := &bread.TransactionActionRequest{
				Type: "settle",
			}
			_, err = bc.SettleTransaction(bt.BreadTransactionId, order.BreadHost(), settleRequest)
			if err != nil {
				log.WithFields(log.Fields{
					"error":              err.Error(),
					"request":            req,
					"order":              order,
					"shop":               shop,
					"shopifyTransaction": st,
					"transactionID":      bt.BreadTransactionId,
				}).Error("(NewOrderTransaction) partial settle after partial concel produced error")
				return
			}
		}
	case "void":
		log.Println("(NewOrderTransaction) VOID")
	case "refund":
		log.Println("(NewOrderTransaction) REFUND")
		refundAmountCentsFloat := samount * 100.00
		refundAmountCents := int(refundAmountCentsFloat)
		refundRequest := &bread.TransactionActionRequest{
			Type:   "refund",
			Amount: refundAmountCents,
		}
		_, err = bc.RefundTransaction(bt.BreadTransactionId, order.BreadHost(), refundRequest)
		if err != nil {
			log.WithFields(log.Fields{
				"error":              err.Error(),
				"request":            req,
				"order":              order,
				"shop":               shop,
				"shopifyTransaction": st,
				"transactionID":      bt.BreadTransactionId,
			}).Error("(NewOrderTransaction) refunding transaction produced error")
			return
		}
	case "sale":
		log.Println("new SALE transaction")
	default:
		log.WithFields(log.Fields{
			"request":            req,
			"order":              order,
			"shop":               shop,
			"shopifyTransaction": st,
			"transactionID":      bt.BreadTransactionId,
		}).Error("(NewOrderTransaction) switch miss on transaction.Kind")
	}
}
