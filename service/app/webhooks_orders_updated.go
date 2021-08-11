package app

import (
	"database/sql"
	"strings"

	"github.com/getbread/breadkit/desmond"
	"github.com/getbread/breadkit/zeus/searcher"
	zeus "github.com/getbread/breadkit/zeus/types"
	"github.com/getbread/shopify_plugin_backend/service/search"
	"github.com/getbread/shopify_plugin_backend/service/shopify"
	"github.com/getbread/shopify_plugin_backend/service/update"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type OrderUpdatedRequest struct {
	ID                int              `json:"id"`
	CheckoutID        int              `json:"checkout_id"`
	Gateway           string           `json:"gateway"`
	OrderID           int              `json:"order_id"`
	OrderNumber       int              `json:"order_number"`
	Customer          shopify.Customer `json:"customer"`
	TotalPrice        string           `json:"total_price"`
	FinancialStatus   string           `json:"financial_status"`
	FulfillmentStatus string           `json:"fulfillment_status"`
	Test              bool             `json:"test"`
	BillingAddress    shopify.Address  `json:"billing_address"`
	ShippingAddress   shopify.Address  `json:"shipping_address"`
}

func (h *Handlers) OrderUpdated(c *gin.Context, dc desmond.Context) {
	c.String(200, "order updated")

	var req OrderUpdatedRequest
	if err := c.BindJSON(&req); err != nil {
		logrus.WithFields(logrus.Fields{
			"error":       err.Error(),
			"queryString": c.Request.URL.RawQuery,
		}).Error("(WebhookOrderUpdated) binding request to model produced error")
		return
	}

	shopDomain := c.Request.Header.Get("X-Shopify-Shop-Domain")
	shopName := strings.Split(shopDomain, ".")[0]

	if orderPlacedInUSorCA(req.BillingAddress.CountryCode, req.ShippingAddress.CountryCode, req.Customer.DefaultAddress.CountryCode) {
		go updateAnalyticsOrder(req, shopName, h)
	}
	return
}

func updateAnalyticsOrder(r OrderUpdatedRequest, shopName string, h *Handlers) {
	// Find Analytics Order by order_id
	searchRequest := search.AnalyticsOrderSearchRequest{}
	searchRequest.AddFilter(search.AnalyticsOrderSearch_OrderID, r.ID, searcher.Operator_EQ, searcher.Condition_AND)
	searchRequest.AddFilter(search.AnalyticsOrderSearch_ShopName, shopName, searcher.Operator_EQ, searcher.Condition_AND)
	searchRequest.Limit = 1
	analyticsOrders, err := h.AnalyticsOrderSearcher.Search(searchRequest)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":         err.Error(),
			"orderID":       r.ID,
			"shop":          shopName,
			"searchRequest": searchRequest,
		}).Error("(WebhookOrderUpdated) searching for Analytics Order produced an error")
		return
	}

	if len(analyticsOrders) == 0 {
		logrus.WithFields(logrus.Fields{
			"orderID": r.ID,
			"shop":    shopName,
		}).Info("(WebhookOrderUpdated) Analytics Order not found")
		return
	}

	order := analyticsOrders[0]

	savedfinancialStatus, _ := order.FinancialStatus.Value()
	savedfulfillmentStatus, _ := order.FulfillmentStatus.Value()

	// Exit early if neither status has changed
	if savedfinancialStatus == r.FinancialStatus && savedfulfillmentStatus == r.FulfillmentStatus {
		return
	}

	// Update Analytics Order
	updateRequest := update.AnalyticsOrderUpdateRequest{
		Id:      zeus.Uuid(order.ID),
		Updates: map[update.AnalyticsOrderUpdateField]interface{}{},
	}
	updateRequest.Updates[update.AnalyticsOrderUpdate_FinancialStatus] = zeus.NullString{NullString: sql.NullString{String: r.FinancialStatus, Valid: true}}
	updateRequest.Updates[update.AnalyticsOrderUpdate_FulfillmentStatus] = zeus.NullString{NullString: sql.NullString{String: r.FulfillmentStatus, Valid: true}}
	err = h.AnalyticsOrderUpdater.Update(updateRequest)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":         err.Error(),
			"orderID":       r.ID,
			"shop":          shopName,
			"updateRequest": updateRequest,
		}).Error("(WebhookOrderUpdated) updating Analytics Order produced an error")
		return
	}
	return
}
