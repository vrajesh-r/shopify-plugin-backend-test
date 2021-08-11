package app

import (
	"github.com/getbread/breadkit/desmond"
	"github.com/getbread/shopify_plugin_backend/service/shopify"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type cartShippingOptionsRequest struct {
	ShopName  string                     `json:"shopName"`
	State     string                     `json:"state"`
	Zip       string                     `json:"zip"`
	LineItems []shopify.AddToCartRequest `json:"lineItems"`
}

type cartShippingOptionsResponse []shopify.ShippingLine

func (h *Handlers) CartShippingOptions(c *gin.Context, dc desmond.Context) {
	// parse request
	var req cartTaxTotalRequest
	if err := c.BindJSON(&req); err != nil {
		log.WithFields(log.Fields{
			"error":       err.Error(),
			"queryString": c.Request.URL.RawQuery,
		}).Error("(CartShippingOptions) binding request to model produced error")
		c.String(400, err.Error())
		return
	}

	// create new server side session with Shopify store
	ss, err := shopify.NewSession(req.ShopName)
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err.Error(),
			"request": req,
		}).Error("(CartShippingOptions) starting Shopify session produced error")
		c.String(400, err.Error())
		return
	}

	// add items to cart
	// must be done serially
	for _, li := range req.LineItems {
		var res shopify.AddToCartResponse
		if err := ss.AddToCart(&li, &res); err != nil {
			log.WithFields(log.Fields{
				"error":    err.Error(),
				"request":  req,
				"lineItem": li,
			}).Error("(CartShippingOptions) adding line item to cart produced error")
			c.String(400, err.Error())
			return
		}
	}

	// request shipping rates
	var res shopify.ShippingRatesResponse
	if err = ss.GetCartShippingRates(req.Zip, "United States", req.State, &res); err != nil {
		log.WithFields(log.Fields{
			"error":   err.Error(),
			"request": req,
		}).Error("(CartShippingOptions) query for shipping rates produced error")
		c.String(400, err.Error())
		return
	}

	c.JSON(200, res.ShippingRates)
}
