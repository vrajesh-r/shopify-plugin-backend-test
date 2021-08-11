package app

import (
	"fmt"

	"github.com/getbread/breadkit/desmond"
	"github.com/getbread/breadkit/featureflags"
	"github.com/getbread/shopify_plugin_backend/service/update"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type updateSettingsRequest struct {
	BreadApiKey                string `json:"breadApiKey"`
	BreadSecretKey             string `json:"breadApiSecret"`
	BreadSandboxApiKey         string `json:"breadSandboxApiKey"`
	BreadSandboxSecretKey      string `json:"breadSandboxSecretKey"`
	CustomCSS                  string `json:"customCss"`
	CustomCSSCart              string `json:"customCssCart"`
	AutoAuthorize              bool   `json:"autoAuthorize"`
	AutoSettle                 bool   `json:"autoSettle"`
	ActsAsLabel                bool   `json:"actsAsLabel"`
	CreateCustomers            bool   `json:"createCustomers"`
	Production                 bool   `json:"production"`
	ManualEmbedScript          bool   `json:"manualEmbedScript"`
	AsLowAs                    bool   `json:"asLowAs"`
	EnableOrderWebhooks        bool   `json:"enableOrderWebhooks"`
	AllowCheckoutPDP           bool   `json:"allowCheckoutPDP"`
	EnableAddToCart            bool   `json:"enableAddToCart"`
	AllowCheckoutCart          bool   `json:"allowCheckoutCart"`
	HealthcareMode             bool   `json:"healthcareMode"`
	TargetedFinancing          bool   `json:"targetedFinancing"`
	TargetedFinancingID        string `json:"targetedFinancingID"`
	TargetedFinancingThreshold int64  `json:"targetedFinancingThreshold"`
	DraftOrderTax              bool   `json:"draftOrderTax"`
	RemainderPayAutoCancel     bool   `json:"remainderPayAutoCancel"`
	IntegrationKey             string `json:"integrationKey"`
	SandboxIntegrationKey      string `json:"sandboxIntegrationKey"`
	PlatformProduction         bool   `json:"platformProduction"`
}

type updateVersionRequest struct {
	ActiveVersion string `json:"activeVersion"`
}

func (h *Handlers) UpdateSettings(c *gin.Context, dc desmond.Context) {
	version := c.Params.ByName("version")

	// find valid session
	var sessionId string
	if cookie, err := c.Request.Cookie(ADMIN_COOKIE_NAME); err == nil {
		sessionId = cookie.Value
	}

	// Short circuit error logging for empty session id
	if sessionId == "" {
		c.String(400, "empty session id")
		return
	}

	session, err := findValidSessionById(sessionId, h)
	if err != nil {
		log.WithFields(log.Fields{
			"error":     err.Error(),
			"sessionId": sessionId,
		}).Error("(UpdateSettings) query for session produced error")
		c.String(400, err.Error())
		return
	}

	// find shop
	shop, err := findShopById(session.ShopId, h)
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err.Error(),
			"session": session,
		}).Error("(UpdateSettings) query for shop produced error")
		c.String(400, err.Error())
		return
	}

	var req updateSettingsRequest

	err = c.BindJSON(&req)
	if err != nil {
		log.WithField("error", err.Error()).Error("(UpdateSettings) binding request to model produced error")
		c.String(400, err.Error())
		return
	}

	activeVersion := shop.ActiveVersion
	if activeVersion == "" {
		activeVersion = BreadClassic
	}

	if req.ManualEmbedScript != shop.ManualEmbedScript && req.ManualEmbedScript {
		//Settings has been updated to manually embed scripted
		// Delete all embedded added scripts
		if err := removeAllEmbeddedScripts(shop); err != nil {
			log.WithFields(log.Fields{
				"error":   err.Error(),
				"session": session,
				"request": req,
			}).Error("(UpdateSettings) removing embedded scripts failed")
			c.String(400, err.Error())
			return
		}
	} else if version == activeVersion { // The settings of the active version is being updated
		if featureflags.GetBool("milton-enable-asset-caching", false) {
			if err := updateEmbeddedScriptFromSettings(shop, req); err != nil {
				log.WithFields(log.Fields{
					"error":   err.Error(),
					"session": session,
					"request": req,
				}).Error("(UpdateSettings) updating embedded script failed")
				c.String(500, err.Error())
				return
			}
		} else {
			// add embedded script
			if err := embedScriptTag(shop); err != nil {
				log.WithFields(log.Fields{
					"error":   err.Error(),
					"session": session,
					"request": req,
				}).Error("(UpdateSettings) adding embedded script failed")
				c.String(400, err.Error())
				return
			}
		}
	}

	// update shop
	sur := update.ShopUpdateRequest{
		Id:      shop.Id,
		Updates: map[update.ShopUpdateField]interface{}{},
	}

	sur.Updates[update.ShopUpdate_CreateCustomers] = req.CreateCustomers
	sur.Updates[update.ShopUpdate_Production] = req.Production
	sur.Updates[update.ShopUpdate_ManualEmbedScript] = req.ManualEmbedScript
	sur.Updates[update.ShopUpdate_AsLowAs] = req.AsLowAs
	sur.Updates[update.ShopUpdate_EnableOrderWebhooks] = req.EnableOrderWebhooks

	if version == BreadClassic {
		sur.Updates[update.ShopUpdate_BreadApiKey] = req.BreadApiKey
		sur.Updates[update.ShopUpdate_BreadSecretKey] = req.BreadSecretKey
		sur.Updates[update.ShopUpdate_BreadSandboxApiKey] = req.BreadSandboxApiKey
		sur.Updates[update.ShopUpdate_BreadSandboxSecretKey] = req.BreadSandboxSecretKey
		sur.Updates[update.ShopUpdate_CSS] = req.CustomCSS
		sur.Updates[update.ShopUpdate_CSSCart] = req.CustomCSSCart
		sur.Updates[update.ShopUpdate_AutoSettle] = req.AutoSettle
		sur.Updates[update.ShopUpdate_ActsAsLabel] = req.ActsAsLabel
		sur.Updates[update.ShopUpdate_Production] = req.Production
		sur.Updates[update.ShopUpdate_AsLowAs] = req.AsLowAs
		sur.Updates[update.ShopUpdate_AllowCheckoutPDP] = req.AllowCheckoutPDP
		sur.Updates[update.ShopUpdate_EnableAddToCart] = req.EnableAddToCart
		sur.Updates[update.ShopUpdate_AllowCheckoutCart] = req.AllowCheckoutCart
		sur.Updates[update.ShopUpdate_HealthcareMode] = req.HealthcareMode
		sur.Updates[update.ShopUpdate_TargetedFinancing] = req.TargetedFinancing
		sur.Updates[update.ShopUpdate_TargetedFinancingID] = req.TargetedFinancingID
		sur.Updates[update.ShopUpdate_TargetedFinancingThreshold] = req.TargetedFinancingThreshold
		sur.Updates[update.ShopUpdate_DraftOrderTax] = req.DraftOrderTax
		sur.Updates[update.ShopUpdate_RemainderPayAutoCancel] = req.RemainderPayAutoCancel
	} else { //version  is platform
		sur.Updates[update.ShopUpdate_IntegrationKey] = req.IntegrationKey
		sur.Updates[update.ShopUpdate_SandboxIntegrationKey] = req.SandboxIntegrationKey
		sur.Updates[update.ShopUpdate_PlatformProduction] = req.PlatformProduction
	}

	if err := h.ShopUpdater.Update(sur); err != nil {
		log.WithFields(log.Fields{
			"error":   fmt.Sprintf("%+v", err),
			"session": fmt.Sprintf("%+v", session),
			"shop":    fmt.Sprintf("%+v", shop),
		}).Error("(UpdateSettings) updating shop produced en error")
		c.String(400, err.Error())
		return
	}

	// If the store-products list changed, we need to update the webhooks.
	if shop.EnableOrderWebhooks != req.EnableOrderWebhooks {
		shop.EnableOrderWebhooks = req.EnableOrderWebhooks
		if _, err := updateWebhooks(shop, true); err != nil {
			log.WithFields(log.Fields{
				"error":   fmt.Sprintf("%+v", err),
				"session": fmt.Sprintf("%+v", session),
				"shop":    fmt.Sprintf("%+v", shop),
			}).Errorf("Failed to update the webhooks due to error: %s", err.Error())

			// Put the store-products flag back the way it was.
			sur := update.ShopUpdateRequest{
				Id:      shop.Id,
				Updates: map[update.ShopUpdateField]interface{}{},
			}
			sur.Updates[update.ShopUpdate_EnableOrderWebhooks] = shop.EnableOrderWebhooks
			if err := h.ShopUpdater.Update(sur); err != nil {
				log.WithFields(log.Fields{
					"error":   fmt.Sprintf("%+v", err),
					"session": fmt.Sprintf("%+v", session),
					"shop":    fmt.Sprintf("%+v", shop),
				}).Errorf("Failed to reset the store-products flag due to error: %s", err.Error())
			}
			c.String(500, "Failed to update the store-products flag.")
			return
		}
	}

	c.JSON(200, gin.H{
		"complete": true,
	})
}

func (h *Handlers) UpdateVersion(c *gin.Context, dc desmond.Context) {
	// find valid session
	var sessionId string
	if cookie, err := c.Request.Cookie(ADMIN_COOKIE_NAME); err == nil {
		sessionId = cookie.Value
	}

	// Short circuit error logging for empty session id
	if sessionId == "" {
		c.String(400, "empty session id")
		return
	}

	session, err := findValidSessionById(sessionId, h)
	if err != nil {
		log.WithFields(log.Fields{
			"error":     err.Error(),
			"sessionId": sessionId,
		}).Error("(UpdateSettings) query for session produced error")
		c.String(400, err.Error())
		return
	}

	// find shop
	shop, err := findShopById(session.ShopId, h)
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err.Error(),
			"session": session,
		}).Error("(UpdateSettings) query for shop produced error")
		c.String(400, err.Error())
		return
	}

	var req updateVersionRequest

	err = c.BindJSON(&req)
	if err != nil {
		log.WithField("error", err.Error()).Error("(UpdateSettings) binding request to model produced error")
		c.String(400, err.Error())
		return
	}

	if !shop.ManualEmbedScript {
		if featureflags.GetBool("milton-enable-asset-caching", false) {
			if err := updateEmbeddedScriptFromVersion(shop, req); err != nil {
				log.WithFields(log.Fields{
					"error":   err.Error(),
					"session": session,
					"request": req,
				}).Error("(UpdateSettings) updating embedded script failed")
				c.String(500, err.Error())
				return
			}
		} else {
			// update embedded script
			if err := embedScriptTagFromVersion(shop, req); err != nil {
				log.WithFields(log.Fields{
					"error":   err.Error(),
					"session": session,
					"request": req,
				}).Error("(UpdateSettings) adding embedded script failed")
				c.String(400, err.Error())
				return
			}
		}
	}

	// update shop
	sur := update.ShopUpdateRequest{
		Id:      shop.Id,
		Updates: map[update.ShopUpdateField]interface{}{},
	}
	sur.Updates[update.ShopUpdate_ActiveVersion] = req.ActiveVersion

	if err := h.ShopUpdater.Update(sur); err != nil {
		log.WithFields(log.Fields{
			"error":   fmt.Sprintf("%+v", err),
			"session": fmt.Sprintf("%+v", session),
			"shop":    fmt.Sprintf("%+v", shop),
		}).Error("(UpdateSettings) updating shop produced en error")
		c.String(400, err.Error())
		return
	}

	c.JSON(200, gin.H{
		"complete": true,
	})
}
