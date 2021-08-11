package app

import (
	"strconv"
	"strings"
	"time"

	"github.com/getbread/breadkit/desmond"
	"github.com/getbread/shopify_plugin_backend/service/bread"
	"github.com/getbread/shopify_plugin_backend/service/shopify"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type cartTaxTotalDraftOrderRequest struct {
	ShopName        string                     `json:"shopName"`
	LineItems       []shopify.LineItem         `json:"lineItems"`
	ShippingLine    shopify.CustomShippingLine `json:"shippingLine"`
	ShippingAddress bread.Contact              `json:"shippingAddress"`
}

type cartTaxTotalDraftOrderResponse struct {
	TotalTax int `json:"totalTax"`
}

func (h *Handlers) CartTaxTotalDraftOrder(c *gin.Context, dc desmond.Context) {
	// Parse body into request struct
	var req cartTaxTotalDraftOrderRequest
	if err := c.BindJSON(&req); err != nil {
		logrus.WithFields(logrus.Fields{
			"error":       err.Error(),
			"queryString": c.Request.URL.RawQuery,
		}).Error("(CartTaxTotalDraftOrder) binding request to model produced error")
		c.String(400, err.Error())
		return
	}

	// Find shop by name
	shop, err := findShopByName(req.ShopName, h)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":   err.Error(),
			"request": req,
		}).Error("(CartTaxTotalDraftOrder) querying shop by name produced error")
		c.String(400, err.Error())
		return
	}

	// Create temporary draft order
	sc := shopify.NewClient(shop.Shop, shop.AccessToken)
	var res shopify.GetDraftOrderResponse

	cdor := &shopify.CreateDraftOrderRequest{
		DraftOrder: shopify.DraftOrderRequest{
			LineItems:       req.LineItems,
			ShippingAddress: *convertContactToShopifyAddress(&req.ShippingAddress),
			ShippingLine:    req.ShippingLine,
		},
	}

	statusCode, err := sc.CreateDraftOrderRequest(cdor, &res)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
			"data":  cdor,
		}).Error("(CartTaxTotalDraftOrder) creating draft order produced error")
		c.String(400, err.Error())
		return
	}

	draftOrderID := strconv.Itoa(res.DraftOrder.ID)

	// 202 => accepted but still processing - poll until we get 201 or 200 limit 3 attempts
	attempts := 0
	for statusCode == 202 || attempts > 3 {
		time.Sleep(time.Second * 1)
		attempts++
		statusCode, err = sc.GetDraftOrderRequest(draftOrderID, &res)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error":        err.Error(),
				"draftOrderID": draftOrderID,
				"request":      cdor,
			}).Error("(CartTaxTotalDraftOrder) querying draft order produced error")
			go deleteTempDraftOrder(draftOrderID, sc)
			c.String(400, err.Error())
			return
		}
	}

	if statusCode == 202 {
		logrus.WithFields(logrus.Fields{
			"draftOrder": res.DraftOrder,
			"request":    cdor,
		}).Error("(CartTaxTotalDraftOrder) receiving 202 response code after 3 draft order query attempts")
		go deleteTempDraftOrder(draftOrderID, sc)
		c.String(503, "Shopify API timeout")
		return
	}

	totalTaxString := strings.Replace(res.DraftOrder.TotalTax, ".", "", -1)
	totalTax, err := strconv.Atoi(totalTaxString)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":    err.Error(),
			"totaltax": res.DraftOrder.TotalTax,
		}).Error("(CartTaxTotalDraftOrder) converting tax produced error")
		go deleteTempDraftOrder(draftOrderID, sc)
		c.String(400, err.Error())
		return
	}

	go deleteTempDraftOrder(draftOrderID, sc)

	// Return tax total
	c.JSON(200, &cartTaxTotalDraftOrderResponse{
		TotalTax: totalTax,
	})
	return
}

func deleteTempDraftOrder(orderId string, client *shopify.Client) {
	var res shopify.DeleteDraftOrderResponse
	err := client.DeleteDraftOrder(orderId, &res)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":   err.Error(),
			"orderid": orderId,
		}).Error("(CartTaxTotalDraftOrder) deleting draft order produced error")
	}
}
