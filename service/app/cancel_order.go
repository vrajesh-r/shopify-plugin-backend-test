package app

import (
	"github.com/getbread/breadkit/desmond"
	"github.com/getbread/shopify_plugin_backend/service/bread"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type CancelOrderRequest struct {
	Id      int    `json:"id"`
	Gateway string `json:"gateway"`
}

func (h *Handlers) CancelOrder(c *gin.Context, dc desmond.Context) {
	c.String(200, "done")

	var req CancelOrderRequest
	if err := c.BindJSON(&req); err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
			"req":   req,
		}).Error("(CancelOrder) binding request to model produced error")
		return
	}

	// Check if order was created by Bread App
	if !isBreadAppOrder(req.Gateway) && !isBreadPOSOrder(req.Gateway) {
		return
	}

	// query bread-shopify order to get transaction_id
	order, err := findOrderByOrderId(req.Id, h)
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err.Error(),
			"request": req,
		}).Info("(CancelOrder) query for order produced error")
		return
	}

	// query shop
	shop, err := findShopById(order.ShopId, h)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
			"order": order,
		}).Error("(CancelOrder) query for shop produced error")
		return
	}

	// bread client
	bc := bread.NewClient(shop.GetAPIKeys())

	// query order on Ostia
	bt, err := bc.QueryTransaction(string(order.TxId), order.BreadHost())
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
			"order": order,
			"shop":  shop,
		}).Error("(CancelOrder) query for Bread transaction produced error")
		return
	}

	if bt.Status == "AUTHORIZED" || bt.Status == "PENDING" {
		cancelRequest := &bread.TransactionActionRequest{
			Type: "cancel",
		}
		ctr, err := bc.CancelTransaction(string(order.TxId), order.BreadHost(), cancelRequest)
		if err != nil {
			log.WithFields(log.Fields{
				"error":         err.Error(),
				"order":         order,
				"shop":          shop,
				"transactionID": bt.BreadTransactionId,
			}).Error("(CancelOrder) cancelling Bread order produced error")
			return
		}
		log.WithField("ostiaResponse", ctr).Info("(CancelOrder) order successfully cancelled")
	} else {
		log.WithFields(log.Fields{
			"transactionID": bt.BreadTransactionId,
		}).Info("(CancelOrder) skipped cancelling order, transaction past AUTHORIZED state")
	}
}
