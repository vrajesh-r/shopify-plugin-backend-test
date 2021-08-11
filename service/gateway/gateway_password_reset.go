package gateway

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"time"

	"github.com/getbread/breadkit/desmond"
	"github.com/getbread/breadkit/zeus/searcher"
	zeus "github.com/getbread/breadkit/zeus/types"
	"github.com/getbread/shopify_plugin_backend/service/mailer"
	"github.com/getbread/shopify_plugin_backend/service/search"
	"github.com/getbread/shopify_plugin_backend/service/types"
	"github.com/getbread/shopify_plugin_backend/service/update"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handlers) ForgotGatewayAccountPassword(c *gin.Context, dc desmond.Context) {
	var req forgotPasswordRequest
	err := c.BindJSON(&req)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("(ForgotGatewayAccountPassword) binding request to model produced an error")
		c.JSON(500, gin.H{})
		return
	}
	c.JSON(202, gin.H{})

	account, err := findGatewayAccountByEmail(req.Email, h)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":        err.Error(),
			"accountEmail": req.Email,
		}).Warn("(ForgotGatewayAccountPassword) account with requested email not found")
		return
	}

	token, err := generateResetToken(account.Id, h)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":     err.Error(),
			"accountID": account.Id,
		}).Error("(ForgotGatewayAccountPassword) generating password reset token produced an error")
		return
	}

	resetLink := gatewayConfig.HostConfig.MiltonHost + "/gateway/#reset?token=" + token
	err = mailer.SendPasswordResetLink(account.Email, resetLink)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":     err.Error(),
			"accountID": account.Id,
		}).Error("(ForgotGatewayAccountPassword) sending password reset email produced an error")
		return
	}
	return
}

func (h *Handlers) ResetGatewayAccountPassword(c *gin.Context, dc desmond.Context) {
	var req resetPasswordRequest
	err := c.BindJSON(&req)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("(ResetGatewayAccountPassword) binding request to model produced an error")
		c.JSON(500, gin.H{})
		return
	}

	if req.NewPassword != req.NewPasswordVerify {
		logrus.Warn("(ResetGatewayAccountPassword) new passwords did not match")
		c.JSON(400, gin.H{
			"error": "Passwords do not match",
		})
		return
	}

	t, err := findResetToken(req.ResetToken, h)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Warn("(ResetGatewayAccountPassword) invalid reset token")
		c.JSON(400, gin.H{})
		return
	}

	account, err := findGatewayAccountById(t.AccountID, h)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), HASH_COST)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":     err.Error(),
			"accountID": account.Id,
		}).Error("(ResetGatewayAccountPassword) hashing password produced error")
		c.JSON(500, gin.H{})
		return
	}

	gaur := update.GatewayAccountUpdateRequest{
		Id:      zeus.Uuid(account.Id),
		Updates: map[update.GatewayAccountUpdateField]interface{}{},
	}
	gaur.Updates[update.GatewayAccountUpdate_PasswordHash] = string(hashedPassword)
	err = h.GatewayAccountUpdater.Update(gaur)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":     err.Error(),
			"accountID": account.Id,
		}).Error("(ResetGatewayAccountPassword) updating gateway account produced error")
		c.JSON(500, gin.H{})
		return
	}

	err = expireResetToken(t.ID, h)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"error":     err.Error(),
			"accountID": account.Id,
		}).Error("(ResetGatewayAccountPassword) expiring password reset token produced error")
		c.JSON(500, gin.H{})
	}
	c.JSON(202, gin.H{})
	return
}

func (h *Handlers) ValidateResetToken(c *gin.Context, dc desmond.Context) {
	var req resetPasswordRequest
	err := c.BindJSON(&req)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("(ValidateResetToken) binding request to model produced an error")
		c.JSON(500, gin.H{})
		return
	}

	_, err = findResetToken(req.ResetToken, h)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Warn("(ValidateResetToken) invalid reset token")
		c.JSON(400, gin.H{})
		return
	}
	c.JSON(202, gin.H{
		"authorized": true,
	})
	return
}

func generateResetToken(accountID zeus.Uuid, h *Handlers) (string, error) {
	tp, err := generateTokenPair()
	if err != nil {
		return "", err
	}
	expiration := types.GenerateResetTokenExpiration()
	resetRequest := types.GatewayPasswordResetRequest{
		AccountID:  accountID,
		TokenHash:  tp.TokenHash,
		Expiration: expiration,
	}
	_, err = h.GatewayPasswordResetRequestCreator.Create(resetRequest)
	if err != nil {
		return "", err
	}
	return tp.Token, nil
}

func findResetToken(t string, h *Handlers) (token types.GatewayPasswordResetRequest, err error) {
	searchRequest := search.GatewayPasswordResetRequestSearchRequest{}
	searchRequest.AddFilter(search.GatewayPasswordResetRequestSearch_Expiration, time.Now().Unix(), searcher.Operator_GTE, searcher.Condition_AND)
	tokens, err := h.GatewayPasswordResetRequestSearcher.Search(searchRequest)
	if err != nil {
		return
	}
	if len(tokens) == 0 {
		err = fmt.Errorf("no tokens found")
		return
	}
	foundToken := false
	foundTokenIndex := 0
	for i := range tokens {
		err = bcrypt.CompareHashAndPassword([]byte(tokens[i].TokenHash), []byte(t))
		if err == nil {
			foundToken = true
			foundTokenIndex = i
			break
		}
	}
	if !foundToken {
		return
	}
	token = tokens[foundTokenIndex]
	return
}

func expireResetToken(tokenID zeus.Uuid, h *Handlers) (err error) {
	searchRequest := search.GatewayPasswordResetRequestSearchRequest{}
	searchRequest.AddFilter(search.GatewayPasswordResetRequestSearch_ID, tokenID, searcher.Operator_EQ, searcher.Condition_AND)
	searchRequest.Limit = 1
	tokens, err := h.GatewayPasswordResetRequestSearcher.Search(searchRequest)
	if err != nil {
		return
	}
	if len(tokens) == 0 {
		err = fmt.Errorf("no tokens found")
		return
	}
	token := tokens[0]
	updateRequest := update.GatewayPasswordResetRequestUpdateRequest{
		Id:      zeus.Uuid(token.ID),
		Updates: map[update.GatewayPasswordResetRequestUpdateField]interface{}{},
	}
	updateRequest.Updates[update.GatewayPasswordResetRequestUpdate_Expiration] = time.Now().AddDate(0, 0, -1).Unix()
	err = h.GatewayPasswordResetRequestUpdater.Update(updateRequest)
	return
}

type TokenPair struct {
	Token     string
	TokenHash string
}

func generateTokenPair() (t TokenPair, err error) {
	t.Token, err = random(256)
	if err != nil {
		return
	}
	tokenHashBytes, err := bcrypt.GenerateFromPassword([]byte(t.Token), HASH_COST)
	if err != nil {
		return
	}
	t.TokenHash = string(tokenHashBytes)
	return
}

func random(bits int) (string, error) {
	result := make([]byte, bits/8)
	_, err := io.ReadFull(rand.Reader, result)
	if err != nil {
		return "", fmt.Errorf("Error generating random values: %v", err)
	}
	return base64.RawURLEncoding.EncodeToString(result), nil
}
