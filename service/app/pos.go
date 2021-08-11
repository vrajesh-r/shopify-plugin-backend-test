package app

import (
	"strings"

	"github.com/getbread/breadkit/desmond"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handlers) RenderPOS(c *gin.Context, dc desmond.Context) {
	shopUrl := c.Query("shop")
	shopName := strings.Split(shopUrl, ".")[0]

	shop, err := findShopByName(shopName, h)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":       err.Error(),
			"queryString": c.Request.URL.RawQuery,
		}).Error("(AuthorizePOS) query for shop produced error")
		c.AbortWithStatus(401)
		return
	}

	if !shop.POSAccess {
		logrus.WithFields(logrus.Fields{
			"shop": shop.Shop,
		}).Error("(AuthorizePOS) unauthorized shop accessing Bread POS")
		c.HTML(401, "pos-restrict.html", gin.H{
			"ShopifyAPIKey": appConfig.ShopifyConfig.ShopifyApiKey.Unmask(),
			"ShopName":      shop.Shop,
		})
		return
	}

	breadJS := shop.CheckoutHost() + "/bread.js"
	apiKey, _ := shop.GetAPIKeys()

	c.HTML(200, "pos.html", gin.H{
		"ShopifyAPIKey":              appConfig.ShopifyConfig.ShopifyApiKey.Unmask(),
		"ShopName":                   shop.Shop,
		"BreadJS":                    breadJS,
		"BreadAPIKey":                apiKey,
		"Production":                 shop.Production,
		"TargetedFinancing":          shop.TargetedFinancing,
		"TargetedFinancingID":        shop.TargetedFinancingID,
		"TargetedFinancingThreshold": shop.TargetedFinancingThreshold,
	})
	return
}
