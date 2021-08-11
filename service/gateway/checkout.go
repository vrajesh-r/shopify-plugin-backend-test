package gateway

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/getbread/breadkit/desmond"
	"github.com/getbread/breadkit/featureflags"
	"github.com/getbread/shopify_plugin_backend/service/bread"
	"github.com/getbread/shopify_plugin_backend/service/gateway/security"
	"github.com/getbread/shopify_plugin_backend/service/shopify"
	"github.com/getbread/shopify_plugin_backend/service/types"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func queryStringToMap(query string) map[string]string {
	pairs := strings.Split(query, "&")

	lookup := map[string]string{}
	for _, pair := range pairs {
		pieces := strings.Split(pair, "=")
		if len(pieces) == 2 {
			value, err := url.QueryUnescape(pieces[1])
			if err == nil {
				lookup[pieces[0]] = value
			} // catch
		} // catch
	}
	return lookup
}

func signGatewayCheckoutResponse(response *shopify.GatewayCheckoutCompleteRequest, secret string) {
	signatureInputs := map[string]string{
		"x_account_id":        response.AccountId,
		"x_reference":         response.Reference,
		"x_currency":          response.Currency,
		"x_test":              strconv.FormatBool(response.Test),
		"x_amount":            strconv.FormatFloat(response.Amount, 'f', 2, 64),
		"x_gateway_reference": response.GatewayReference,
		"x_timestamp":         response.Timestamp,
		"x_result":            string(response.Result),
		"x_transaction_type":  string(response.TransactionType),
	}
	response.Signature = security.GenerateGatewaySignature(signatureInputs, secret)
}

func stateIsNotValid(billingState, shippingState string) bool {
	bs := strings.ToLower(billingState)
	ss := strings.ToLower(shippingState)

	_, invalidBilling := invalidStateCodes[bs]
	_, invalidShipping := invalidStateCodes[ss]

	return invalidBilling || invalidShipping
}

func (h *Handlers) GatewayCheckout(c *gin.Context, dc desmond.Context) {
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

	// Query gateway by Gateway Key (x_account_id in request)
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

	// Transform hash to GatewayCheckoutRequest
	req := gatewayCheckoutRequest{
		AccountID:                 hash["x_account_id"],
		Currency:                  hash["x_currency"],
		Amount:                    amountDollarFloat,
		Reference:                 hash["x_reference"],
		ShopCountry:               hash["x_shop_country"],
		ShopName:                  hash["x_shop_name"],
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

	if account.ActiveVersion == BreadPlatform {
		// Save offsite checkout
		gc := types.GatewayCheckout{
			AccountID:    account.Id,
			Test:         req.Test,
			Reference:    req.Reference,
			Currency:     req.Currency,
			Amount:       amountDollarFloat,
			CallbackUrl:  req.CallbackURL,
			CompleteUrl:  req.CompleteURL,
			CancelUrl:    req.CancelURL,
			AmountStr:    hash["x_amount"],
			BreadVersion: BreadPlatform,
		}
		miltonCheckoutID, err := h.GatewayCheckoutCreator.Create(gc)
		if err != nil {
			log.WithFields(log.Fields{
				"error":           err.Error(),
				"request":         req,
				"gatewayCheckout": gc,
				"account":         account,
				"breadVersion":    BreadPlatform,
			}).Error("(GatewayCheckout) saving gateway checkout produced error")
			c.HTML(400, "checkout_error.html", gin.H{
				"cancel":           req.CancelURL,
				"messagePrimary":   "An error occurred while processing your request.",
				"messageSecondary": "Please contact customer support or choose another payment method.",
			})
			return
		}

		myShopifySubdomain := getMyshopifySubdomain(req.CallbackURL)
		shopifyCheckoutID := req.Invoice[1:]
		redisCheckout, httpError := getCheckoutFromRedis(h.RedisPool.Get(), shopifyCheckoutID)
		if httpError != nil {
			log.WithFields(log.Fields{
				"error":      httpError.Error(),
				"checkoutId": shopifyCheckoutID,
			}).Warn("(GatewayCheckout) Could not retrieve checkout from Redis")
		}

		platformPlacement, err := initPlatformPlacement(redisCheckout, hash["x_amount"], hash["x_currency"])
		if err != nil {
			log.WithFields(log.Fields{
				"error":    err.Error(),
				"checkout": redisCheckout,
			}).Error("(GatewayCheckout) Attempt to initiate bread platform placement failed")
			c.HTML(400, "checkout_error.html", gin.H{
				"cancel":           req.CancelURL,
				"messagePrimary":   "An error occurred while processing your request.",
				"messageSecondary": "Please contact customer support or choose another payment method.",
			})
			return
		}

		var integrationKey string
		if testTransaction {
			integrationKey = account.SandboxIntegrationKey
		} else {
			integrationKey = account.IntegrationKey
		}

		var platformCheckoutHost string
		if testTransaction {
			platformCheckoutHost = gatewayConfig.HostConfig.PlatformCheckoutHostDevelopment
		} else {
			platformCheckoutHost = gatewayConfig.HostConfig.PlatformCheckoutHost
		}

		c.HTML(200, "embedded_checkout_platform.html", gin.H{
			"BreadJS":            fmt.Sprintf("%s/sdk.js", platformCheckoutHost),
			"Setup":              initPlatformSetup(req, integrationKey),
			"Placement":          platformPlacement,
			"MyShopifySubdomain": myShopifySubdomain, // shop name in milton
			"ShopName":           req.ShopName,
			"HeapAppId":          gatewayConfig.HeapAppId,
			"Env":                gatewayConfig.Environment,
			"DatadogToken":       gatewayConfig.DataDogToken.Unmask(),
			"DatadogSite":        gatewayConfig.DatadogSite,
			"ShopifyCheckoutID":  shopifyCheckoutID,
			"CustomTotal":        amountMillicents.ToCents(),
			"CancelURL":          fmt.Sprintf("%s/gateway/checkout/cancel?orderRef=%s", gatewayConfig.HostConfig.MiltonHost, string(miltonCheckoutID)),
			"MiltonCheckoutID":   string(miltonCheckoutID),
			"CompleteURL":        fmt.Sprintf("%s/gateway/checkout/confirmation-platform", gatewayConfig.HostConfig.MiltonHost),
		})
		return
	}

	// Check for Shopify Plus setting and find matching PlusTransactionRecord
	if account.PlusEmbeddedCheckout {
		// Info log to diagnose embedded checkout pending transactions
		log.Infof("(TxTracker) Embedded checkout transaction processing initiated. Checkout id: [%s]", req.Reference)
		// Find PlusTransactionRecord
		checkout, err := findPlusGatewayCheckoutByCheckoutID(req.Reference, h)
		if err == nil {
			var host string
			var apiKey string
			var secretKey string
			if testTransaction {
				host = gatewayConfig.HostConfig.BreadHostDevelopment
				apiKey = account.SandboxApiKey
				secretKey = account.SandboxSharedSecret
			} else {
				host = gatewayConfig.HostConfig.BreadHost
				apiKey = account.ApiKey
				secretKey = account.SharedSecret
			}
			bc := bread.NewBreadClient(apiKey, secretKey)
			// Authorize and/or settle transaction
			err := authAndSettleTransaction(bc, checkout.TransactionID, account, host)
			if err != nil {
				log.WithFields(log.Fields{
					"error":         err.Error(),
					"transactionID": checkout.TransactionID,
					"account":       account,
				}).Error("(ShopifyPlusGatewayCheckout) authorizing or settling transaction produced error")
				c.String(400, err.Error())
				return
			}

			// Save "offsite" checkout -- gateway order management relies on this record
			gc := types.GatewayCheckout{
				AccountID:     account.Id,
				Test:          req.Test,
				Reference:     req.Reference,
				Currency:      req.Currency,
				Amount:        amountDollarFloat,
				TransactionID: checkout.TransactionID,
				Completed:     true,
				AmountStr:     hash["x_amount"],
				BreadVersion:  BreadClassic,
			}
			_, err = h.GatewayCheckoutCreator.Create(gc)
			if err != nil {
				log.WithFields(log.Fields{
					"error":           err.Error(),
					"request":         req,
					"gatewayCheckout": gc,
					"account":         account,
				}).Error("(ShopifyPlusGatewayCheckout) saving gateway checkout produced error")
			}

			// POST complete form to Shopify
			form := createShopifyCheckoutCompleteForm(account, req, checkout.TransactionID)
			if err := HTTPFormRequest("POST", req.CallbackURL, form, struct{}{}); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"request":  req,
					"checkout": checkout,
					"form":     form,
				}).Error("(ShopifyPlusGatewayCheckout) making HTTP complete request produced error")
				c.String(400, err.Error())
				return
			}

			// Redirect customer to Shopify confirmation page
			c.Redirect(302, req.CompleteURL)
			return
		} else {
			log.WithFields(log.Fields{
				"error":      err.Error(),
				"checkoutID": req.Reference,
			}).Error("(ShopifyPlusGatewayCheckout) Checkout not found")
		}
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

	// Perform request validation
	if req.CustomerFirstName == "" || (req.CustomerShippingFirstName == "" && req.CustomerShippingLastName != "") {
		log.WithFields(log.Fields{
			"billingFirstName":  req.CustomerFirstName,
			"shippingFirstName": req.CustomerShippingFirstName,
			"account":           account,
		}).Info("(GatewayCheckout) missing customer first name, cancelling checkout")
		c.HTML(400, "checkout_error.html", gin.H{
			"cancel":           req.CancelURL,
			"messagePrimary":   "Full name is required for Bread financing",
			"messageSecondary": "Please return to cart and provide your first name",
		})
		return
	}
	if req.CustomerBillingCountry != "US" || (req.CustomerShippingCountry != "US" && req.CustomerShippingCountry != "") {
		log.WithFields(log.Fields{
			"billingCountry":  req.CustomerBillingCountry,
			"shippingCountry": req.CustomerShippingCountry,
			"account":         account,
		}).Info("(GatewayCheckout) international country code, cancelling checkout")
		c.HTML(400, "checkout_error.html", gin.H{
			"cancel":           req.CancelURL,
			"messagePrimary":   "Unfortunately, Bread financing is only available to U.S. residents.",
			"messageSecondary": "Please choose another payment method.",
		})
		return
	}
	if stateIsNotValid(req.CustomerBillingState, req.CustomerShippingState) {
		log.WithFields(log.Fields{
			"billingState":  req.CustomerBillingState,
			"shippingState": req.CustomerShippingState,
			"account":       account,
		}).Info("(GatewayCheckout) U.S. territory or APO billing/shipping state used, cancelling checkout")
		c.HTML(400, "checkout_error.html", gin.H{
			"cancel":           req.CancelURL,
			"messagePrimary":   "Unfortunately, Bread financing is not available for U.S. territory or APO addresses.",
			"messageSecondary": "Please try a different address or choose another payment method.",
		})
		return
	}

	// Save offsite checkout
	gc := types.GatewayCheckout{
		AccountID:    account.Id,
		Test:         req.Test,
		Reference:    req.Reference,
		Currency:     req.Currency,
		Amount:       amountDollarFloat,
		CallbackUrl:  req.CallbackURL,
		CompleteUrl:  req.CompleteURL,
		CancelUrl:    req.CancelURL,
		AmountStr:    hash["x_amount"],
		BreadVersion: BreadClassic,
	}
	gcID, err := h.GatewayCheckoutCreator.Create(gc)
	if err != nil {
		log.WithFields(log.Fields{
			"error":           err.Error(),
			"request":         req,
			"gatewayCheckout": gc,
			"account":         account,
		}).Error("(GatewayCheckout) saving gateway checkout produced error")
		c.HTML(400, "checkout_error.html", gin.H{
			"cancel":           req.CancelURL,
			"messagePrimary":   "An error occurred while processing your request.",
			"messageSecondary": "Please contact customer support or choose another payment method.",
		})
		return
	}
	gc.Id = gcID

	// Transform req into request for POST /api/checkout/offsite
	CartOptions := bread.CartOptions{
		OrderRef:            string(gc.Id),
		CompleteUrl:         gatewayConfig.HostConfig.MiltonHost + "/gateway/checkout/confirmation",
		ErrorUrl:            gatewayConfig.HostConfig.MiltonHost + "/gateway/checkout/cancel?orderRef=" + string(gc.Id),
		CustomTotal:         amountMillicents.ToCents(),
		DisableEditShipping: true,
	}

	if account.TargetedFinancing && CartOptions.CustomTotal >= types.Cents(account.TargetedFinancingThreshold*100) {
		CartOptions.FinancingProgramID = account.TargetedFinancingID
	}

	CartOptions.BillingContact = bread.OptsContact{
		FullName: fmt.Sprintf("%s %s", req.CustomerFirstName, req.CustomerLastName),
		Address:  req.CustomerBillingAddress1,
		Address2: req.CustomerBillingAddress2,
		City:     req.CustomerBillingCity,
		State:    req.CustomerBillingState,
		Zip:      req.CustomerBillingZip,
		Phone:    req.CustomerBillingPhone,
		Email:    req.CustomerEmail,
	}
	CartOptions.ShippingContact = bread.OptsContact{
		FullName: fmt.Sprintf("%s %s", req.CustomerShippingFirstName, req.CustomerShippingLastName),
		Address:  req.CustomerShippingAddress1,
		Address2: req.CustomerShippingAddress2,
		City:     req.CustomerShippingCity,
		State:    req.CustomerShippingState,
		Zip:      req.CustomerShippingZIP,
		Phone:    req.CustomerShippingPhone,
	}

	//Evaluate LaunchDarkly MiltonHostedCheckoutFlag
	const FeatureFlagName = "milton-hosted-checkout-flag"
	enableMiltonHostedCheckout := featureflags.GetBool(FeatureFlagName, false, map[string]interface{}{
		"email": account.Email,
	})

	if enableMiltonHostedCheckout {
		var checkoutHost string
		if testTransaction {
			checkoutHost = gatewayConfig.HostConfig.CheckoutHostDevelopment
		} else {
			checkoutHost = gatewayConfig.HostConfig.CheckoutHost
		}

		log.Info("Rendering embedded checkout")

		c.HTML(200, "embedded_checkout.html", gin.H{
			"BreadJS":            fmt.Sprintf("%s/bread.js", checkoutHost),
			"BreadAPIKey":        apiKey,
			"Invoice":            req.Invoice,
			"CartOptions":        CartOptions,
			"MyShopifySubdomain": getMyshopifySubdomain(req.CallbackURL),
			"ShopName":           req.ShopName,
			"HeapAppId":          gatewayConfig.HeapAppId,
			"Env":                gatewayConfig.Environment,
			"DatadogToken":       gatewayConfig.DataDogToken.Unmask(),
			"DatadogSite":        gatewayConfig.DatadogSite,
		})

		return
	}

	cartCreateRequest := bread.Cart{
		Options:    CartOptions,
		CartOrigin: "shopify_redirect",
	}

	// Make request to Ostia and get redirect url
	bc := bread.NewClient(apiKey, secretKey)
	var bhost string
	if testTransaction {
		bhost = gatewayConfig.HostConfig.BreadHostDevelopment
	} else {
		bhost = gatewayConfig.HostConfig.BreadHost
	}

	savedCart, err := bc.SaveCart(bhost, &cartCreateRequest)
	if err != nil {
		log.WithFields(log.Fields{
			"request": cartCreateRequest,
			"account": account,
		}).Errorf("(GatewayCheckout) request for cart url produced error: %s", err.Error())
		c.HTML(400, "checkout_error.html", gin.H{
			"cancel":           req.CancelURL,
			"messagePrimary":   "An error occurred while processing your request.",
			"messageSecondary": "Please contact customer support or choose another payment method.",
		})
		return
	}

	// Redirect request to offsite checkout
	c.Redirect(302, savedCart.Url)
}

// Strip subdomain from callback URL request value. This maps to shop name on Milton
func getMyshopifySubdomain(callbackURL string) string {
	idx := strings.Index(callbackURL, "//")
	return strings.Split(callbackURL[idx+2:], ".")[0]
}

func aggregateErrors(e error, errors []error) {
	if e != nil {
		errors = append(errors, e)
	}
}

func initPlatformPlacement(
	res *types.CreateCheckoutRequest,
	checkoutTotalUSD string,
	checkoutCurrency string) (*bread.PlatformPlacement, error) {

	var currency string
	var err error
	var errors []error
	var orderItem bread.PlatformOrderItem
	var subTotal, totalShipping, totalDiscounts, totalTax, grandTotal, totalLineItemsPrice types.Cents

	orderItems := []bread.PlatformOrderItem{}
	if res != nil {
		totalLineItemsPrice, err = types.USDToCents(res.TotalPrice)
		aggregateErrors(err, errors)
	}

	checkoutTotal, err := types.USDToCents(checkoutTotalUSD)
	aggregateErrors(err, errors)

	if res == nil || totalLineItemsPrice != checkoutTotal {
		// Checkout total sent in the request to the gateway
		// differs from the total calculated from the lineitems details saved when shopify calls
		// the checkout create or checkout update webhook.
		// In scenarios like this the total sent to the gateway takes precedence

		currency = checkoutCurrency
		orderItem = bread.PlatformOrderItem{
			Name:                "N/A",
			Sku:                 "",
			UnitPrice:           bread.TrxAmount{Currency: checkoutCurrency, Value: checkoutTotal},
			ShippingCost:        bread.TrxAmount{Currency: checkoutCurrency, Value: 0},
			ShippingDescription: "",
			UnitTax:             bread.TrxAmount{Currency: checkoutCurrency, Value: 0},
			Brand:               "",
			Currency:            checkoutCurrency,
			Quantity:            1,
		}
		orderItems = append(orderItems, orderItem)
		grandTotal = checkoutTotal
		subTotal = checkoutTotal
	} else {
		currency := res.PresentmentCurrency

		var itemTax types.Cents
		totalTax, err = types.USDToCents(res.TotalTax)
		aggregateErrors(err, errors)

		for i, item := range res.LineItems {
			unitPrice, err := types.USDToCents(item.Price)
			aggregateErrors(err, errors)

			if i == 0 {
				// Assign total tax to first itemTax
				// Tax per product is not available from shopify
				// This is done to make validation of the sum of each line item tax against total tax
				// by the SDK pass

				itemTax = totalTax
			}

			orderItem = bread.PlatformOrderItem{
				Name:                item.Title,
				Sku:                 item.Sku,
				UnitPrice:           bread.TrxAmount{Currency: currency, Value: unitPrice},
				ShippingCost:        bread.TrxAmount{Currency: currency, Value: 0},
				ShippingDescription: "",
				UnitTax:             bread.TrxAmount{Currency: currency, Value: itemTax},
				Brand:               "",
				Currency:            currency,
				Quantity:            item.Quantity,
			}

			orderItems = append(orderItems, orderItem)
		}

		subTotal, err = types.USDToCents(res.SubtotalPrice)
		aggregateErrors(err, errors)
		totalShipping, err = calculateTotalShipping(res.ShippingLines)
		aggregateErrors(err, errors)
		totalDiscounts, err = types.USDToCents(res.TotalDiscounts)
		aggregateErrors(err, errors)
		grandTotal = totalLineItemsPrice
	}

	if len(errors) != 0 {
		return nil, errors[0]
	}

	platformOrder := bread.PlatformOrder{
		Items:          orderItems,
		SubTotal:       bread.TrxAmount{Currency: currency, Value: subTotal},
		TotalTax:       bread.TrxAmount{Currency: currency, Value: totalTax},
		TotalShipping:  bread.TrxAmount{Currency: currency, Value: totalShipping},
		TotalDiscounts: bread.TrxAmount{Currency: currency, Value: totalDiscounts},
		TotalPrice:     bread.TrxAmount{Currency: currency, Value: grandTotal},
	}

	return &bread.PlatformPlacement{
		DomID:         "placement-checkout",
		AllowCheckout: true,
		Order:         platformOrder,
	}, nil
}

func calculateTotalShipping(shippingLines []types.ShippingLine) (types.Cents, error) {
	totalShipping := types.Cents(0)

	for _, line := range shippingLines {
		cents, err := types.USDToCents(line.Price)
		if err != nil {
			return types.Cents(0), err
		}
		totalShipping += cents
	}

	return totalShipping, nil
}

func initPlatformSetup(req gatewayCheckoutRequest, integrationKey string) bread.PlatformSetup {
	billingAddress := bread.PlatformSetupAddress{
		Address1:   req.CustomerBillingAddress1,
		Address2:   req.CustomerBillingAddress2,
		Country:    req.CustomerBillingCountry,
		Region:     req.CustomerBillingState,
		Locality:   req.CustomerBillingCity,
		PostalCode: req.CustomerBillingZip,
	}

	shippingAddress := bread.PlatformSetupAddress{
		Address1:   req.CustomerShippingAddress1,
		Address2:   req.CustomerShippingAddress2,
		Country:    req.CustomerShippingCountry,
		Region:     req.CustomerShippingState,
		Locality:   req.CustomerShippingCity,
		PostalCode: req.CustomerShippingZIP,
	}

	buyer := bread.PlatformSetupBuyer{
		GivenName:       req.CustomerFirstName,
		FamilyName:      req.CustomerLastName,
		Email:           req.CustomerEmail,
		Phone:           req.CustomerPhone,
		BillingAddress:  billingAddress,
		ShippingAddress: shippingAddress,
	}

	return bread.PlatformSetup{
		IntegrationKey: integrationKey,
		Buyer:          buyer,
	}
}

func authAndSettleTransaction(bc bread.IBreadClient, transactionID string, account types.GatewayAccount, host string) error {
	// Authorize and/or settle transaction
	log.Infof("(TxTracker)[%s] ShopifyPlusGatewayCheckout authorize transaction initiated ...", transactionID)

	authorizeRequest := &bread.TransactionActionRequest{
		Type: "authorize",
	}
	_, err := bc.AuthorizeTransaction(transactionID, host, authorizeRequest)
	if err != nil {
		return err
	}

	// auto_settle the transaction if needed
	if account.AutoSettle {
		settleRequest := &bread.TransactionActionRequest{
			Type: "settle",
		}
		_, err := bc.SettleTransaction(transactionID, host, settleRequest)
		if err != nil {
			return err
		}
	}
	return nil
}

func createShopifyCheckoutCompleteForm(account types.GatewayAccount, req gatewayCheckoutRequest, transactionID string) url.Values {
	// Mark checkout as complete on Shopify
	response := &shopify.GatewayCheckoutCompleteRequest{
		AccountId:        account.GatewayKey,
		Reference:        req.Reference,
		Currency:         req.Currency,
		Test:             req.Test,
		Amount:           req.Amount,
		GatewayReference: transactionID,
		Timestamp:        time.Now().UTC().Format(time.RFC3339),
		Result:           shopify.ResultComplete,
		TransactionType:  shopify.TxTypeAuthorization,
	}
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

	return form
}

func getCheckoutFromRedis(conn redis.Conn, checkoutId string) (*types.CreateCheckoutRequest, *HttpError) {

	var key string = fmt.Sprintf("checkout-%s", checkoutId)

	// Get checkout from redis
	rawCheckout, err := conn.Do("GET", key)
	defer conn.Close()

	if err != nil || rawCheckout == nil {
		return nil, NewHttpError(fmt.Sprintf("Checkout not found: %s", checkoutId), 400)
	}

	// Unmarshal checkout to a CreateCheckoutRequest
	bytesCheckout, ok := rawCheckout.([]byte)

	if !ok {
		return nil, NewHttpError("Cannot convert checkout to bytes array", 500)
	}

	checkoutResp := &types.CreateCheckoutRequest{}

	if err := json.Unmarshal(bytesCheckout, checkoutResp); err != nil {
		return nil, NewHttpError(fmt.Sprintf("Failed to unmarshal JSON: %s", err), 500)
	}

	return checkoutResp, nil
}

func (h *Handlers) GetCheckout(c *gin.Context, dc desmond.Context) {
	var checkoutId string = c.Params.ByName("CheckoutId")

	checkoutResp, httpError := getCheckoutFromRedis(h.RedisPool.Get(), checkoutId)
	if httpError != nil {
		log.Error(fmt.Sprintf("%s %s", "(GetCheckout)", httpError.Error()))
		c.JSON(httpError.Code, gin.H{})
		return
	}

	c.JSON(200, checkoutResp)
}

func (h *Handlers) GetTenant(c *gin.Context) {
	GatewayTenant := gatewayConfig.GatewayTenant
	if GatewayTenant == "RBC" {
		c.JSON(200, gin.H{"tenant": "RBC"})
	} else {
		c.JSON(200, gin.H{"tenant": "BREAD"})
	}
}
