package app

import (
	"strings"

	"github.com/getbread/breadkit/desmond"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type CustomerRedactRequest struct {
	ShopID         int            `json:"shop_id"`
	ShopDomain     string         `json:"shop_domain"`
	Customer       CustomerRecord `json:"customer"`
	OrdersToRedact []int          `json:"orders_to_redact"`
}

type ShopRedactRequest struct {
	ShopID     int    `json:"shop_id"`
	ShopDomain string `json:"shop_domain"`
}

type CustomerDataRequest struct {
	ShopID          int            `json:"shop_id"`
	ShopDomain      string         `json:"shop_domain"`
	Customer        CustomerRecord `json:"customer"`
	OrdersRequested []int          `json:"orders_requested"`
}

type CustomerRecord struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

func (h *Handlers) RedactCustomer(c *gin.Context, dc desmond.Context) {
	c.String(200, "Success")

	var req CustomerRedactRequest
	if err := c.BindJSON(&req); err != nil {
		logrus.WithFields(logrus.Fields{
			"error":       err.Error(),
			"queryString": c.Request.URL.RawQuery,
		}).Error("(RedactCustomer) binding request to model produced error")
	}

	// Logging as an error so we can track in Sentry
	logrus.WithFields(logrus.Fields{
		"shopId":         req.ShopID,
		"shopDomain":     req.ShopDomain,
		"customer":       req.Customer,
		"ordersToRedact": req.OrdersToRedact,
	}).Error("(RedactCustomer) received customer redact request")

	shopName := strings.Split(req.ShopDomain, ".")[0]
	go redactOrders(shopName, req.OrdersToRedact, h)
	return
}

func (h *Handlers) RedactShop(c *gin.Context, dc desmond.Context) {
	c.String(200, "Success")

	var req ShopRedactRequest
	if err := c.BindJSON(&req); err != nil {
		logrus.WithFields(logrus.Fields{
			"error":       err.Error(),
			"queryString": c.Request.URL.RawQuery,
		}).Error("(RedactShop) binding request to model produced error")
	}

	// Logging as an error so we can track in Sentry
	logrus.WithFields(logrus.Fields{
		"shopId":     req.ShopID,
		"shopDomain": req.ShopDomain,
	}).Error("(RedactShop) received shop redact request")

	shopName := strings.Split(req.ShopDomain, ".")[0]
	go redactOrdersByShopName(shopName, h)
	return
}

func (h *Handlers) CustomerDataRequest(c *gin.Context, dc desmond.Context) {
	c.String(200, "Success")

	var req CustomerDataRequest
	if err := c.BindJSON(&req); err != nil {
		logrus.WithFields(logrus.Fields{
			"error":       err.Error(),
			"queryString": c.Request.URL.RawQuery,
		}).Error("(CustomerDataRequest) binding request to model produced error")
	}

	// Logging as an error so we can track in Sentry
	logrus.WithFields(logrus.Fields{
		"shopId":          req.ShopID,
		"shopDomain":      req.ShopDomain,
		"customer":        req.Customer,
		"ordersRequested": req.OrdersRequested,
	}).Error("(CustomerDataRequest) received customer data request")
	return
}

func redactOrders(shopName string, orders []int, h *Handlers) {
	_, err := findShopByName(shopName, h)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":    err.Error(),
			"shopName": shopName,
		}).Error("(RedactOrders) shop not found")
		return
	}

	logMessage := "(RedactOrders) analytics orders successfully redacted"
	for _, orderID := range orders {
		err := redactOrderByOrderID(orderID, h)
		if err != nil {
			logMessage = "(RedactOrders) redacting orders producced an error"
			logrus.WithFields(logrus.Fields{
				"shopName": shopName,
				"orderID":  orderID,
				"error":    err.Error(),
			}).Error("(RedactOrders) redacting order produced an error")
		}
	}

	logrus.WithFields(logrus.Fields{
		"shopName":       shopName,
		"ordersRedacted": orders,
	}).Info(logMessage)
	return
}

func redactOrdersByShopName(shopName string, h *Handlers) {
	// Throw error if Shop doesn't exist
	_, err := findShopByName(shopName, h)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":    err.Error(),
			"shopName": shopName,
		}).Error("(RedactOrdersByShop) shop not found")
		return
	}

	result := h.DB.MustExec(
		`UPDATE shopify_analytics_orders
		SET
			customer_email = null,
			total_price = null,
			financial_status = null,
			fulfillment_status = null,
			gateway = null,
			redacted = true
		WHERE shop_name = $1;`, shopName)

	// log successful order redaction
	logrus.WithFields(logrus.Fields{
		"sqlResult": result,
		"shopName":  shopName,
	}).Info("(RedactOrdersByShop) analytics orders successfully redacted")
	return
}
