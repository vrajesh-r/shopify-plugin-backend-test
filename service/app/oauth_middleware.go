package app

import (
	"github.com/getbread/breadkit/desmond"
	"github.com/getbread/shopify_plugin_backend/service/shopify"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handlers) VerifyAppOAuthPermissionsUpToDate(c *gin.Context, dc desmond.Context) {
	_, shop, err := h.validateAppSession(c)
	if err != nil {
		logrus.Warn("Failed to verify OAuth permissions")
		// Let the next handler manage session verification
		c.Next()
		return
	}

	if !shop.OAuthPermissionsUpToDate {
		clearNoncesByShopId(shop.Id, h)
		nonce, err := createNonceByShopId(shop.Id, h)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error":    err.Error(),
				"shopName": shop.Shop,
			}).Error("(verifyAppOAuthPermissionsUpToDate) creating nonce produced error")
			c.String(400, err.Error())
			return
		}

		// Render page which redirects to install URL
		c.HTML(200, "iframe_authorize.html", gin.H{
			"url": shopify.InstallUrl(shop.Shop, nonce.Nonce, OAUTH_REDIRECT_PATH),
		})

		c.Abort()
		return
	}

	c.Next()
}
