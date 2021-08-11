package app

import (
	"strings"

	"github.com/getbread/breadkit/desmond"
	"github.com/getbread/shopify_plugin_backend/service/shopify"
	"github.com/getbread/shopify_plugin_backend/service/types"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func (h *Handlers) InstallApp(c *gin.Context, dc desmond.Context) {
	// pull, verify query params
	shopName := strings.ToLower(c.Query("shop"))
	if len(shopName) == 0 {
		log.WithField("queryString", c.Request.URL.RawQuery).Error("(InstallApp) shop name not in request")
		c.String(400, "invalid request")
		return
	}

	// ref
	var shop types.Shop

	// find existing shop
	shop, err := findShopByName(shopName, h)
	if err != nil {
		shop, err = createShopByName(shopName, h)
		if err != nil {
			log.WithFields(log.Fields{
				"error":    err.Error(),
				"shopName": shopName,
			}).Error("(InstallApp) creating shop produced error")
			c.String(400, err.Error())
			return
		}
	}

	// clear all nonces for shop
	clearNoncesByShopId(shop.Id, h)

	// create nonce
	nonce, err := createNonceByShopId(shop.Id, h)
	if err != nil {
		log.WithFields(log.Fields{
			"error":    err.Error(),
			"shopName": shopName,
		}).Error("(InstallApp) creating nonce produced error")
		c.String(400, err.Error())
		return
	}

	// redirect
	c.Redirect(302, shopify.InstallUrl(shop.Shop, nonce.Nonce, OAUTH_REDIRECT_PATH))
}
