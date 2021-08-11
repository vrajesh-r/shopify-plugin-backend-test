package gateway

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/getbread/breadkit/desmond"
	"github.com/getbread/breadkit/zeus/searcher"
	zeus "github.com/getbread/breadkit/zeus/types"
	"github.com/getbread/shopify_plugin_backend/service/search"
	"github.com/getbread/shopify_plugin_backend/service/types"
	"github.com/getbread/shopify_plugin_backend/service/update"
	"github.com/gin-gonic/gin"
	"github.com/pborman/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

func emailIsValid(email string) bool {
	Re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return Re.MatchString(email)
}

func setGatewayPortalCookie(session types.GatewaySession, c *gin.Context) {
	// add session cookie to HTTP response
	expire := time.Unix(session.Expiration, 0)
	cookie := &http.Cookie{
		Name:       GATEWAY_PORTAL_COOKIE_NAME,
		Value:      string(session.Id),
		Expires:    expire,
		RawExpires: expire.Format(time.UnixDate),
		Path:       "/",
	}
	http.SetCookie(c.Writer, cookie)
}

func newGatewayPortalSession(account types.GatewayAccount, h *Handlers) (types.GatewaySession, error) {
	session := types.GatewaySession{
		GatewayAccountID: account.Id,
		Expiration:       types.GenerateGatewaySessionExpiration(),
	}
	sessionID, err := h.GatewaySessionCreator.Create(session)
	if err != nil {
		logrus.WithError(err).WithField("account", account).
			Error("(GatewaySignUp) saving http session produced error")
		return types.GatewaySession{}, err
	}
	session.Id = sessionID
	return session, nil
}

func verifySession(sessionID string, h *Handlers) (string, error) {
	// find a valid session
	gssr := search.GatewaySessionSearchRequest{}
	gssr.AddFilter(search.GatewaySessionSearch_Id, zeus.Uuid(sessionID), searcher.Operator_EQ, searcher.Condition_AND)
	gssr.AddFilter(search.GatewaySessionSearch_Expiration, time.Now().Unix(), searcher.Operator_GT, searcher.Condition_AND)
	gssr.Limit = 1
	gatewaySessions, err := h.GatewaySessionSearcher.Search(gssr)
	if err != nil {
		return "", err
	}
	if len(gatewaySessions) == 0 {
		return "", fmt.Errorf("no gateway session found")
	}
	gs := gatewaySessions[0]

	// refresh session
	// log and ignore error
	go func(gatewaySession types.GatewaySession, h *Handlers) {
		if err := refreshSession(gs, h); err != nil {
			logrus.WithError(err).WithField("gatewaySession", gs).
				Error("(verifySession) session refresh produced error")
		}
		return
	}(gs, h)

	return string(gs.GatewayAccountID), nil
}

// Takes in a Gateway Session and refreshes expiration
func refreshSession(gs types.GatewaySession, h *Handlers) error {
	increase := 60 * 20 // 20 minutes
	gsur := update.GatewaySessionUpdateRequest{
		Id:      gs.Id,
		Updates: map[update.GatewaySessionUpdateField]interface{}{},
	}
	gsur.Updates[update.GatewaySessionUpdate_Expiration] = time.Now().Unix() + int64(increase)
	return h.GatewaySessionUpdater.Update(gsur)
}

func generateGatewayKeyPair() (string, string) {
	return uuid.New(), uuid.New()
}

func (h *Handlers) GatewayIsLoggedIn(c *gin.Context, dc desmond.Context) {
	var sessionID string
	if cookie, err := c.Request.Cookie(GATEWAY_PORTAL_COOKIE_NAME); err == nil {
		sessionID = cookie.Value
	}
	_, err := verifySession(sessionID, h)
	if err != nil {
		c.String(401, "")
		return
	}
	c.JSON(200, gin.H{})
	return
}

func (h *Handlers) GatewayLogOut(c *gin.Context, dc desmond.Context) {
	session := types.GatewaySession{}
	setGatewayPortalCookie(session, c)
	c.JSON(200, gin.H{})
}

func (h *Handlers) GatewaySignUp(c *gin.Context, dc desmond.Context) {
	genericErrorResponse := "There was error processing your request, please try again or reach out to support@getbread.com"

	// Parse request
	var req gatewayAccountSignUpRequest
	if err := c.BindJSON(&req); err != nil {
		logrus.WithError(err).Error("(GatewaySignUp) binding request produced error")
		c.JSON(400, gin.H{
			"error": genericErrorResponse,
		})
		return
	}

	// Sanitize
	if req.Password != req.PasswordVerify {
		logrus.WithFields(logrus.Fields{
			"info":  "New gateway account request password did not match",
			"email": req.Email,
		}).Info("(GatewaySignUp) New account request passwords did not match")
		c.JSON(400, gin.H{
			"error": "Passwords do not match.",
		})
		return
	}
	if !emailIsValid(req.Email) {
		c.JSON(400, gin.H{
			"error": "Email invalid.",
		})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), HASH_COST)
	if err != nil {
		logrus.WithError(err).WithField("email", req.Email).
			Error("(GatewaySignUp) hashing password produced error")
		c.JSON(400, gin.H{
			"error": genericErrorResponse,
		})
		return
	}

	// Create shop
	gatewayKey, gatewaySecret := generateGatewayKeyPair()
	account := types.GatewayAccount{
		Email:         req.Email,
		PasswordHash:  string(hashedPassword),
		GatewayKey:    gatewayKey,
		GatewaySecret: gatewaySecret,
	}
	accountID, err := h.GatewayAccountCreator.Create(account)
	if err != nil {
		logrus.WithError(err).WithField("email", req.Email).Error("(GatewaySignUp) creating gateway account produced error")
		c.JSON(400, gin.H{
			"error": genericErrorResponse,
		})
		return
	}
	account.Id = accountID

	// Create session
	session, err := newGatewayPortalSession(account, h)
	if err != nil {
		logrus.WithError(err).WithField("account", account).
			Error("(GatewaySignUp) creating a gateway sesion produced error")
		c.JSON(400, gin.H{
			"error": genericErrorResponse,
		})
		return
	}

	// Add session cookie to HTTP response
	setGatewayPortalCookie(session, c)

	// Redirect to management page
	c.JSON(200, gin.H{})
}

func (h *Handlers) GatewaySignIn(c *gin.Context, dc desmond.Context) {
	genericErrorResponse := "There was error with sign in, please try again or reach out to support@getbread.com"

	var req gatewayAccountSignInRequest
	if err := c.Bind(&req); err != nil {
		logrus.WithError(err).Error("(GatewaySignIn) binding request produced error")
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Query gateway account
	gasr := search.GatewayAccountSearchRequest{}
	gasr.AddFilter(search.GatewayAccountSearch_Email, req.Email, searcher.Operator_EQ, searcher.Condition_AND)
	gasr.Limit = 1
	accounts, err := h.GatewayAccountSearcher.Search(gasr)
	if err != nil {
		logrus.WithError(err).WithField("email", req.Email).Error("(GatewaySignIn) searching for gateway account produced error")
		c.JSON(400, gin.H{
			"error": genericErrorResponse,
		})
		return
	}
	if len(accounts) == 0 {
		logrus.WithField("email", req.Email).Error("(GatewaySignIn) search for gateway account returned none")
		c.JSON(400, gin.H{
			"error": "Gateway account was not found, please verify credentials & try again",
		})
		return
	}
	account := accounts[0]

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(req.Password)); err != nil {
		logrus.WithError(err).WithField("email", req.Email).
			Error("(GatewaySignIn) password match failed")
		c.JSON(400, gin.H{
			"error": "Credentials are incorrect, please try again.",
		})
		return
	}

	// Create and save session
	session, err := newGatewayPortalSession(account, h)
	if err != nil {
		logrus.WithError(err).WithField("account", account).
			Error("(GatewaySignIn) creating a gateway sesion produced error")
		c.JSON(400, gin.H{
			"error": genericErrorResponse,
		})
		return
	}

	setGatewayPortalCookie(session, c)

	c.JSON(200, gin.H{})
}

func (h *Handlers) PullGatewayAccount(c *gin.Context, dc desmond.Context) {
	// Verify session on request is valid
	var sessionID string
	cookie, err := c.Request.Cookie(GATEWAY_PORTAL_COOKIE_NAME)

	if err == nil {
		sessionID = cookie.Value
	}
	if err != nil {
		logrus.WithError(err).Error("(PullGatewayAccount) parsing cookie produced error")
	}
	accountID, err := verifySession(sessionID, h)
	if err != nil {
		logrus.WithError(err).WithField("sessionID", sessionID).
			Info("(PullGatewayAccount) verifying session produced error")
		c.String(401, "")
		return
	}

	gasr := search.GatewayAccountSearchRequest{}
	gasr.AddFilter(search.GatewayAccountSearch_Id, accountID, searcher.Operator_EQ, searcher.Condition_AND)
	gasr.Limit = 1
	accounts, err := h.GatewayAccountSearcher.Search(gasr)
	// Not 100% on how to handle this error, for now just redirecting to /gateway/signin
	if err != nil || len(accounts) == 0 {
		var errorS string
		if err != nil {
			errorS = err.Error()
		} else {
			errorS = "no gateway account found"
		}
		logrus.WithFields(logrus.Fields{
			"error":     errorS,
			"sessionID": sessionID,
			"accountID": accountID,
		}).Error("(PullGatewayAccount) search for gateway account produced error")
		c.String(401, "")
		return
	}
	account := accounts[0]

	c.JSON(200, gin.H{
		"gatewayKey":                  account.GatewayKey,
		"gatewaySecret":               account.GatewaySecret,
		"apiKey":                      account.ApiKey,
		"sharedSecret":                h.maskSecret(account.SharedSecret),
		"sandboxApiKey":               account.SandboxApiKey,
		"sandboxSharedSecret":         h.maskSecret(account.SandboxSharedSecret),
		"autoSettle":                  account.AutoSettle,
		"healthcareMode":              account.HealthcareMode,
		"targetedFinancing":           account.TargetedFinancing,
		"targetedFinancingID":         account.TargetedFinancingID,
		"targetedFinancingThreshold":  account.TargetedFinancingThreshold,
		"plusEmbeddedCheckout":        account.PlusEmbeddedCheckout,
		"production":                  account.Production,
		"remainderPayAutoCancel":      account.RemainderPayAutoCancel,
		"platformApiKey":              account.PlatformApiKey,
		"platformSharedSecret":        h.maskSecret(account.PlatformSharedSecret),
		"platformSandboxApiKey":       account.PlatformSandboxApiKey,
		"platformSandboxSharedSecret": h.maskSecret(account.PlatformSandboxSharedSecret),
		"platformAutoSettle":          account.PlatformAutoSettle,
		"activeVersion":               account.ActiveVersion,
		"integrationKey":              account.IntegrationKey,
		"sandboxIntegrationKey":       account.SandboxIntegrationKey,
	})
}

func (h *Handlers) maskSecret(secret string) string {
	var masked string
	for i, char := range secret {
		var maskedChar string
		if i < len(secret)-7 && fmt.Sprintf("%c", char) != "-" {
			maskedChar = "*"
		} else {
			maskedChar = fmt.Sprintf("%c", char)
		}
		masked += maskedChar
	}
	return masked
}

func (h *Handlers) UpdateGatewayAccount(c *gin.Context, dc desmond.Context) {
	version := c.Params.ByName("version")
	// Verify session on request is valid
	var sessionID string
	if cookie, err := c.Request.Cookie(GATEWAY_PORTAL_COOKIE_NAME); err == nil {
		sessionID = cookie.Value
	}
	accountID, err := verifySession(sessionID, h)
	if err != nil {
		logrus.WithError(err).WithField("sessionID", sessionID).Info("(UpdateGatewayAccount) verifying session produced error")
		c.String(401, "")
		return
	}

	gasr := search.GatewayAccountSearchRequest{}
	gasr.AddFilter(search.GatewayAccountSearch_Id, accountID, searcher.Operator_EQ, searcher.Condition_AND)
	gasr.Limit = 1
	accounts, err := h.GatewayAccountSearcher.Search(gasr)
	// Not 100% on how to handle this error, for now just redirecting to /gateway/signin
	if err != nil || len(accounts) == 0 {
		var errorS string
		if err != nil {
			errorS = err.Error()
		} else {
			errorS = "no gateway account found"
		}
		logrus.WithFields(logrus.Fields{
			"error":     errorS,
			"sessionID": sessionID,
			"accountID": accountID,
		}).Error("(UpdateGatewayAccount) search for gateway account produced error")
		c.String(401, "")
		return
	}
	account := accounts[0]

	var req updateGatewayAccountRequest
	if err := c.Bind(&req); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"sessionID": sessionID,
			"accountID": accountID,
		}).Error("(UpdateGatewayAccount) binding request produced error")
		c.JSON(400, gin.H{
			"error": "Improper request, please try again.",
		})
		return
	}

	// Commit updates
	gaur := update.GatewayAccountUpdateRequest{
		Id:      zeus.Uuid(accountID),
		Updates: map[update.GatewayAccountUpdateField]interface{}{},
	}

	if version == BreadClassic {
		gaur.Updates[update.GatewayAccountUpdate_ApiKey] = req.ApiKey
		if !strings.Contains(req.SharedSecret, "*") {
			gaur.Updates[update.GatewayAccountUpdate_SharedSecret] = req.SharedSecret
		}
		gaur.Updates[update.GatewayAccountUpdate_SandboxApiKey] = req.SandboxApiKey
		if !strings.Contains(req.SandboxSharedSecret, "*") {
			gaur.Updates[update.GatewayAccountUpdate_SandboxSharedSecret] = req.SandboxSharedSecret
		}
		gaur.Updates[update.GatewayAccountUpdate_AutoSettle] = req.AutoSettle
		gaur.Updates[update.GatewayAccountUpdate_HealthcareMode] = req.HealthcareMode
		gaur.Updates[update.GatewayAccountUpdate_TargetedFinancing] = req.TargetedFinancing
		gaur.Updates[update.GatewayAccountUpdate_TargetedFinancingID] = req.TargetedFinancingID
		gaur.Updates[update.GatewayAccountUpdate_TargetedFinancingThreshold] = req.TargetedFinancingThreshold
		gaur.Updates[update.GatewayAccountUpdate_PlusEmbeddedCheckout] = req.PlusEmbeddedCheckout
		gaur.Updates[update.GatewayAccountUpdate_Production] = req.Production
		gaur.Updates[update.GatewayAccountUpdate_RemainderPayAutoCancel] = req.RemainderPayAutoCancel
	} else { // version is BreadPlatform
		gaur.Updates[update.GatewayAccountUpdate_PlatformApiKey] = req.PlatformApiKey
		if !strings.Contains(req.PlatformSharedSecret, "*") {
			gaur.Updates[update.GatewayAccountUpdate_PlatformSharedSecret] = req.PlatformSharedSecret
		}
		gaur.Updates[update.GatewayAccountUpdate_PlatformSandboxApiKey] = req.PlatformSandboxApiKey
		if !strings.Contains(req.PlatformSandboxSharedSecret, "*") {
			gaur.Updates[update.GatewayAccountUpdate_PlatformSandboxSharedSecret] = req.PlatformSandboxSharedSecret
		}
		gaur.Updates[update.GatewayAccountUpdate_IntegrationKey] = req.IntegrationKey
		gaur.Updates[update.GatewayAccountUpdate_SandboxIntegrationKey] = req.SandboxIntegrationKey

		gaur.Updates[update.GatewayAccountUpdate_PlatformAutoSettle] = req.PlatformAutoSettle
	}

	if err = h.GatewayAccountUpdater.Update(gaur); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"sessionID": sessionID,
			"accountID": account.Id,
		}).Error("(UpgateGatewayAccount) updating gateway account produced error")
		c.JSON(400, gin.H{
			"error": "An unexpected error occurred, please try again",
		})
		return
	}
	c.String(204, "")
}

func (h *Handlers) UpdateGatewayAccountVersion(c *gin.Context, dc desmond.Context) {
	// Verify session on request is valid
	var sessionID string
	if cookie, err := c.Request.Cookie(GATEWAY_PORTAL_COOKIE_NAME); err == nil {
		sessionID = cookie.Value
	}
	accountID, err := verifySession(sessionID, h)
	if err != nil {
		logrus.WithError(err).WithField("sessionID", sessionID).
			Info("(UpdateGatewayAccountVersion)) verifying session produced error")
		c.String(401, "")
		return
	}

	gasr := search.GatewayAccountSearchRequest{}
	gasr.AddFilter(search.GatewayAccountSearch_Id, accountID, searcher.Operator_EQ, searcher.Condition_AND)
	gasr.Limit = 1
	accounts, err := h.GatewayAccountSearcher.Search(gasr)
	// Not 100% on how to handle this error, for now just redirecting to /gateway/signin
	if err != nil || len(accounts) == 0 {
		var errorS string
		if err != nil {
			errorS = err.Error()
		} else {
			errorS = "no gateway account found"
		}
		logrus.WithFields(logrus.Fields{
			"error":     errorS,
			"sessionID": sessionID,
			"accountID": accountID,
		}).Error("(UpdateGatewayAccountVersion)) search for gateway account produced error")
		c.String(401, "")
		return
	}
	account := accounts[0]

	var req updateGatewayVersionRequest
	if err := c.Bind(&req); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"sessionID": sessionID,
			"accountID": accountID,
		}).Error("(UpdateGatewayAccountVersion) binding request produced error")
		c.JSON(400, gin.H{
			"error": "Improper request, please try again.",
		})
		return
	}

	gaur := update.GatewayAccountUpdateRequest{
		Id:      zeus.Uuid(accountID),
		Updates: map[update.GatewayAccountUpdateField]interface{}{},
	}
	gaur.Updates[update.GatewayAccountUpdate_ActiveVersion] = req.ActiveVersion
	if err = h.GatewayAccountUpdater.Update(gaur); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"sessionID": sessionID,
			"accountID": account.Id,
		}).Error("(UpdateGatewayAccountVersion)) updating gateway account produced error")
		c.JSON(400, gin.H{
			"error": "An unexpected error occurred, please try again",
		})
		return
	}
	c.String(204, "")

}

func (h *Handlers) UpdateGatewayAccountPassword(c *gin.Context, dc desmond.Context) {
	// Verify session on request is valid
	var sessionID string
	if cookie, err := c.Request.Cookie(GATEWAY_PORTAL_COOKIE_NAME); err == nil {
		sessionID = cookie.Value
	}
	accountID, err := verifySession(sessionID, h)
	if err != nil {
		logrus.WithError(err).WithField("sessionID", sessionID).
			Info("(UpdateGatewayAccountPassword) verifying session produced error")
		c.String(401, "")
		return
	}

	gasr := search.GatewayAccountSearchRequest{}
	gasr.AddFilter(search.GatewayAccountSearch_Id, accountID, searcher.Operator_EQ, searcher.Condition_AND)
	gasr.Limit = 1
	accounts, err := h.GatewayAccountSearcher.Search(gasr)
	// Not 100% on how to handle this error, for now just redirecting to /gateway/signin
	if err != nil || len(accounts) == 0 {
		var errorS string
		if err != nil {
			errorS = err.Error()
		} else {
			errorS = "no gateway account found"
		}
		logrus.WithFields(logrus.Fields{
			"error":     errorS,
			"sessionID": sessionID,
			"accountID": accountID,
		}).Error("(UpdateGatewayAccountPassword) search for gateway account produced error")
		c.String(401, "")
		return
	}
	account := accounts[0]

	var req updateGatewayAccountPasswordRequest
	if err := c.Bind(&req); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"sessionID": sessionID,
			"accountID": accountID,
		}).Error("(UpdateGatewayAccountPassword) binding request produced error")
		c.JSON(400, gin.H{
			"error": "Improper request, please try again.",
		})
		return
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(req.OldPassword)); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"sessionID": sessionID,
			"accountID": accountID,
		}).Error("(UpdateGatewayAccountPassword) password match failed")
		c.JSON(400, gin.H{
			"error": "You cannot update password without correctly providing the existing password",
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), HASH_COST)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":     err.Error(),
			"accountID": account.Id,
		}).Error("(UpdateGatewayAccountPassword) hashing password produced error")
		c.JSON(400, gin.H{
			"error": "An unexpected error occurred, please try again.",
		})
		return
	}

	gaur := update.GatewayAccountUpdateRequest{
		Id:      zeus.Uuid(accountID),
		Updates: map[update.GatewayAccountUpdateField]interface{}{},
	}
	gaur.Updates[update.GatewayAccountUpdate_PasswordHash] = string(hashedPassword)
	if err = h.GatewayAccountUpdater.Update(gaur); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"sessionID": sessionID,
			"accountID": account.Id,
		}).Error("(UpgateGatewayAccountPassword) updating gateway account produced error")
		c.JSON(400, gin.H{
			"error": "An unexpected error occurred, please try again",
		})
		return
	}
	c.String(204, "")

}

func (h *Handlers) RefreshGatewayCredentials(c *gin.Context, dc desmond.Context) {
	// Verify session on request is valid
	var sessionID string
	if cookie, err := c.Request.Cookie(GATEWAY_PORTAL_COOKIE_NAME); err == nil {
		sessionID = cookie.Value
	}
	accountID, err := verifySession(sessionID, h)
	if err != nil {
		logrus.WithError(err).WithField("sessionID", sessionID).
			Info("(RefreshGatewayCredentials) verifying session produced error")
		c.String(401, "")
		return
	}

	// Refresh keys
	gatewayKey, gatewaySecret := generateGatewayKeyPair()

	gaur := update.GatewayAccountUpdateRequest{
		Id:      zeus.Uuid(accountID),
		Updates: map[update.GatewayAccountUpdateField]interface{}{},
	}
	gaur.Updates[update.GatewayAccountUpdate_GatewayKey] = gatewayKey
	gaur.Updates[update.GatewayAccountUpdate_GatewaySecret] = gatewaySecret
	if err = h.GatewayAccountUpdater.Update(gaur); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"sessionID": sessionID,
			"accountID": accountID,
		}).Error("(UpgateGatewayAccount) updating gateway account produced error")
		c.JSON(400, gin.H{
			"error": "An unexpected error occurred, please try again",
		})
		return
	}

	c.JSON(200, gin.H{
		"gatewayKey":    gatewayKey,
		"gatewaySecret": gatewaySecret,
	})
}
