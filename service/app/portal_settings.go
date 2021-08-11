package app

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/getbread/breadkit/desmond"
	"github.com/getbread/shopify_plugin_backend/service/shopify"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handlers) AuthorizePortal(c *gin.Context, dc desmond.Context) {
	// ref
	responseError := "An error occurred, please try again"

	// pull params from query
	shopUrl := c.Query("shop")
	shopName := strings.Split(shopUrl, ".")[0]

	// pull shop
	shop, err := findShopByName(shopName, h)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":       err.Error(),
			"queryString": c.Request.URL.RawQuery,
		}).Error("(RenderPortal) query for shop produced error")
		c.String(400, responseError)
		return
	}

	// run shop through install process
	if shop.AccessToken == "" {
		shop.Shop = shopName
		// clear existing shop nonces?
		nonce, err := createNonceByShopId(shop.Id, h) // will shop always be defined
		if err != nil {                               // handle this better
			logrus.WithFields(logrus.Fields{
				"error":       err.Error(),
				"queryString": c.Request.URL.RawQuery,
				"shop":        shop,
			}).Error("(RenderPortal) creating nonce produced error")
			c.String(400, responseError)
			return
		}
		c.HTML(200, "iframe_authorize.html", gin.H{
			"url": shopify.InstallUrl(shop.Shop, nonce.Nonce, OAUTH_REDIRECT_PATH),
		})
		return
	}

	// generate new session
	session, err := createSessionByShopId(shop.Id, h)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":       err.Error(),
			"queryString": c.Request.URL.RawQuery,
			"shop":        shop,
		}).Error("(RenderPortal) creating session produced error")
		c.String(400, responseError)
		return
	}

	// add cookie
	expire := time.Unix(session.Expiration, 0)
	cookie := &http.Cookie{
		Name:       ADMIN_COOKIE_NAME,
		Value:      string(session.Id),
		Expires:    expire,
		RawExpires: expire.Format(time.UnixDate),
		Secure:     true,
		Path:       "/",
		HttpOnly:   true,
		SameSite:   http.SameSiteNoneMode,
	}
	http.SetCookie(c.Writer, cookie)

	c.Redirect(302, fmt.Sprintf("%s/portal/settings", appConfig.HostConfig.MiltonHost))
}

func (h *Handlers) PortalSettings(c *gin.Context, dc desmond.Context) {
	session, shop, err := h.validateAppSession(c)
	if err != nil {
		if err.Error() == "empty session id" {
			c.HTML(400, "app_error.html", gin.H{
				"messagePrimary":   "Session expired",
				"messageSecondary": "Please restart your session by selecting Apps and then Bread",
			})
			return
		}
		logrus.WithError(err).WithFields(logrus.Fields{
			"session": fmt.Sprintf("%+v", session),
		}).Error("(PortalSettings) validating session failed")
		c.String(400, err.Error())
		return
	}

	webhooksUpdated, err := updateWebhooks(shop, false)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"error": err,
			"shop":  shop,
		}).Info("(PortalSettings) checking webhooks produced error")
	}

	// Render portal template
	c.HTML(200, "settings.html", gin.H{
		"apiKey":                       appConfig.ShopifyConfig.ShopifyApiKey.Unmask(),
		"host":                         appConfig.HostConfig.MiltonHost,
		"shopId":                       shop.Id,
		"shopName":                     shop.Shop,
		"webhooksUpdated":              webhooksUpdated,
		"breadApiKey":                  shop.BreadApiKey,
		"breadSecretKey":               shop.BreadSecretKey,
		"breadSandboxApiKey":           shop.BreadSandboxApiKey,
		"breadSandboxSecretKey":        shop.BreadSandboxSecretKey,
		"breadCustomCss":               shop.CSS,
		"breadCustomCssCart":           shop.CSSCart,
		"breadAutoAuthorize":           shop.AutoAuthorize,
		"breadAutoSettle":              shop.AutoSettle,
		"breadActsAsLabel":             shop.ActsAsLabel,
		"breadCreateCustomers":         shop.CreateCustomers,
		"breadTestMode":                !shop.Production,
		"breadManualEmbedScript":       shop.ManualEmbedScript,
		"breadAsLowAs":                 shop.AsLowAs,
		"breadEnableOrderWebhooks":     shop.EnableOrderWebhooks,
		"breadAllowCheckoutPDP":        shop.AllowCheckoutPDP,
		"breadEnableAddToCart":         shop.EnableAddToCart,
		"breadAllowCheckoutCart":       shop.AllowCheckoutCart,
		"breadHealthcareMode":          shop.HealthcareMode,
		"targetedFinancing":            shop.TargetedFinancing,
		"targetedFinancingID":          shop.TargetedFinancingID,
		"targetedFinancingThreshold":   shop.TargetedFinancingThreshold,
		"draftOrderTax":                shop.DraftOrderTax,
		"acceleratedCheckoutPermitted": shop.AcceleratedCheckoutPermitted,
		"remainderPayAutoCancel":       shop.RemainderPayAutoCancel,
		"integrationKey":               shop.IntegrationKey,
		"sandboxIntegrationKey":        shop.SandboxIntegrationKey,
		"platformBreadTestMode":        !shop.PlatformProduction,
		"activeVersion":                shop.ActiveVersion,
	})
}
