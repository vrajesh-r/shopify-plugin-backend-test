package gateway

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/getbread/breadkit/desmond"
	zeus "github.com/getbread/breadkit/zeus/types"
	"github.com/getbread/shopify_plugin_backend/service/shopify"
	"github.com/getbread/shopify_plugin_backend/service/bread"
	"github.com/getbread/shopify_plugin_backend/service/cache"
	"github.com/getbread/shopify_plugin_backend/service/types"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handlers) GatewayCheckoutConfirmation(c *gin.Context, dc desmond.Context) {
	// Log the transaction complete request for diagnosing offsite checkout pending transactions
	logrus.Info("(TxTracker) complete request start")
	// Pull id from params
	gatewayCheckoutID := c.Query("orderRef")
	breadTransactionID := c.Query("transactionId")

	// Log the transaction complete request for diagnosing offsite checkout pending transactions
	logrus.Infof("(TxTracker)[%s] complete request bind transaction id", breadTransactionID)

	// Query for offsite checkout by reference
	checkout, err := findGatewayCheckoutById(zeus.Uuid(gatewayCheckoutID), h)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"gatewayCheckoutId":  gatewayCheckoutID,
			"breadTransactionId": breadTransactionID,
		}).Error("(GatewayCheckoutConfirmation) search for gateway checkout produced error")
		c.String(400, err.Error())
		return
	}

	// Query shop
	account, err := findGatewayAccountById(checkout.AccountID, h)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"gatewayCheckoutID":  gatewayCheckoutID,
			"breadTransactionID": breadTransactionID,
			"checkout":           checkout,
		}).Error("(GatewayCheckoutConfirmation) error querying for shop by id")
		c.Redirect(302, checkout.CancelUrl)
		return
	}

	logrus.Infof("(TxTracker)[%s] complete request processing checkout", breadTransactionID)

	// Keep track of authorization success state
	authSuccess := true
	remainderPayDecline := false

	err = authorizeTransaction(breadTransactionID, account, checkout)
	if err != nil {
		err := authorizeTransaction(breadTransactionID, account, checkout)
		if err != nil {
			authSuccess = false
			remainderPayDecline = strings.Contains(err.Error(), "There's an issue with authorizing the credit card portion")

			logrus.WithError(err).WithFields(logrus.Fields{
				"account":           account,
				"gatewayCheckoutId": gatewayCheckoutID,
				"transactionId":     breadTransactionID,
			}).Error("(GatewayCheckoutConfirmation) authorizing transaction produced error")

			if remainderPayDecline && account.RemainderPayAutoCancel {
				err := cancelTransaction(breadTransactionID, account, checkout, 0)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"error":         err.Error(),
						"transactionID": breadTransactionID,
						"account":       account,
					}).Error("(GatewayCheckoutConfirmation) canceling remainder-pay declined transaction failed")
				}
			}

			if remainderPayDecline {
				c.HTML(400, "checkout_error.html", gin.H{
					"cancel":           checkout.CancelUrl,
					"messagePrimary":   remainderPayErrMessage,
					"messageSecondary": remainderPayErrMessageAction,
				})
				return
			}

			// Short circuit flow and redirect customer back to checkout page
			c.HTML(400, "checkout_error.html", gin.H{
				"cancel":           checkout.CancelUrl,
				"messagePrimary":   defaultErrMessage,
				"messageSecondary": defaultErrMessageAction,
			})
			return
		}
	}
	logrus.Infof("(TxTracker)[%s] complete request authorize", breadTransactionID)

	// Auto settle the transaction if needed
	if account.AutoSettle && authSuccess {
		if err := settleTransaction(breadTransactionID, account, checkout); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"account":           account,
				"gatewayCheckoutId": gatewayCheckoutID,
				"transactionId":     breadTransactionID,
			}).Error("(GatewayCheckoutConfirmation) settling transaction produced error")
			c.HTML(400, "checkout_error.html", gin.H{
				"cancel":           checkout.CancelUrl,
				"messagePrimary":   defaultErrMessage,
				"messageSecondary": defaultErrMessageAction,
			})
			return
		}
		logrus.Infof("(TxTracker)[%s] complete request settle", breadTransactionID)
	}

	// Update gateway checkout
	// Should probably move this step to the success callback POST request
	err = completeGatewayCheckout(checkout.Id, breadTransactionID, "", h)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"gatewayCheckoutID":  gatewayCheckoutID,
			"breadTransactionID": breadTransactionID,
			"checkout":           checkout,
		}).Error("(GatewayCheckoutConfirmation) update to gateway checkout produced error")
	}

	// Calculate signature for get request query values
	response := &shopify.GatewayCheckoutCompleteRequest{
		AccountId:        string(checkout.AccountID),
		Reference:        checkout.Reference,
		Currency:         checkout.Currency,
		Test:             checkout.Test,
		Amount:           checkout.Amount,
		GatewayReference: breadTransactionID,
		Timestamp:        time.Now().UTC().Format(time.RFC3339),
		Result:           shopify.ResultComplete,
		TransactionType:  shopify.TxTypeAuthorization,
	}

	// Transaction Type should be "sale" for AutoSettle merchants
	if account.AutoSettle {
		response.TransactionType = shopify.TxTypeSale
	}

	signGatewayCheckoutResponse(response, account.GatewaySecret)

	v := url.Values{}
	v.Set("x_account_id", response.AccountId)
	v.Set("x_reference", response.Reference)
	v.Set("x_currency", response.Currency)
	v.Set("x_test", strconv.FormatBool(response.Test))
	v.Set("x_amount", strconv.FormatFloat(response.Amount, 'f', 2, 64))
	v.Set("x_gateway_reference", response.GatewayReference)
	v.Set("x_result", string(response.Result))
	v.Set("x_transaction_type", string(response.TransactionType))
	v.Set("x_timestamp", response.Timestamp)
	v.Set("x_signature", response.Signature)

	// Post callback to Shopify async
	go func() {
		err := HTTPFormRequest("POST", checkout.CallbackUrl, v, struct{}{})

		// Shopify policy: 5 retries on 60 second interval
		attempts := 1
		for err != nil && attempts < 6 {
			logrus.Info("(GatewayCheckoutConfirmation) retrying Shopify callback POST request")
			time.Sleep(time.Second * 60)
			attempts++
			err = HTTPFormRequest("POST", checkout.CallbackUrl, v, struct{}{})
		}

		// Log error if unsuccessful after 5 retries
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"checkout": checkout,
				"form":     v,
				"attempts": attempts,
			}).Error("(GatewayCheckoutComfirmation) making HTTP complete request produced error")
		}
		return
	}()

	redirectUrl := fmt.Sprintf("%s?%s", checkout.CompleteUrl, v.Encode())
	logrus.Infof("(TxTracker)[%s] complete request successful", breadTransactionID)
	c.Redirect(302, redirectUrl)
	return
}

func (h *Handlers) PlatformGatewayCheckoutConfirmation(c *gin.Context, dc desmond.Context) {
	// Log the transaction complete request for diagnosing offsite checkout pending transactions
	logrus.Info("(PlatformTxTracker) complete request start")

	// Pull id from params
	gatewayCheckoutID := c.Query("gatewayCheckoutId")
	transactionID := c.Query("transactionId")
	merchantID := c.Query("merchantId")

	// Log the transaction complete request for diagnosing offsite checkout pending transactions
	logrus.Infof("(PlatformTxTracker)[%s] complete request bind transaction id", transactionID)

	// Query for offsite checkout by reference
	checkout, err := findGatewayCheckoutById(zeus.Uuid(gatewayCheckoutID), h)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"gatewayCheckoutId":  gatewayCheckoutID,
			"breadTransactionId": transactionID,
		}).Error("(PlatformGatewayCheckoutConfirmation) search for gateway checkout produced error")

		// Short circuit flow and redirect customer back to checkout error page
		c.HTML(400, "checkout_error.html", gin.H{
			"messagePrimary":   defaultErrMessage,
			"messageSecondary": defaultErrMessageAction,
		})

		return
	}

	// Query for gateway
	account, err := findGatewayAccountById(checkout.AccountID, h)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"breadTransactionID": transactionID,
		}).Error("(PlatformGatewayGatewayCheckoutConfirmation) error querying for gateway by id")
		c.HTML(400, "checkout_error.html", gin.H{
			"cancel":           checkout.CancelUrl,
			"messagePrimary":   defaultErrMessage,
			"messageSecondary": defaultErrMessageAction,
		})
		return
	}

	logrus.Infof("(PlatformTxTracker)[%s] complete request processing checkout", transactionID)

	amountCents, err := types.USDToCents(checkout.AmountStr)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"breadTransactionID": transactionID,
			"amountStr":          checkout.AmountStr,
		}).Error("(PlatformGatewayGatewayCheckoutConfirmation) invalid checkout amount")
		c.HTML(400, "checkout_error.html", gin.H{
			"cancel":           checkout.CancelUrl,
			"messagePrimary":   defaultErrMessage,
			"messageSecondary": defaultErrMessageAction,
		})
		return
	}
	amount := bread.TrxAmount{
		Currency: checkout.Currency,
		Value:    amountCents,
	}
	trxReq := bread.TrxRequest{
		Amount:             amount,
		MerchantOfRecordID: merchantID,
	}

	cache := cache.NewCache(h.RedisPool)

	apiKey, apiSecret, host := platformApiParams(checkout.Test, account)

	var httpError *HttpError
	httpError = platformAuthorizeTransaction(host, apiKey, apiSecret, transactionID, trxReq, cache)
	if httpError != nil {
		logrus.WithFields(logrus.Fields{
			"account":       account,
			"transactionId": transactionID,
		}).Error("(PlatformGatewayCheckoutConfirmation) authorizing transaction produced error")

		// Short circuit flow and redirect customer back to checkout error page
		c.HTML(400, "checkout_error.html", gin.H{
			"cancel":           checkout.CancelUrl,
			"messagePrimary":   defaultErrMessage,
			"messageSecondary": defaultErrMessageAction,
		})
		return
	}
	logrus.Infof("(PlatformTxTracker)[%s] complete request authorize", transactionID)

	// Auto settle the transaction if needed
	if account.PlatformAutoSettle {
		if httpError = platformSettleTransaction(host, apiKey, apiSecret, transactionID, trxReq, cache); httpError != nil {
			logrus.WithFields(logrus.Fields{
				"account":       account,
				"transactionId": transactionID,
			}).Error("(PlatformGatewayCheckoutConfirmation) settling transaction produced error")

			// Short circuit flow and redirect customer back to checkout page
			c.HTML(400, "checkout_error.html", gin.H{
				"cancel":           checkout.CancelUrl,
				"messagePrimary":   defaultErrMessage,
				"messageSecondary": defaultErrMessageAction,
			})
			return
		}
		logrus.Infof("(PlatformTxTracker)[%s] complete request settle", transactionID)
	}

	// Update gateway checkout
	// Should probably move this step to the success callback POST request
	err = completeGatewayCheckout(checkout.Id, transactionID, merchantID, h)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"gatewayCheckoutID":  gatewayCheckoutID,
			"breadTransactionID": transactionID,
			"merchantID":         merchantID,
			"checkout":           checkout,
		}).Error("(PlatformGatewayCheckoutConfirmation) update to gateway checkout produced error")
	}

	// Calculate signature for get request query values
	response := &shopify.GatewayCheckoutCompleteRequest{
		AccountId:        string(checkout.AccountID),
		Reference:        checkout.Reference,
		Currency:         checkout.Currency,
		Test:             checkout.Test,
		Amount:           checkout.Amount,
		GatewayReference: transactionID,
		Timestamp:        time.Now().UTC().Format(time.RFC3339),
		Result:           shopify.ResultComplete,
		TransactionType:  shopify.TxTypeAuthorization,
	}

	// Transaction Type should be "sale" for AutoSettle merchants
	if account.AutoSettle {
		response.TransactionType = shopify.TxTypeSale
	}

	signGatewayCheckoutResponse(response, account.GatewaySecret)

	v := url.Values{}
	v.Set("x_account_id", response.AccountId)
	v.Set("x_reference", response.Reference)
	v.Set("x_currency", response.Currency)
	v.Set("x_test", strconv.FormatBool(response.Test))
	v.Set("x_amount", strconv.FormatFloat(response.Amount, 'f', 2, 64))
	v.Set("x_gateway_reference", response.GatewayReference)
	v.Set("x_result", string(response.Result))
	v.Set("x_transaction_type", string(response.TransactionType))
	v.Set("x_timestamp", response.Timestamp)
	v.Set("x_signature", response.Signature)

	// Post callback to Shopify async
	go func() {
		err := HTTPFormRequest("POST", checkout.CallbackUrl, v, struct{}{})

		// Shopify policy: 5 retries on 60 second interval
		attempts := 1
		for err != nil && attempts < 6 {
			logrus.Info("(GatewayCheckoutConfirmation) retrying Shopify callback POST request")
			time.Sleep(time.Second * 60)
			attempts++
			err = HTTPFormRequest("POST", checkout.CallbackUrl, v, struct{}{})
		}

		// Log error if unsuccessful after 5 retries
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"checkout": checkout,
				"form":     v,
				"attempts": attempts,
			}).Error("(GatewayCheckoutComfirmation) making HTTP complete request produced error")
		}
		return
	}()

	redirectUrl := fmt.Sprintf("%s?%s", checkout.CompleteUrl, v.Encode())
	logrus.Infof("(TxTracker)[%s] complete request successful", transactionID)
	c.Redirect(302, redirectUrl)
}
