package gateway

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/getbread/breadkit/desmond"
	"github.com/getbread/shopify_plugin_backend/service/gateway/security"
	"github.com/getbread/shopify_plugin_backend/service/types"
	"github.com/getbread/shopify_plugin_backend/service/bread"
	"github.com/getbread/shopify_plugin_backend/service/cache"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handlers) GatewayOrderManagement(c *gin.Context, dc desmond.Context) {
	// Unmarshal request and find gateway account
	req, account, err := processGatewayOrderManagementRequest(c, h)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":       err.Error(),
			"account":     account,
			"requestBody": c.Request.Body,
		}).Error("(GatewayOrderManagement) processing gateway order management request produced error")
		c.String(400, "INVALID_REQUEST")
	}

	// Acknowledge request and return HTTP 200 early
	c.String(200, "OK")

	// Create request signature map
	signatureCheck := createRequestSignatureMap(req)

	// Verify authenticity
	if !security.GatewayRequestAuthentic(signatureCheck, account.GatewaySecret, req.Signature) {
		logrus.WithFields(logrus.Fields{
			"requestMap": req,
		}).Error("(GatewayOrderManagement) request signature invalid")
		go postFailedResponse(req, account, "Could not verify request authenticity - please contact Bread integrations for support")
		return
	}

	// Query for Gateway Checkout
	checkout, err := findGatewayCheckoutByTxId(req.GatewayReference, h)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"checkout": checkout,
		}).Error("(GatewayOrderManagement) query for gateway checkout produced error")
		go postFailedResponse(req, account, "Could not find requested transaction - please contact Bread integrations for support")
		return
	}

	breadVersion := checkout.BreadVersion
	if breadVersion == "" {
		breadVersion = BreadClassic
	}
	cache := cache.NewCache(h.RedisPool)

	if req.TransactionType == "capture" {
		// Hotfix for Shopify Autocapture setting, reject capture request if it happens within 10 seconds
		waitFor := gatewayConfig.MiltonGatewayAutoSettleTimeout
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error":                err.Error(),
				"autoSettleTimeoutVar": gatewayConfig.MiltonGatewayAutoSettleTimeout,
			}).Error("(GatewayOrderManagement) error converting auto-settle timeout variable to integer")
		}
		if waitFor == 0 {
			waitFor = 10
		}
		captureTimeGap := time.Now().Sub(checkout.UpdatedAt)

		autoSettle := account.AutoSettle
		if breadVersion == BreadPlatform {
			autoSettle = account.PlatformAutoSettle
		}

		if captureTimeGap < time.Duration(waitFor)*time.Second && !autoSettle {
			go postFailedResponse(req, account, "Automatic capture request was ignored because auto-settle is disabled - log into shopify.getbread.com for more information")
		} else {
			if breadVersion == BreadClassic {
				go attemptCapture(req, account, checkout)
			} else {
				go attemptPlatformCapture(req, account, checkout, cache)
			}

		}
	} else if req.TransactionType == "void" {
		if breadVersion == BreadClassic {
			go attemptCancel(req, account, checkout)
		} else {
			go attemptPlatformCancel(req, account, checkout, cache)
		}

	} else if req.TransactionType == "refund" {
		if breadVersion == BreadClassic {
			go attemptRefund(req, account, checkout)
		} else {
			go attemptPlatformRefund(req, account, checkout, cache)
		}

	} else {
		logrus.WithFields(logrus.Fields{
			"request":         req,
			"account":         account,
			"gatewayCheckout": checkout,
		}).Error("(GatewayOrderManagement) Unrecognized gateway order management request")
		go postFailedResponse(req, account, "Unrecognized request")
	}
	return
}

func attemptCapture(req gatewayOrderManagementRequest, account types.GatewayAccount, checkout types.GatewayCheckout) {
	bt, err := queryTransaction(req.GatewayReference, account, checkout)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"account":       account,
			"transactionId": req.GatewayReference,
		}).Error("(GatewayOrderManagement) querying transaction produced error")
		postFailedResponse(req, account, "Could not find Bread transaction")
		return
	}
	// If Bread transaction is already settled, POST successful response
	if bt.Status == "SETTLED" {
		postCompletedResponse(req, account, "Payment was already captured successfully")
		return
	}
	if bt.Status == "PENDING" {
		// Attempt to authorize transaction
		if err := authorizeTransaction(req.GatewayReference, account, checkout); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"account":       account,
				"transactionId": req.GatewayReference,
			}).Error("(GatewayOrderManagement) authorizing transaction produced error")
			postFailedResponse(req, account, "Attempt to authorize before capture failed")
			return
		}
	}

	// Format Bread tx and Shopify capture amounts for comparison
	btAmount := float64(bt.AdjustedTotal) / 100.00
	scAmount, err := strconv.ParseFloat(req.Amount, 64)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"account":                account,
			"transactionId":          req.GatewayReference,
			"breadTransactionAmount": bt.AdjustedTotal,
			"shopifyCaptureAmount":   req.Amount,
		}).Error("(GatewayOrderManagement) string converting Shopify capture amount produced error")
		postFailedResponse(req, account, "Capture request failed")
		return
	}

	if btAmount == scAmount {
		// Settle Bread transaction
		if err := settleTransaction(req.GatewayReference, account, checkout); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"account":       account,
				"transactionId": req.GatewayReference,
			}).Error("(GatewayOrderManagement) settling transaction produced error")
			postFailedResponse(req, account, "Capture request failed")
			return
		}
	} else {
		// Throw error if capture amount > transaction amount
		if scAmount > btAmount {
			logrus.WithFields(logrus.Fields{
				"error":         "capture request amount > transaction amount",
				"account":       account,
				"request":       req,
				"transactionId": req.GatewayReference,
			}).Error("(GatewayOrderManagement) capture request amount > transaction amount")
			postFailedResponse(req, account, "Capture amount is greater than Bread transaction total")
			return
		}
		// Partial cancel, then full settle
		cancelAmountCentsFloat := (btAmount - scAmount) * 100.00
		cancelAmountCents := int(cancelAmountCentsFloat)
		// Cancel Bread transaction
		if err := cancelTransaction(req.GatewayReference, account, checkout, cancelAmountCents); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"account":       account,
				"transactionId": req.GatewayReference,
			}).Error("(GatewayOrderManagement) partial cancel before partial settle produced error")
			postFailedResponse(req, account, "Capture request failed")
			return
		}
		if err := settleTransaction(req.GatewayReference, account, checkout); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"account":       account,
				"transactionId": req.GatewayReference,
			}).Error("(GatewayOrderManagement) partial settle after partial cancel produced error")
			postFailedResponse(req, account, "Attempt to settle transaction failed")
			return
		}
	}
	postCompletedResponse(req, account, "Payment captured successfully")
	return
}

func attemptCancel(req gatewayOrderManagementRequest, account types.GatewayAccount, checkout types.GatewayCheckout) {
	// Cancel Bread transaction
	// Shopify doesn't support partial cancels, pass 0 for `amount` to perform a full cancel request
	if err := cancelTransaction(req.GatewayReference, account, checkout, 0); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"account":       account,
			"transactionId": req.GatewayReference,
		}).Error("(GatewayOrderManagement) canceling transaction produced error")
		postFailedResponse(req, account, "Cancel request failed")
		return
	}
	postCompletedResponse(req, account, "Payment canceled successfully")
	return
}

func attemptRefund(req gatewayOrderManagementRequest, account types.GatewayAccount, checkout types.GatewayCheckout) {
	bt, err := queryTransaction(req.GatewayReference, account, checkout)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"account":       account,
			"transactionId": req.GatewayReference,
		}).Error("(GatewayOrderManagement) querying transaction produced error")
		postFailedResponse(req, account, "Could not find Bread transaction")
		return
	}
	// If Bread transaction is already refunded, POST successful response
	if bt.Status == "REFUNDED" {
		postCompletedResponse(req, account, "Payment was already refunded successfully")
		return
	}

	// Format refund amount to int
	refundAmount := strings.Replace(req.Amount, ".", "", 1)
	refundAmountInt, err := strconv.Atoi(refundAmount)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"account":       account,
			"transactionId": req.GatewayReference,
			"refundAmount":  req.Amount,
		}).Error("(GatewayOrderRefund) converting refund amount to int produced error")
		postFailedResponse(req, account, "Refund request failed")
		return
	}

	// Refund Bread transaction
	if err := refundTransaction(req.GatewayReference, account, checkout, refundAmountInt); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"account":       account,
			"transactionId": req.GatewayReference,
		}).Error("(GatewayOrderRefund) refunding transaction produced error")
		postFailedResponse(req, account, "Refund request failed")
		return
	}
	postCompletedResponse(req, account, "Payment refunded successfully")
	return
}

func postCompletedResponse(req gatewayOrderManagementRequest, account types.GatewayAccount, msg string) {
	if err := postGatewayOrderResponse(req, account, "completed", msg); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"requestType": "completed",
			"message":     msg,
			"request":     req,
			"account":     account,
		}).Error("(GatewayOrderManagement) making successful HTTP callback request produced error")
	}
	return
}

func postFailedResponse(req gatewayOrderManagementRequest, account types.GatewayAccount, msg string) {
	if err := postGatewayOrderResponse(req, account, "failed", msg); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"requestType": "failed",
			"message":     msg,
			"request":     req,
			"account":     account,
		}).Error("(GatewayOrderManagement) making failed HTTP callback request produced error")
	}
	return
}

func attemptPlatformCapture(
	req gatewayOrderManagementRequest,
	account types.GatewayAccount,
	checkout types.GatewayCheckout,
	cache cache.Cache) {

	transactionID := req.GatewayReference
	apiKey, apiSecret, host := platformApiParams(checkout.Test, account)
	trxResponse, httpError := platformGetTransaction(host, apiKey, apiSecret, transactionID, cache)
	if httpError != nil {
		logrus.WithFields(logrus.Fields{
			"account":       account,
			"checkout":      checkout,
			"transactionId": transactionID,
		}).Error("(GatewayOrderManagement) Transaction not found")
		postFailedResponse(req, account, fmt.Sprintf("Transaction not found: %s", transactionID))
		return
	}

	if strings.ToLower(trxResponse.Status) == "settled" {
		postCompletedResponse(req, account, "Payment was already captured successfully")
		return
	}

	transactionAmount, err := types.USDToCents(checkout.AmountStr)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"transactionID":     transactionID,
			"transactionAmount": checkout.AmountStr,
		}).Error("(GatewayOrderManagement) Invalid transaction amount encountered when attempting capture")

		postFailedResponse(req, account, fmt.Sprintf("Invalid transaction amount: %s", checkout.AmountStr))
		return
	}

	captureAmount, err := types.USDToCents(req.Amount)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"transactionID": transactionID,
			"captureAmount": req.Amount,
		}).Error("(GatewayOrderManagement) Invalid capture amount encountered when attempting capture")

		postFailedResponse(req, account, fmt.Sprintf("Invalid capture amount: %s", req.Amount))
		return
	}

	amount := bread.TrxAmount{Currency: checkout.Currency, Value: transactionAmount}
	trxReq := bread.TrxRequest{Amount: amount, MerchantOfRecordID: checkout.MerchantId}

	// Attempt to authorize transaction
	if strings.ToLower(trxResponse.Status) == "pending" {
		httpError = platformAuthorizeTransaction(host, apiKey, apiSecret, transactionID, trxReq, cache)
		if httpError != nil {
			logrus.WithFields(logrus.Fields{
				"account":       account,
				"checkout":      checkout,
				"transactionId": transactionID,
			}).Error("(GatewayOrderManagement) Attempt to authorize transaction resulted in an error")
			postFailedResponse(req, account, "Attempt to authorize transaction resulted in an error")
			return
		}
	}

	if transactionAmount < captureAmount {
		logrus.WithError(err).WithFields(logrus.Fields{
			"transactionID":     transactionID,
			"captureAmount":     req.Amount,
			"transactionAmount": transactionAmount,
		}).Error("(GatewayOrderManagement) Bread transaction amount is less than capture amount")

		postFailedResponse(req, account, "Total transaction amount is less than capture amount.")
		return
	} else if transactionAmount > captureAmount {
		// Process partial cancel. Cancel the excess on the capture amount
		cancelAmount := bread.TrxAmount{Currency: checkout.Currency, Value: transactionAmount - captureAmount}
		cancelReq := bread.TrxRequest{Amount: cancelAmount, MerchantOfRecordID: checkout.MerchantId}

		httpError = platformCancelTransaction(host, apiKey, apiSecret, transactionID, cancelReq, cache)
		if httpError != nil {
			logrus.WithFields(logrus.Fields{
				"account":       account,
				"transactionId": transactionID,
			}).Error("(GatewayOrderManagement) Partial cancel preceeding partial settling failed")
			postFailedResponse(req, account, "Attempt to capture transaction failed")
			return
		}

		//Process partial settle. Settle the capture amount
		settleAmount := bread.TrxAmount{Currency: checkout.Currency, Value: captureAmount}
		settleReq := bread.TrxRequest{Amount: settleAmount, MerchantOfRecordID: checkout.MerchantId}

		httpError = platformSettleTransaction(host, apiKey, apiSecret, transactionID, settleReq, cache)
		if httpError != nil {
			logrus.WithFields(logrus.Fields{
				"account":       account,
				"checkout":      checkout,
				"transactionId": transactionID,
			}).Error("(GatewayOrderManagement) Partial settling failed")
			postFailedResponse(req, account, "Attempt to settle transaction failed")
			return
		}
	} else {
		httpError = platformSettleTransaction(host, apiKey, apiSecret, transactionID, trxReq, cache)
		if httpError != nil {
			logrus.WithFields(logrus.Fields{
				"account":       account,
				"checkout":      checkout,
				"transactionId": transactionID,
			}).Error("(GatewayOrderManagement) Attempt to settle transaction resulted in an error")
			postFailedResponse(req, account, "Attempt to settle transaction resulted in an error")
			return
		}
	}

	postCompletedResponse(req, account, "Payment successfully captured")
}

func attemptPlatformCancel(
	req gatewayOrderManagementRequest,
	account types.GatewayAccount,
	checkout types.GatewayCheckout,
	cache cache.Cache) {

	transactionID := req.GatewayReference
	apiKey, apiSecret, host := platformApiParams(checkout.Test, account)
	trxResponse, httpError := platformGetTransaction(host, apiKey, apiSecret, transactionID, cache)
	if httpError != nil {
		logrus.WithFields(logrus.Fields{
			"account":       account,
			"checkout":      checkout,
			"transactionId": transactionID,
		}).Error("(GatewayOrderManagement) Transaction not found")
		postFailedResponse(req, account, fmt.Sprintf("Transaction not found: %s", transactionID))
		return
	}

	if strings.ToLower(trxResponse.Status) == "cancelled" {
		postCompletedResponse(req, account, "Payment was already cancelled")
		return
	}

	transactionAmount, err := types.USDToCents(checkout.AmountStr)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"transactionID":     transactionID,
			"transactionAmount": checkout.AmountStr,
		}).Error("(GatewayOrderManagement) Invalid transaction amount encountered when attempting cancellation")

		postFailedResponse(req, account, fmt.Sprintf("Invalid transaction amount: %s", checkout.AmountStr))
		return
	}

	amount := bread.TrxAmount{Currency: checkout.Currency, Value: transactionAmount}
	trxReq := bread.TrxRequest{Amount: amount, MerchantOfRecordID: checkout.MerchantId}

	httpError = platformCancelTransaction(host, apiKey, apiSecret, transactionID, trxReq, cache)
	if httpError != nil {
		logrus.WithFields(logrus.Fields{
			"account":       account,
			"transactionId": transactionID,
		}).Error("(GatewayOrderManagement) Attempt to cancel transaction resulted in an error")
		postFailedResponse(req, account, "Attempt to cancel transaction resulted in an error")
		return
	}

	postCompletedResponse(req, account, "Payment successfully cancelled")
}

func attemptPlatformRefund(
	req gatewayOrderManagementRequest,
	account types.GatewayAccount,
	checkout types.GatewayCheckout,
	cache cache.Cache) {

	transactionID := req.GatewayReference
	apiKey, apiSecret, host := platformApiParams(checkout.Test, account)
	trxResponse, httpError := platformGetTransaction(host, apiKey, apiSecret, transactionID, cache)
	if httpError != nil {
		logrus.WithFields(logrus.Fields{
			"account":       account,
			"checkout":      checkout,
			"transactionId": transactionID,
		}).Error("(GatewayOrderManagement) Transaction not found")
		postFailedResponse(req, account, fmt.Sprintf("Transaction not found: %s", transactionID))
		return
	}

	if strings.ToLower(trxResponse.Status) == "refunded" {
		postCompletedResponse(req, account, "Payment was already refunded")
		return
	}

	transactionAmount, err := types.USDToCents(checkout.AmountStr)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"transactionID":     transactionID,
			"transactionAmount": checkout.AmountStr,
		}).Error("(GatewayOrderManagement) Invalid transaction amount encountered when attempting refund")

		postFailedResponse(req, account, fmt.Sprintf("Invalid transaction amount: %s", checkout.AmountStr))
		return
	}

	refundAmount, err := types.USDToCents(req.Amount)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"transactionID": transactionID,
			"refundAmount":  req.Amount,
		}).Error("(GatewayOrderManagement) Amount to refund is invalid")

		postFailedResponse(req, account, fmt.Sprintf("Invalid refund amount: %s", req.Amount))
		return
	}

	if transactionAmount < refundAmount {
		logrus.WithError(err).WithFields(logrus.Fields{
			"transactionID":     transactionID,
			"refundAmount":      req.Amount,
			"transactionAmount": transactionAmount,
		}).Error("(GatewayOrderManagement) Refund amount is greater than transaction total")

		postFailedResponse(req, account, "Amount to refund is greater than the total transaction amount.")
		return
	}

	amount := bread.TrxAmount{Currency: checkout.Currency, Value: refundAmount}
	trxReq := bread.TrxRequest{Amount: amount, MerchantOfRecordID: checkout.MerchantId}

	httpError = platformRefundTransaction(host, apiKey, apiSecret, transactionID, trxReq, cache)
	if httpError != nil {
		logrus.WithFields(logrus.Fields{
			"account":       account,
			"transactionId": transactionID,
		}).Error("(GatewayOrderManagement)  Attempt to refund transaction resulted in an error")
		postFailedResponse(req, account, "Attempt to refund transaction resulted in an error")
		return
	}

	postCompletedResponse(req, account, "Payment successfully refunded")
}
