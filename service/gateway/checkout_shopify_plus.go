package gateway

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"strconv"
	"time"

	"github.com/getbread/breadkit/desmond"
	"github.com/getbread/shopify_plugin_backend/service/bread"
	"github.com/getbread/shopify_plugin_backend/service/gateway/security"
	"github.com/getbread/shopify_plugin_backend/service/shopify"
	"github.com/getbread/shopify_plugin_backend/service/types"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type transactionRecordRequest struct {
	CheckoutID    string `json:"checkoutID"`
	TransactionID string `json:"transactionID"`
}

func (h *Handlers) PlusGatewayTransactionRecord(c *gin.Context, dc desmond.Context) {
	// parse body into request struct
	var req transactionRecordRequest
	err := c.BindJSON(&req)
	if err != nil {
		log.WithFields(log.Fields{
			"error":       err.Error(),
			"queryString": c.Request.URL.RawQuery,
		}).Error("(PlusGatewayTransactionRecord) binding request to produced an error")
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	pgc := types.PlusGatewayCheckout{
		CheckoutID:    req.CheckoutID,
		TransactionID: req.TransactionID,
	}

	pgcID, err := h.PlusGatewayCheckoutCreator.Create(pgc)
	c.JSON(200, gin.H{
		"pgcID": pgcID,
	})
}

func (h *Handlers) PlusGatewayCheckout(c *gin.Context) {
	// Parse request string into map[string]string
	bb, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"error":       err.Error(),
			"requestBody": c.Request.Body,
		}).Error("(GatewayCheckout) reading request body produced error")
		c.HTML(400, "checkout_error.html", gin.H{
			"messagePrimary":   "An error occurred while processing your request.",
			"messageSecondary": "Please contact customer support or choose another payment method.",
		})
		return
	}
	hash := queryStringToMap(string(bb))

	// Query shop by Gateway Key (x_account_id in request)
	account, err := findGatewayAccountByGatewayKey(hash["x_account_id"], h)
	if err != nil {
		log.WithFields(log.Fields{
			"error":      err.Error(),
			"requestMap": hash,
		}).Error("(GatewayCheckout) query for gateway account produced error")
		c.HTML(400, "checkout_error.html", gin.H{
			"cancel":           hash["x_url_cancel"],
			"messagePrimary":   "Gateway account not found. Please confirm your gateway credentials.",
			"messageSecondary": "Review Shopify documentation for more information.",
			"link":             "https://docs.getbread.com/docs/integration/shopify/troubleshooting/#gateway-keys-versus-bread-api-keys",
		})
		return
	}

	// Verify authenticity
	if !security.GatewayRequestAuthentic(hash, account.GatewaySecret, hash["x_signature"]) {
		log.WithFields(log.Fields{
			"requestMap": hash,
		}).Error("(GatewayCheckout) request signature invalid")
		c.HTML(400, "checkout_error.html", gin.H{
			"cancel":           hash["x_url_cancel"],
			"messagePrimary":   "An error occurred while processing your request.",
			"messageSecondary": "Please contact customer support or choose another payment method.",
		})
		return
	}

	// Type conversions
	var errors []error
	var aggregateErrors = func(e error, errors []error) {
		if e != nil {
			errors = append(errors, e)
		}
	}

	amountMillicents, err := types.USDToMillicents(hash["x_amount"])
	amountDollarFloat := float64(amountMillicents.ToCents()) / 100.00
	aggregateErrors(err, errors)
	var testTransaction bool
	if hash["x_test"] == "true" {
		testTransaction = true
	} else {
		testTransaction = false
	}

	if len(errors) > 0 {
		log.WithFields(log.Fields{
			"error":      errors[0].Error(),
			"requestMap": hash,
		}).Error("(GatewayCheckout) type conversions from request body produced error")
		c.HTML(400, "checkout_error.html", gin.H{
			"cancel":           hash["x_url_cancel"],
			"messagePrimary":   "An error occurred while processing your request.",
			"messageSecondary": "Please contact customer support or choose another payment method.",
		})
		return
	}

	// Determine API and Secret Key
	var apiKey string
	var secretKey string
	if testTransaction {
		apiKey = account.SandboxApiKey
		secretKey = account.SandboxSharedSecret
	} else {
		apiKey = account.ApiKey
		secretKey = account.SharedSecret
	}
	if apiKey == "" || secretKey == "" {
		log.WithFields(log.Fields{
			"apiKey":    apiKey,
			"secretKey": secretKey,
			"account":   account,
		}).Error("(GatewayCheckout) missing Bread API keys")
		c.HTML(400, "checkout_error.html", gin.H{
			"cancel":           hash["x_url_cancel"],
			"messagePrimary":   "Missing Bread API keys. Please confirm your Bread API keys in the Shopify Portal.",
			"messageSecondary": "Review Shopify documentation for more information.",
			"link":             "https://docs.getbread.com/docs/integration/shopify/installing-the-shopify-bread-app/#the-bread-payment-gateway",
		})
		return
	}

	// Transform hash to GatewayCheckoutRequest
	req := gatewayCheckoutRequest{
		AccountID:                 hash["x_account_id"],
		Currency:                  hash["x_currency"],
		Amount:                    amountDollarFloat,
		Reference:                 hash["x_reference"],
		ShopCountry:               hash["x_shop_country"],
		ShopName:                  hash["x_shop_country"],
		TransactionType:           hash["x_transaction_type"],
		Description:               hash["x_description"],
		Invoice:                   hash["x_invoice"],
		Test:                      testTransaction,
		CustomerFirstName:         hash["x_customer_first_name"],
		CustomerLastName:          hash["x_customer_last_name"],
		CustomerEmail:             hash["x_customer_email"],
		CustomerPhone:             hash["x_customer_phone"],
		CustomerBillingCity:       hash["x_customer_billing_city"],
		CustomerBillingAddress1:   hash["x_customer_billing_address1"],
		CustomerBillingAddress2:   hash["x_customer_billing_address2"],
		CustomerBillingState:      hash["x_customer_billing_state"],
		CustomerBillingZip:        hash["x_customer_billing_zip"],
		CustomerBillingCountry:    hash["x_customer_billing_country"],
		CustomerBillingPhone:      hash["x_customer_billing_phone"],
		CustomerShippingFirstName: hash["x_customer_shipping_first_name"],
		CustomerShippingLastName:  hash["x_customer_shipping_last_name"],
		CustomerShippingCity:      hash["x_customer_shipping_city"],
		CustomerShippingAddress1:  hash["x_customer_shipping_address1"],
		CustomerShippingAddress2:  hash["x_customer_shipping_address2"],
		CustomerShippingState:     hash["x_customer_shipping_state"],
		CustomerShippingZIP:       hash["x_customer_shipping_zip"],
		CustomerShippingCountry:   hash["x_customer_shipping_country"],
		CustomerShippingPhone:     hash["x_customer_shipping_phone"],
		CallbackURL:               hash["x_url_callback"],
		CancelURL:                 hash["x_url_cancel"],
		CompleteURL:               hash["x_url_complete"],
		Timestamp:                 hash["x_timestamp"],
		Signature:                 hash["x_signature"],
	}

	// Find PlusTransactionRecord
	checkout, err := findPlusGatewayCheckoutByCheckoutID(req.Reference, h)
	if err != nil {
		c.HTML(400, "checkout_error.html", gin.H{
			"cancel":           hash["x_url_cancel"],
			"messagePrimary":   "We couldn't find your transaction record.",
			"messageSecondary": "Review Shopify documentation for more information.",
			"link":             "https://docs.getbread.com/docs/integration/shopify/installing-the-shopify-bread-app/#the-bread-payment-gateway",
		})
		return
	}

	var host string
	if testTransaction {
		host = gatewayConfig.HostConfig.BreadHostDevelopment
	} else {
		host = gatewayConfig.HostConfig.BreadHost
	}
	bc := bread.NewClient(apiKey, secretKey)

	authorizeRequest := &bread.TransactionActionRequest{
		Type: "authorize",
	}
	_, err = bc.AuthorizeTransaction(checkout.TransactionID, host, authorizeRequest)
	if err != nil {
		fmt.Println("Error:", err)
		c.String(400, err.Error())
		return
	}

	// auto_settle the transaction if needed
	if account.AutoSettle {
		settleRequest := &bread.TransactionActionRequest{
			Type: "settle",
		}
		_, err := bc.SettleTransaction(checkout.TransactionID, host, settleRequest)
		if err != nil {
			fmt.Println("Error:", err)
			c.String(400, err.Error())
			return
		}
	}

	// Mark checkout as complete on Shopify
	response := &shopify.GatewayCheckoutCompleteRequest{
		AccountId:        account.GatewayKey,
		Reference:        req.Reference,
		Currency:         req.Currency,
		Test:             req.Test,
		Amount:           req.Amount,
		GatewayReference: checkout.TransactionID,
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

	if err := HTTPFormRequest("POST", req.CallbackURL, form, struct{}{}); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"request":  req,
			"checkout": checkout,
			"form":     form,
		}).Error("(GatewayCheckoutComplete) making HTTP complete request produced error")
		c.String(400, err.Error())
		return
	}

	// Redirect request to offsite checkout
	c.Redirect(302, req.CompleteURL)
}
