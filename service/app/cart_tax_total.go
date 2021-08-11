package app

import (
	"strconv"
	"strings"

	"github.com/getbread/breadkit/desmond"
	"github.com/getbread/shopify_plugin_backend/service/shopify"
	"github.com/getbread/shopify_plugin_backend/service/tax"
	"github.com/getbread/shopify_plugin_backend/service/types"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type cartTaxTotalRequest struct {
	ShopName  string                     `json:"shopName"`
	State     string                     `json:"state"`
	Zip       string                     `json:"zip"`
	LineItems []shopify.AddToCartRequest `json:"lineItems"`
}

type cartTaxTotalResponse struct {
	TotalTax string `json:"totalTax"`
}

func (h *Handlers) CartTaxTotal(c *gin.Context, dc desmond.Context) {
	// parse request body
	var req cartTaxTotalRequest
	if err := c.BindJSON(&req); err != nil {
		log.WithFields(log.Fields{
			"error":       err.Error(),
			"queryString": c.Request.URL.RawQuery,
		}).Error("(CartTaxTotal) binding request to model produced error")
		c.String(400, err.Error())
		h.BackupCartTaxTotal(c, req)
		return
	}

	// create new server side session with Shopify store
	ss, err := shopify.NewSession(req.ShopName)
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err.Error(),
			"request": req,
		}).Error("(CartTaxTotal) creating shopify session produced error")
		c.String(400, err.Error())
		h.BackupCartTaxTotal(c, req)
		return
	}

	// recreate cart in new Shopify Session
	// items must be added serially
	for _, li := range req.LineItems {
		var res shopify.AddToCartResponse
		if err := ss.AddToCart(&li, &res); err != nil {
			log.WithFields(log.Fields{
				"error":    err.Error(),
				"request":  req,
				"lineItem": li,
			}).Error("(CartTaxTotal) adding line item to cart produced error")
			h.BackupCartTaxTotal(c, req)
			return
		}
	}

	// checkout & get tax from Shopify as first plan
	if err = ss.CreateCheckout(); err != nil {
		log.WithFields(log.Fields{
			"error":   err.Error(),
			"request": req,
		}).Error("(CartTaxTotal) creating checkout produced error")
		h.BackupCartTaxTotal(c, req)
		return
	}
	taxCheckRequest := shopify.CartTaxCheckRequest{
		AuthenticityToken: ss.AuthenticityToken,
		Country:           "United States",
		Province:          req.State,
		Zip:               req.Zip,
	}
	totalTax, err := ss.CartTaxCheck(&taxCheckRequest)
	if err == nil {
		c.JSON(200, &cartTaxTotalResponse{
			TotalTax: totalTax,
		})
		return
	}
	// bring error to attention
	log.WithFields(log.Fields{
		"error":   err.Error(),
		"request": req,
	}).Error("(CartTaxTotal) submitting shipping info produced error, attemping back tax solution")

	// Backup Tax Strategy
	// This will be used as life line when
	// scraping Shopify checkout process fails

	h.BackupCartTaxTotal(c, req)
}

func (h *Handlers) BackupCartTaxTotal(c *gin.Context, req cartTaxTotalRequest) {
	var totalTax string

	// pull shop
	shopUrl := c.Query("shop")
	shopName := strings.Split(shopUrl, ".")[0]
	shop, err := findShopByName(shopName, h)
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err.Error(),
			"request": req,
		}).Error("(CartTaxTotal) query for shop produced error")
		c.String(400, err.Error())
		return
	}

	// shopify client
	sc := shopify.NewClient(shop.Shop, shop.AccessToken)

	// spin off queries for shop locations
	done := make(chan *[]string, 2)
	go func(done chan *[]string) {
		var res shopify.SearchLocationsResponse
		var states []string
		err := sc.QueryLocations(&res)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
				"shop":  shop,
			}).Error("(CartTaxTotal) query for shop locations produced error")
			done <- &states
			return
		}
		for _, location := range res.Locations {
			states = append(states, location.Province)
		}
		done <- &states
	}(done)
	go func(done chan *[]string) {
		var res shopify.SearchShopResponse
		var states []string
		err := sc.QueryShop(&res)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
				"shop":  shop,
			}).Error("(CartTaxTotal) query for Shopify shop in backup tax produced error")
			done <- &states
			return
		}
		states = append(states, res.Shop.Province)
		done <- &states
	}(done)

	// collect results
	var states []string
	for x := 0; x < cap(done); x++ {
		s := <-done
		states = append(states, *s...)
	}

	// attempt to match shop locations to state provided in request
	var stringMatches = func(tests []string, control string) bool {
		for _, test := range tests {
			if strings.ToLower(test) == strings.ToLower(control) {
				return true
			}
		}
		return false
	}
	// store does not have presence where order is being shipped
	// order has no tax liabilities, return total tax of 0 cents
	var test string
	if len(req.State) == 2 {
		fullState, err := tax.StateAbbreviationToFullName(req.State)
		if err != nil {
			log.WithFields(log.Fields{
				"error":   err.Error(),
				"request": req,
			}).Error("(CartTaxTotal) state abbreviation conversion to full name produced error")
			fullState = ""
		}
		test = fullState
	} else {
		test = req.State
	}
	if !stringMatches(states, test) {
		c.JSON(200, &cartTaxTotalResponse{
			TotalTax: "0",
		})
		return
	}

	// pull rates for zip
	var ratesRes tax.TaxRateResponse
	zipString, err := strconv.Atoi(req.Zip)
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err.Error(),
			"request": req,
		}).Error("(CartTaxTotal) converting string zip to int zip produced error")
		c.String(400, err.Error())
		return
	}
	if err := tax.TaxRateByZip(zipString, &ratesRes); err != nil {
		log.WithFields(log.Fields{
			"error":   err.Error(),
			"request": req,
		}).Error("(CartTaxTotal) query for tax rates from Avalara produced error")
		c.JSON(200, &cartTaxTotalResponse{
			TotalTax: "0",
		})
		return
	}

	// get the sub total for the cart
	// query all product variants
	var getProductVariantPriceInMillicents = func(variantId string, results chan types.Millicents) {
		var res shopify.SearchProductVariantByIdResponse
		if err := sc.QueryProductVariant(variantId, &res); err != nil {
			log.WithFields(log.Fields{
				"error":   err.Error(),
				"request": req,
			}).Error("(CartTaxTotal) query for product variant in back-up tax calcs produced error")
			results <- types.Millicents(0)
			return
		}
		m, err := types.USDToMillicents(res.Variant.Price)
		if err != nil {
			log.WithFields(log.Fields{
				"error":   err.Error(),
				"request": req,
				"variant": res.Variant,
			}).Error("(CartTaxTotal) converting variant price to Millicents in back-up tax calcs produced error")
			results <- types.Millicents(0)
			return
		}
		results <- m
	}
	results := make(chan types.Millicents, len(req.LineItems))
	for _, li := range req.LineItems {
		go getProductVariantPriceInMillicents(strconv.Itoa(li.Id), results)
	}

	// collect results
	var subtotalMillicents types.Millicents
	for x := 0; x < cap(results); x++ {
		subtotalMillicents += <-results
	}

	// calculate tax on cart subtotal
	taxRatePercent := ratesRes.TotalRate / 100.00
	totalTaxF := float64(subtotalMillicents) * taxRatePercent           // remove this float multiplication
	totalTax = strconv.Itoa(int(types.Millicents(totalTaxF).ToCents())) // we could lose fidelity here as well

	// respond
	c.JSON(200, &cartTaxTotalResponse{
		TotalTax: totalTax,
	})
}
