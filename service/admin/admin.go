package admin

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	jwt "github.com/gbrlsnchs/jwt/v2"
	"github.com/getbread/breadkit/desmond"
	"github.com/getbread/shopify_plugin_backend/service/app"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	log "github.com/sirupsen/logrus"
)

var (
	oauthClientID     string
	oauthClientSecret string
	store             *sessions.CookieStore
	MiltonHost        string
)

var oauthScopes = []string{"profile", "email", "openid"}

const SESSION_NAME = "milton-admin"

func InitConfig(ClientId string, ClientSecret string, MiltonAdminSessionSecret string, miltonHost string) {
	oauthClientID = ClientId
	oauthClientSecret = ClientSecret
	store = sessions.NewCookieStore([]byte(MiltonAdminSessionSecret))
	store.MaxAge(3600) // Set cookie Max-Age to 1 hour
	store.Options.HttpOnly = true
	store.Options.Secure = true
	MiltonHost = miltonHost
}

func (h *Handlers) AuthMiddleware(c *gin.Context, dc desmond.Context) {
	session, _ := store.Get(c.Request, SESSION_NAME)
	trusted := session.Values["trusted"]
	if trusted != true {
		c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("Not an unathorized user"))
		return
	}
	c.Next()
	return
}

func (h *Handlers) AdminPortal(c *gin.Context, dc desmond.Context) {
	// Check for existing session - Get() always returns a session, even if empty
	session, _ := store.Get(c.Request, SESSION_NAME)
	trusted := session.Values["trusted"]

	if trusted != true {
		csrf := generateToken()
		session.Values["state"] = csrf
		err := session.Save(c.Request, c.Writer)
		if err != nil {
			log.WithFields(log.Fields{
				"token": csrf,
				"error": err.Error(),
			}).Error("(AdminPortal) saving session token produced an error")
			c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("Internal server error"))
			return
		}

		oauthLink, err := generateGoogleAuthURL(csrf)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("(AdminPortal) generating google authorization URL produced an error")
			c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("Internal server error"))
			return
		}
		c.HTML(http.StatusOK, "login.html", gin.H{
			"oauthURL": oauthLink,
		})
		return
	}

	c.HTML(http.StatusOK, "admin.html", gin.H{})
	return
}

type googleIDToken struct {
	Issuer              string `json:"iss"`
	AtHash              string `json:"at_hash"`
	EmailVerified       bool   `json:"email_verified"`
	Subject             string `json:"sub"`
	AuthorizedPresenter string `json:"azp"`
	Email               string `json:"email"`
	Audience            string `json:"aud"`
	IssuedAt            int64  `json:"iat"`
	Expiration          int64  `json:"exp"`
	Nonce               string `json:"nonce"`
	HostDomain          string `json:"hd"`
}

func (h *Handlers) HandleOauthRedirect(c *gin.Context, dc desmond.Context) {
	session, _ := store.Get(c.Request, SESSION_NAME)
	q := c.Request.URL.Query()

	// Check if state matches
	csrf := q.Get("state")
	if csrf != session.Values["state"] {
		// Logging as error for Sentry notification
		log.WithFields(log.Fields{
			"csrf":         csrf,
			"sessionState": session.Values["state"],
		}).Error("(HandleOauthRedirect) CSRF check failed: request state does not match session state")
		c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("Unauthorized"))
		return
	}

	// Exchange one-time Code for access_token and id
	authCode := q.Get("code")
	res, err := exchangeAuthCodeForAccessToken(authCode)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("(HandleOauthRedirect) exchanging auth code for access token produced an error")
		c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("Internal server error"))
		return
	}

	payload, _, err := jwt.Parse(res.IdentityToken)
	if err != nil {
		log.WithFields(log.Fields{
			"id_token": res.IdentityToken,
			"error":    err.Error(),
		}).Error("(HandleOauthRedirect) parsing id_token produced an error")
		c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("Internal server error"))
		return
	}

	var profile *googleIDToken
	err = jwt.Unmarshal(payload, &profile)
	if err != nil {
		log.WithFields(log.Fields{
			"payload": payload,
			"error":   err.Error(),
		}).Error("(HandleOauthRedirect) unmarshaling id_token produced an error")
		c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("Internal server error"))
		return
	}

	if profile.HostDomain != "breadfinance.com" {
		// Logging as error for Sentry notification
		log.WithFields(log.Fields{
			"hostDomain":   profile.HostDomain,
			"sessionState": session.Values["state"],
		}).Error("(HandleOauthRedirect) CSRF check failed: request state does not match session state")
		c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("Unauthorized"))
		return
	}

	session.Values["hd"] = profile.HostDomain
	session.Values["email"] = profile.Email
	session.Values["trusted"] = true
	err = session.Save(c.Request, c.Writer)
	if err != nil {
		log.WithFields(log.Fields{
			"profile": profile,
			"error":   err.Error(),
		}).Error("(HandleOauthRedirect) saving session produced an error")
		c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("Internal server error"))
		return
	}
	c.Redirect(http.StatusFound, "/admin")
	return
}

func (h *Handlers) PortalLogout(c *gin.Context, dc desmond.Context) {
	session, _ := store.Get(c.Request, SESSION_NAME)
	session.Values["trusted"] = false
	err := session.Save(c.Request, c.Writer)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		})
	}
	c.Redirect(http.StatusFound, "/admin")
	return
}

func (h *Handlers) GetShopifyShops(c *gin.Context, dc desmond.Context) {
	shops, err := findAllShops(h)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Unable to process request",
		})
		return
	}

	formattedShops := make([]map[string]interface{}, len(shops))
	for i, shop := range shops {
		formattedShops[i] = map[string]interface{}{
			"shopName":            shop.Shop,
			"acceleratedCheckout": shop.AcceleratedCheckoutPermitted,
			"posAccess":           shop.POSAccess,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"shops":  formattedShops,
	})
	return
}

type shopSettingRequest struct {
	ShopName                  string `json:"shop"`
	EnableAcceleratedCheckout bool   `json:"enableAcceleratedCheckout"`
	POSAccess                 bool   `json:"posAccess"`
}

func (h *Handlers) UpdateShopSettings(c *gin.Context, dc desmond.Context) {
	session, _ := store.Get(c.Request, SESSION_NAME)
	user := session.Values["email"]

	var req shopSettingRequest
	err := c.BindJSON(&req)
	if err != nil {
		log.WithFields(log.Fields{
			"user":  user,
			"error": err.Error(),
		}).Error("(UpdateShopSettings) binding request produced an error")
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Bad request please try again",
		})
		return
	}

	err = saveShopSettings(req, h)
	if err != nil {
		log.WithFields(log.Fields{
			"user":          user,
			"shop":          req.ShopName,
			"allowCheckout": req.EnableAcceleratedCheckout,
			"error":         err.Error(),
		}).Error("(UpdateShopSettings) updating accelerated checkout permission setting produced an error")
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Unable to process request",
		})
		return
	}

	log.WithFields(log.Fields{
		"user":    user,
		"shop":    req.ShopName,
		"action":  "UpdateShopSettings",
		"setting": "acceleratedCheckoutPermission",
		"value":   req.EnableAcceleratedCheckout,
	}).Info("(MiltonAudit) recording admin action")

	c.JSON(http.StatusOK, gin.H{
		"shopName": req.ShopName,
	})
	return
}

func generateGoogleAuthURL(state string) (string, error) {
	dd, err := getGoogleDiscoveryDocument()
	if err != nil {
		return "", err
	}

	redirectURI := url.QueryEscape(getOAuthRedirectURI())
	scopes := url.QueryEscape(strings.Join(oauthScopes, " "))
	return fmt.Sprintf("%s?client_id=%s&scope=%s&state=%s&redirect_uri=%s&response_type=code", dd.AuthorizationEndpoint, oauthClientID, scopes, state, redirectURI), nil
}

func getOAuthRedirectURI() string {
	return MiltonHost + "/admin/authenticate"
}

func generateToken() string {
	token := make([]byte, 32)
	rand.Read(token)
	return base64.URLEncoding.EncodeToString(token)
}

type webhookRequest struct {
	ShopName string `json:"shop"`
}

func (h *Handlers) RegisterWebhooks(c *gin.Context, dc desmond.Context) {

	var req webhookRequest
	err := c.BindJSON(&req)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("(RegisterWebhooks) binding request produced an error")
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Bad request please try again",
		})
		return
	}

	shopName := req.ShopName
	shop, err := findShopByName(shopName, h)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error(fmt.Sprintf("(RegisterWebhooks) query for shop (%s) produced error", shopName))
		c.String(400, "An error occurred, please try again")
		return
	}

	_, err = app.UpdateWebhooksExt(shop, true)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"error": err,
			"shop":  shop,
		}).Info("(RegisterWebhooks) checking webhooks produced error")
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": fmt.Sprintf("(RegisterWebhooks) Webhook registration successful: %s", shopName),
	})

	return
}
