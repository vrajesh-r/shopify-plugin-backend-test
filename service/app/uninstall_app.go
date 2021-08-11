package app

import (
	"strings"

	"github.com/getbread/breadkit/desmond"
	"github.com/getbread/shopify_plugin_backend/service/shopify"
	"github.com/getbread/shopify_plugin_backend/service/update"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func (h *Handlers) UninstallApp(c *gin.Context, dc desmond.Context) {
	c.String(200, "complete")

	// parse body
	var req shopify.Shop
	if err := c.BindJSON(&req); err != nil {
		panic(err)
	}

	// query shop
	name := strings.Split(req.Domain, ".")[0]
	shop, err := findShopByName(name, h)
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err.Error(),
			"request": req,
		}).Error("(UninstallApp) query for shop produced error")
		return
	}

	// update shop
	sur := update.ShopUpdateRequest{
		Id:      shop.Id,
		Updates: map[update.ShopUpdateField]interface{}{},
	}
	sur.Updates[update.ShopUpdate_BreadApiKey] = ""
	sur.Updates[update.ShopUpdate_BreadSecretKey] = ""
	sur.Updates[update.ShopUpdate_IntegrationKey] = ""
	sur.Updates[update.ShopUpdate_Production] = false
	sur.Updates[update.ShopUpdate_PlatformProduction] = false
	if err = h.ShopUpdater.Update(sur); err != nil {
		log.WithFields(log.Fields{
			"error":   err.Error(),
			"request": req,
			"shop":    shop,
		}).Error("(UninstallApp) updating shop produced an error")
		return
	}

	// email & log
	log.WithFields(log.Fields{
		"request": req,
		"shop":    shop,
	}).Info("(UninstalledApp) merchant uninstalled app")
}
