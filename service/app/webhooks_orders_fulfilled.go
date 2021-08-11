package app

import (
	"strconv"
	"strings"

	"github.com/getbread/breadkit/desmond"
	"github.com/getbread/shopify_plugin_backend/service/bread"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type OrdersFulfilledRequest struct {
	ID                int                  `json:"id"`
	CheckoutID        int                  `json:"checkout_id"`
	Gateway           string               `json:"gateway"`
	Status            string               `json:"status"`
	Test              bool                 `json:"test"`
	FulfillmentStatus string               `json:"fulfillment_status"`
	Fulfillments      []ShopifyFulfillment `json:"fulfillments"`
}

type ShopifyFulfillment struct {
	ID              int      `json:"id"`
	OrderID         int      `json:"order_id"`
	Status          string   `json:"status"`
	ShipmentStatus  string   `json:"shipment_status"`
	TrackingCompany string   `json:"tracking_company"`
	TrackingNumber  string   `json:"tracking_number"`
	TrackingNumbers []string `json:"tracking_numbers"`
	TrackingURL     string   `json:"tracking_url"`
	TrackingURLs    []string `json:"tracking_urls"`
}

func (h *Handlers) OrdersFulfilled(c *gin.Context, dc desmond.Context) {
	c.String(200, "success")

	shopDomain := c.Request.Header.Get("X-Shopify-Shop-Domain")
	shopName := strings.Split(shopDomain, ".")[0]

	var r OrdersFulfilledRequest
	err := c.BindJSON(&r)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":       err.Error(),
			"shop":        shopName,
			"queryString": c.Request.URL.RawQuery,
		}).Error("(WebhookFulfillmentsCreate) binding request to model produced error")
		return
	}

	go processOrderFulfillment(shopName, r, h)
	return
}

func processOrderFulfillment(shopName string, r OrdersFulfilledRequest, h *Handlers) {
	shop, err := findShopByName(shopName, h)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
			"shop":  shopName,
		}).Error("(WebhookFulfillmentsCreate) looking up shop produced an error")
		return
	}

	if len(r.Fulfillments) == 0 {
		logrus.WithFields(logrus.Fields{
			"shop":              shopName,
			"orderID":           r.ID,
			"fulfillmentStatus": r.FulfillmentStatus,
		}).Info("(WebhookFulfillmentsCreate) no fulfillments found")
		return
	}

	var transactionID, trackingCompany, trackingNumber string
	trackingCompany = r.Fulfillments[0].TrackingCompany
	trackingNumber = r.Fulfillments[0].TrackingNumber

	if trackingCompany == "" || trackingNumber == "" {
		logrus.WithFields(logrus.Fields{
			"shop":              shopName,
			"orderID":           r.ID,
			"checkoutID":        r.CheckoutID,
			"fulfillmentStatus": r.FulfillmentStatus,
			"carrierName":       trackingCompany,
			"trackingNumber":    trackingNumber,
		}).Info("(WebhookFulfillmentsCreate) missing tracking number or shipping carrier, skipping tx update")
		return
	}

	// Check if Order was completed with Bread
	if isBreadAppOrder(r.Gateway) || isBreadPOSOrder(r.Gateway) {
		// For App and POS orders, query shopify_shops_orders by order_id to find Bread transaction ID
		order, err := findOrderByOrderId(r.ID, h)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error":   err.Error(),
				"orderID": r.ID,
				"shop":    shopName,
			}).Error("(WebhookFulfillmentsCreate) looking up order produced an error")
			return
		}

		transactionID = string(order.TxId)
		var apiKey, secretKey, host string
		if order.Production {
			apiKey = shop.BreadApiKey
			secretKey = shop.BreadSecretKey
			host = appConfig.HostConfig.BreadHost
		} else {
			apiKey = shop.BreadSandboxApiKey
			secretKey = shop.BreadSandboxSecretKey
			host = appConfig.HostConfig.BreadHostDevelopment
		}

		bc := bread.NewClient(apiKey, secretKey)
		shipmentRequest := bread.TransactionShipmentRequest{
			CarrierName:    trackingCompany,
			TrackingNumber: trackingNumber,
		}

		_, err = bc.SetShippingDetails(transactionID, host, shipmentRequest)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error":         err.Error(),
				"transactionID": transactionID,
				"apiKey":        apiKey,
				"shop":          shopName,
			}).Error("(WebhookFulfillmentsCreate) updating transaction with shipment details produced an error")
		}
		return
	}

	if isBreadGateway(r.Gateway) {
		// For Gateway orders, query shopify_gateway_checkouts by checkout_id to find Bread transaction ID
		gc, err := findCompletedGatewayCheckoutByReference(strconv.Itoa(r.CheckoutID), h)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error":      err.Error(),
				"orderID":    r.ID,
				"shop":       shopName,
				"checkoutID": r.CheckoutID,
			}).Error("(WebhookFulfillmentsCreate) looking up gateay checkout produced an error")
			return
		}

		if gc.TransactionID == "" {
			logrus.WithFields(logrus.Fields{
				"gateway":           r.Gateway,
				"shop":              shopName,
				"orderID":           r.ID,
				"checkoutID":        r.CheckoutID,
				"accountID":         gc.AccountID,
				"gatewayCheckoutID": gc.Id,
			}).Error("(WebhookFulfillmentsCreate) gateway checkout contains empty transactionID")
			return
		}

		account, err := findGatewayAccountById(gc.AccountID, h)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error":      err.Error(),
				"orderID":    r.ID,
				"shop":       shopName,
				"account_id": gc.AccountID,
			}).Error("(WebhookFulfillmentsCreate) search for gateway account produced error")
			return
		}

		transactionID = gc.TransactionID
		var apiKey, secretKey, host string
		if gc.Test {
			apiKey = account.SandboxApiKey
			secretKey = account.SandboxSharedSecret
			host = appConfig.HostConfig.BreadHostDevelopment
		} else {
			apiKey = account.ApiKey
			secretKey = account.SharedSecret
			host = appConfig.HostConfig.BreadHost
		}

		bc := bread.NewClient(apiKey, secretKey)
		shipmentRequest := bread.TransactionShipmentRequest{
			CarrierName:    trackingCompany,
			TrackingNumber: trackingNumber,
		}

		_, err = bc.SetShippingDetails(transactionID, host, shipmentRequest)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error":         err.Error(),
				"transactionID": transactionID,
				"apiKey":        apiKey,
				"shop":          shopName,
			}).Error("(WebhookFulfillmentsCreate) updating transaction with shipment details produced an error")
		}
		return
	}
	return
}
