package app

import (
	"strconv"
	"strings"
	"text/template"

	"github.com/getbread/breadkit/desmond"
	"github.com/getbread/shopify_plugin_backend/service/shopify"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type CheckoutTemplateData struct {
	OrderNumber         int
	CreatedAt           string
	Email               string
	BillingAddress      shopify.Address
	ShippingAddress     shopify.Address
	LineItems           []shopify.LineItem
	TotalLineItemsPrice string
	TotalTax            string
	TotalShipping       string
	TotalPrice          string
}

func (h *Handlers) OrderConfirmation(c *gin.Context, dc desmond.Context) {
	// ref
	responseError := "An unexpected error occurred. \n 1) Ensure you are viewing the correct page. \n 2) Refresh."

	// set up
	oid := c.Param("order_id")
	shopUrl := c.Query("shop")
	shopName := strings.Split(shopUrl, ".")[0]
	h.createOrderConfirmation(oid, shopName, responseError, c)
}

func (h *Handlers) CartOrderConfirmation(c *gin.Context, dc desmond.Context) {
	// ref
	responseError := "An unexpected error occurred. \n 1) Ensure you are viewing the correct page. \n 2) Refresh."

	transactionID := c.Query("transactionId")
	shopUrl := c.Query("shop")
	shopName := strings.Split(shopUrl, ".")[0]
	order, err := findOrderByTransactionId(transactionID, h)
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err.Error(),
			"shopUrl": shopUrl,
		}).Error("(CartOrderConfirmation) query for shop produced error")
		c.String(400, responseError)
		return
	}

	h.createOrderConfirmation(strconv.Itoa(order.OrderId), shopName, responseError, c)
}

func (h *Handlers) createOrderConfirmation(oid string, shopName string, responseError string, c *gin.Context) {

	// pull shop
	shop, err := findShopByName(shopName, h)
	if err != nil {
		log.WithFields(log.Fields{
			"error":       err.Error(),
			"queryString": c.Request.URL.RawQuery,
			"orderID":     oid,
			"shop":        shopName,
		}).Error("(OrderConfirmation) query for shop produced error")
		c.String(400, responseError)
		return
	}

	// query order
	sc := shopify.NewClient(shop.Shop, shop.AccessToken)
	var res shopify.SearchOrderResponse
	if err = sc.QueryOrder(oid, &res); err != nil {
		log.WithFields(log.Fields{
			"error":       err.Error(),
			"queryString": c.Request.URL.RawQuery,
			"orderID":     oid,
			"shop":        shop.Shop,
			"shopID":      shop.Id,
		}).Error("(OrderConfirmation) query for Shopify order produced error")
		c.String(400, responseError)
		return
	}

	// create custom template with special delimeters
	c.Header("Content-Type", "application/liquid")
	t := template.New("order_confirmation")
	t.Delims("<|", "|>")
	t, err = t.ParseFiles("build/liquid/checkout_confirmation.html")
	if err != nil {
		log.WithFields(log.Fields{
			"error":       err.Error(),
			"queryString": c.Request.URL.RawQuery,
			"orderID":     res.Order.ID,
			"shop":        shop.Shop,
			"shopID":      shop.Id,
		}).Error("(OrderConfirmation) parsing checkout_confirmation.liquid produced error")
		c.String(400, responseError)
		return
	}
	td := &CheckoutTemplateData{
		OrderNumber:         res.Order.OrderNumber,
		CreatedAt:           res.Order.CreatedAt,
		Email:               res.Order.Email,
		BillingAddress:      res.Order.BillingAddress,
		ShippingAddress:     res.Order.ShippingAddress,
		LineItems:           res.Order.LineItems,
		TotalLineItemsPrice: res.Order.TotalLineItemsPrice,
		TotalTax:            res.Order.TotalTax,
		TotalShipping:       aggregateShippingPrices(res.Order.ShippingLines),
		TotalPrice:          res.Order.TotalPrice,
	}
	err = t.ExecuteTemplate(c.Writer, "checkout_confirmation.html", td)
	if err != nil {
		log.WithFields(log.Fields{
			"error":       err.Error(),
			"queryString": c.Request.URL.RawQuery,
			"orderID":     res.Order.ID,
			"shop":        shop.Shop,
			"shopID":      shop.Id,
		}).Error("(OrderConfirmation) executing checkout_confirmation.liquid template produced error")
		c.String(400, responseError)
		return
	}
}

func aggregateShippingPrices(shippingLines []shopify.ShippingLine) string {
	var total float64
	for _, s := range shippingLines {
		sp, err := strconv.ParseFloat(s.Price, 64)
		if err != nil {
			log.WithFields(log.Fields{
				"error":        err.Error(),
				"shippingLine": s,
			}).Error("(OrderConfirmation) aggregating shipping prices produced error")
			return "See email from store"
		}
		total += sp
	}
	totalShipping := strconv.FormatFloat(total, 'f', 2, 64)
	return totalShipping
}
