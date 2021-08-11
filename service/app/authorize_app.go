package app

import (
	"strings"

	"github.com/getbread/breadkit/desmond"
	"github.com/getbread/shopify_plugin_backend/service/shopify"
	"github.com/getbread/shopify_plugin_backend/service/types"
	"github.com/getbread/shopify_plugin_backend/service/update"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type shopifyAccessTokenResponse struct {
	AccessToken string `json:"access_token"`
}

func (h *Handlers) AuthorizeApp(c *gin.Context, dc desmond.Context) {
	// pull params from query
	authCode := c.Query("code")
	state := c.Query("state")
	shopUrl := c.Query("shop")
	shopName := strings.Split(shopUrl, ".")[0]

	// pull shop
	var shop types.Shop
	shop, err := findShopByName(shopName, h)
	if err != nil {
		log.WithFields(log.Fields{
			"error":    err.Error(),
			"shopUrl":  shopUrl,
			"shopName": shopName,
		}).Error("(AuthorizeApp) query for shop produced error")
		c.String(400, err.Error())
		return
	}

	// use nonce
	err = useNonceByShopId(state, shop.Id, h)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
			"shop":  shop,
			"nonce": state,
		}).Error("(AuthorizeApp) using nonce failed")
		c.String(400, err.Error())
		return
	}

	// shopify client
	sc := shopify.NewClient(shop.Shop, shop.AccessToken)

	// exchange auth code for access token
	oauthReq := &shopify.OAuthExchangeRequest{
		ClientID:     appConfig.ShopifyConfig.ShopifyApiKey.Unmask(),
		ClientSecret: appConfig.ShopifyConfig.ShopifySharedSecret.Unmask(),
		Code:         authCode,
	}
	var oauthRes shopify.OAuthExchangeResponse
	if err = sc.ExchangeOAuthCode(oauthReq, &oauthRes); err != nil {
		log.WithFields(log.Fields{
			"error":    err.Error(),
			"shop":     shop,
			"authCode": authCode,
		}).Error("(AuthorizeApp) exchanging oauth code failed")
		c.String(400, err.Error())
		return
	}
	accessToken := oauthRes.AccessToken

	// save access token on shop
	sur := update.ShopUpdateRequest{
		Id: shop.Id,
		Updates: map[update.ShopUpdateField]interface{}{
			update.ShopUpdate_AccessToken:              accessToken,
			update.ShopUpdate_OAuthPermissionsUpToDate: true,
		},
	}
	err = h.ShopUpdater.Update(sur)
	if err != nil {
		log.WithFields(log.Fields{
			"error":       err.Error(),
			"shop":        shop,
			"accessToken": accessToken,
		}).Error("(AuthorizeApp) updating shop with accessToken produced error")
		c.String(400, err.Error())
		return
	}
	sc.AccessToken = accessToken
	shop.AccessToken = accessToken

	// embed scripts
	if err = embedScriptTag(shop); err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
			"shop":  shop,
		}).Error("(AuthorizeApp) embedding script produced error")
		c.String(400, err.Error())
		return
	}

	// respond to client
	c.Redirect(302, shopify.AppAdminUrl(shop.Shop))
}
