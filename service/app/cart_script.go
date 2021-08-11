package app

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/getbread/breadkit/desmond"
	zeus "github.com/getbread/breadkit/zeus/types"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func (h *Handlers) CartJS(c *gin.Context, dc desmond.Context) {
	// query shop
	shopId := zeus.Uuid(c.Param("shopId"))
	shop, err := findShopById(shopId, h)
	if err != nil {
		log.WithFields(log.Fields{
			"shopId": string(shopId),
			"error":  err.Error(),
		}).Error("(CartJS) query for shop produced error")
		c.String(400, err.Error())
		return
	}

	var buf bytes.Buffer
	var cartJs string
	var templateData gin.H

	activeVersion := shop.ActiveVersion
	if activeVersion == "" {
		activeVersion = BreadClassic
	}

	if activeVersion == BreadClassic {
		apiKey, _ := shop.GetAPIKeys()

		// Check if merchant is grandfathered into accelerated checkout
		if !shop.AcceleratedCheckoutPermitted {
			shop.AllowCheckoutPDP = false
			shop.AllowCheckoutCart = false
		}
		cartJs = "cart.js"
		templateData = gin.H{
			"CSS":                        shop.CSS,
			"CSSCart":                    shop.CSSCart,
			"ActsAsLabel":                shop.ActsAsLabel,
			"AsLowAs":                    shop.AsLowAs,
			"ApiKey":                     apiKey,
			"AllowCheckoutPDP":           shop.AllowCheckoutPDP,
			"EnableAddToCart":            shop.EnableAddToCart,
			"AllowCheckoutCart":          shop.AllowCheckoutCart,
			"HealthcareMode":             shop.HealthcareMode,
			"TargetedFinancing":          shop.TargetedFinancing,
			"TargetedFinancingID":        shop.TargetedFinancingID,
			"TargetedFinancingThreshold": shop.TargetedFinancingThreshold,
			"CalculateTaxDraftOrder":     shop.DraftOrderTax,
			"BreadJS":                    shop.CheckoutHost() + "/bread.js",
			"ActiveVersion":              activeVersion,
			"Env":                        appConfig.Environment,
		}
	} else { // shop.ActiveVersion is platform
		cartJs = "cart_platform.js"
		templateData = gin.H{
			"BreadJS":        shop.PlatformCheckoutHost() + "/sdk.js",
			"ActiveVersion":  activeVersion,
			"IntegrationKey": shop.GetIntegrationKey(),
			"Env":            appConfig.Environment,
		}
	}

	// parse template
	tmp, err := template.ParseFiles(fmt.Sprintf("service/cmd/shopify_plugin_backend/build/templates/%s", cartJs))
	if err != nil {
		log.WithFields(log.Fields{
			"shop":  shop,
			"error": err.Error(),
		}).Error(fmt.Sprintf("(CartJS) parsing %s produced error", cartJs))
		c.String(400, err.Error())
		return
	}

	if err = tmp.ExecuteTemplate(&buf, cartJs, templateData); err != nil {
		log.WithFields(log.Fields{
			"shop":  shop,
			"error": err.Error(),
		}).Error(fmt.Sprintf("(CartJS) executing %s produced error", cartJs))
		c.String(400, err.Error())
		return
	}
	serveInMemoryContent(c.Writer, c.Request, cartJs, buf.Bytes())
}

func serveInMemoryContent(w http.ResponseWriter, r *http.Request, name string, content []byte) {
	h := sha1.New()
	h.Write(content)
	eTag := fmt.Sprintf(`W/"%x"`, h.Sum(nil))
	w.Header().Set("ETag", eTag)
	http.ServeContent(w, r, name, time.Now(), bytes.NewReader(content))
}
