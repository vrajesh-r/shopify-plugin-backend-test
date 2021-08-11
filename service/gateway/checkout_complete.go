package gateway

import (
	"net/url"
	"strconv"
	"time"

	"github.com/getbread/breadkit/desmond"
	zeus "github.com/getbread/breadkit/zeus/types"
	"github.com/getbread/shopify_plugin_backend/service/shopify"
	"github.com/getbread/shopify_plugin_backend/service/update"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// IT 8/2/2019: Deprecated handler for checkout callbackURLs. With callbackURLs no longer used in future cart create requests,
// this handler will not be hit by new checkouts. Leaving intact for now to accommodate old outstanding Bread carts.

func (h *Handlers) GatewayCheckoutComplete(c *gin.Context, dc desmond.Context) {
	// Log the transaction callback request for diagnosing offsite checkout pending transactions
	logrus.Info("(TxTracker) callback request start")
	var req gatewayCheckoutCallbackRequest
	if err := c.Bind(&req); err != nil {
		logrus.WithError(err).Error("(GatewayCheckoutComplete) binding request to struct produced error")
		c.String(400, err.Error())
		return
	}

	// Log the transaction callback request for diagnosing offsite checkout pending transactions
	logrus.Infof("(TxTracker)[%s] callback request bind transaction id", req.TransactionId)

	checkout, err := findGatewayCheckoutById(zeus.Uuid(req.OrderRef), h)
	if err != nil {
		logrus.WithError(err).WithField("request", req).Error("(GatewayCheckoutComplete) search for offsite checkout produced error")
		c.String(400, err.Error())
		return
	}
	account, err := findGatewayAccountById(checkout.AccountID, h)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"request":  req,
			"checkout": checkout,
		}).Error("(GatewayCheckoutComplete) query for gateway account produced error")
		c.String(400, err.Error())
		return
	}

	// Short circuit
	if checkout.Completed {
		logrus.Infof("(TxTracker)[%s] callback request short circuit", req.TransactionId)
		c.String(200, "OP_COMPLETE")
		return
	}

	if err = authorizeTransaction(req.TransactionId, account, checkout); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"request": req,
			"account": account,
		}).Error("(GatewayCheckoutComplete) authorizing transaction produced error")
		c.String(400, err.Error())
		return
	}
	logrus.Infof("(TxTracker)[%s] callback request authorize", req.TransactionId)

	// auto_settle the transaction if needed
	if account.AutoSettle {
		if err := settleTransaction(req.TransactionId, account, checkout); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"request":  req,
				"checkout": checkout,
			}).Error("(GatewayCheckoutComplete) settling transaction produced error")
			c.String(400, err.Error())
			return
		}
		logrus.Infof("(TxTracker)[%s] callback request settle", req.TransactionId)
	}

	// Mark checkout as complete on Shopify
	response := &shopify.GatewayCheckoutCompleteRequest{
		AccountId:        account.GatewayKey,
		Reference:        checkout.Reference,
		Currency:         checkout.Currency,
		Test:             checkout.Test,
		Amount:           checkout.Amount,
		GatewayReference: req.TransactionId,
		Timestamp:        time.Now().UTC().Format(time.RFC3339),
		Result:           shopify.ResultComplete,
		TransactionType:  shopify.TxTypeAuthorization,
	}

	// Ensure the OMS shows the order payment as pending
	// letting the merchant employees know they should go to
	// the merchants portal and settle
	if account.AutoSettle {
		response.TransactionType = shopify.TxTypeSale
	}

	signGatewayCheckoutResponse(response, account.GatewaySecret)

	form := url.Values{}
	form.Set("x_account_id", response.AccountId)
	form.Set("x_reference", response.Reference)
	form.Set("x_currency", response.Currency)
	form.Set("x_test", strconv.FormatBool(response.Test))
	form.Set("x_amount", strconv.FormatFloat(response.Amount, 'f', 2, 64))
	form.Set("x_gateway_reference", response.GatewayReference)
	form.Set("x_result", string(response.Result))
	form.Set("x_transaction_type", string(response.TransactionType))
	form.Set("x_timestamp", response.Timestamp)
	form.Set("x_signature", response.Signature)

	if err := HTTPFormRequest("POST", checkout.CallbackUrl, form, struct{}{}); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"request":  req,
			"checkout": checkout,
			"form":     form,
		}).Error("(GatewayCheckoutComplete) making HTTP complete request produced error")
		c.String(400, err.Error())
		return
	}
	logrus.Infof("(TxTracker)[%s] callback request shopify post", req.TransactionId)

	// Mark our order record as completed
	gcur := update.GatewayCheckoutUpdateRequest{
		Id:      checkout.Id,
		Updates: map[update.GatewayCheckoutUpdateField]interface{}{},
	}
	gcur.Updates[update.GatewayCheckoutUpdate_TransactionID] = req.TransactionId
	gcur.Updates[update.GatewayCheckoutUpdate_Completed] = true
	if err := h.GatewayCheckoutUpdater.Update(gcur); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"request":       req,
			"checkout":      checkout,
			"transactionId": req.TransactionId,
		}).Error("(GatewayCheckoutComplete) update to gateway checkout internal record as completed failed")
		// Continue since order has been copied to the merchants OMS
	}
	logrus.Infof("(TxTracker)[%s] callback request successful", req.TransactionId)

	c.String(200, "OP_COMPLETE")
}
