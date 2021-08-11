package app

import (
	"database/sql"
	"strconv"
	"strings"
	"unicode"

	"github.com/getbread/shopify_plugin_backend/service/dbhandlers"

	"github.com/getbread/breadkit/desmond"
	"github.com/getbread/breadkit/featureflags"
	zeus "github.com/getbread/breadkit/zeus/types"
	"github.com/getbread/shopify_plugin_backend/service/bread"
	"github.com/getbread/shopify_plugin_backend/service/shopify"
	"github.com/getbread/shopify_plugin_backend/service/types"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const missingTransactionNote = "DO NOT FULFILL: This order does not contain a Bread transaction and therefore is not a valid Bread POS order. Please direct the customer to complete the Bread checkout form before submitting the Bread POS order."
const failedAuthorizationNote = "DO NOT FULFILL: Bread was unable to authorize this transaction. Please contact integrations@breadfinance.com for help."
const failedSettlementNote = "DO NOT FULFILL: Bread was unable to settle this transaction. Please contact integrations@breadfinance.com for help."

type createOrderRequest struct {
	Id                int                `json:"id"`
	CheckoutId        zeus.NullInt64     `json:"checkout_id"`
	CheckoutToken     zeus.NullString    `json:"checkout_token"`
	Gateway           string             `json:"gateway"`
	NoteAttributes    []Notes            `json:"note_attributes"`
	OrderId           int                `json:"order_id"`
	OrderNumber       int                `json:"order_number"`
	Customer          shopify.Customer   `json:"customer"`
	TotalPrice        string             `json:"total_price"`
	FinancialStatus   string             `json:"financial_status"`
	FulfillmentStatus string             `json:"fulfillment_status"`
	Test              bool               `json:"test"`
	BillingAddress    shopify.Address    `json:"billing_address"`
	ShippingAddress   shopify.Address    `json:"shipping_address"`
	LineItems         []shopify.LineItem `json:"line_items"`
}

type Notes struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

func (h *Handlers) CreateOrder(c *gin.Context, dc desmond.Context) {
	c.String(200, "success")

	var req createOrderRequest
	if err := c.BindJSON(&req); err != nil {
		logrus.WithFields(logrus.Fields{
			"error":       err.Error(),
			"queryString": c.Request.URL.RawQuery,
		}).Error("(WebhookCreateOrder) binding request to model produced error")
		return
	}

	// Pull shop name from request header
	shopDomain := c.Request.Header.Get("X-Shopify-Shop-Domain")
	shopName := strings.Split(shopDomain, ".")[0]

	if isBreadGateway(req.Gateway) {
		go processGatewayOrder(req, shopName, h)
	} else if isBreadPOSOrder(req.Gateway) {
		go processPOSOrder(req, shopName, h)
	}

	if orderPlacedInUSorCA(req.BillingAddress.CountryCode, req.ShippingAddress.CountryCode, req.Customer.DefaultAddress.CountryCode) {
		go saveForAnalytics(req, shopName, h)
	}

	if isBreadOrder(req.Gateway) && featureflags.GetBool("milton-gift-card-tracking", false) {
		go checkForGiftCards(req, shopName, h.GiftCardOrderCreator)
	}

}

func isBreadGateway(gateway string) bool {
	return gateway == "bread" || gateway == "bread_staging_" || gateway == "bread_development_" || gateway == "bread_sandbox_"
}

func isBreadPOSOrder(gateway string) bool {
	return removeSpacesAndDashes(gateway) == breadPOSGatewayName
}

func isBreadAppOrder(gateway string) bool {
	return gateway == breadAppGatewayName
}

func isBreadOrder(gateway string) bool {
	return isBreadGateway(gateway) || isBreadAppOrder(gateway) || isBreadPOSOrder(gateway)
}

func removeSpacesAndDashes(s string) string {
	return strings.ToLower(strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || r == '-' {
			return -1
		}
		return r
	}, s))
}

func orderPlacedInUSorCA(billingCountry, shippingCountry, defaultCountry string) bool {
	if billingCountry == "" && shippingCountry == "" {
		return (defaultCountry == "US" || defaultCountry == "CA")
	}
	return (billingCountry == "US" || billingCountry == "CA" || billingCountry == "") && (shippingCountry == "US" || shippingCountry == "CA" || shippingCountry == "")
}

func saveForAnalytics(req createOrderRequest, shopName string, h *Handlers) {
	ao := types.AnalyticsOrder{
		OrderID:           int64(req.Id),
		ShopName:          shopName,
		CheckoutID:        req.CheckoutId,
		CheckoutToken:     req.CheckoutToken,
		CustomerID:        int64(req.Customer.ID),
		CustomerEmail:     zeus.NullString{NullString: sql.NullString{String: req.Customer.Email, Valid: true}},
		TotalPrice:        zeus.NullString{NullString: sql.NullString{String: req.TotalPrice, Valid: true}},
		Gateway:           zeus.NullString{NullString: sql.NullString{String: req.Gateway, Valid: true}},
		FinancialStatus:   zeus.NullString{NullString: sql.NullString{String: req.FinancialStatus, Valid: true}},
		FulfillmentStatus: zeus.NullString{NullString: sql.NullString{String: req.FulfillmentStatus, Valid: true}},
		Test:              req.Test,
	}

	_, err := h.AnalyticsOrderCreator.Create(ao)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":    err.Error(),
			"order_id": req.Id,
			"shop":     shopName,
			"gateway":  req.Gateway,
			"test":     req.Test,
		}).Error("(WebhookAnalyticsOrder) saving order to db produced an error")
	}
	return
}

// Returns individual bools for whether the string contains "gift" and "card"
func stringContainsGiftOrCard(s string) (bool, bool) {
	s = strings.ToLower(s)
	return strings.Contains(s, "gift"), strings.Contains(s, "card")
}

// Save potential gift cards in our database for tracking according to the following rules:
// 1. Item price is greater than $0.00 dollars to filter out gift items
// 2. Item is a Shopify-created gift card
// 3. Item name contains the phrase "Gift" i.e. "Gift Card" or "Gift Certificate"
func checkForGiftCards(req createOrderRequest, shopName string, giftCardCreator dbhandlers.GiftCardOrderCreator) {
	for _, lineItem := range req.LineItems {
		// Check if item price is greater than $0.00
		itemPriceGreaterThanZero := true
		itemPrice, err := strconv.ParseFloat(lineItem.Price, 32)
		if err == nil {
			itemPriceGreaterThanZero = itemPrice > 0
		} else {
			logrus.WithFields(logrus.Fields{
				"error":    err.Error(),
				"order_id": req.Id,
				"shop":     shopName,
				"gateway":  req.Gateway,
				"test":     req.Test,
			}).Error("(WebhookCreateOrder) error converting line_item price from string to float")
		}

		// Check if item name contains the substring "gift" or "card"
		nameContainsGift, nameContainsCard := stringContainsGiftOrCard(lineItem.Name)

		if itemPriceGreaterThanZero && (lineItem.GiftCard == true || nameContainsGift) {
			gco := types.GiftCardOrder{
				OrderID:              int64(req.Id),
				ShopName:             shopName,
				Gateway:              req.Gateway,
				Test:                 req.Test,
				ItemName:             lineItem.Name,
				ItemPrice:            lineItem.Price,
				Quantity:             int64(lineItem.Quantity),
				RequiresShipping:     lineItem.RequiresShipping,
				IsShopifyGiftCard:    lineItem.GiftCard,
				NameContainsGiftOnly: nameContainsGift && !nameContainsCard,
				NameContainsGiftCard: nameContainsGift && nameContainsCard,
			}

			_, err := giftCardCreator.Create(gco)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error":    err.Error(),
					"order_id": req.Id,
					"shop":     shopName,
					"gateway":  req.Gateway,
					"test":     req.Test,
				}).Error("(WebhookCreateOrder) error saving line_item as gift card")
			}
		}
	}
}

func processGatewayOrder(req createOrderRequest, shopName string, h *Handlers) {
	checkoutID, _ := req.CheckoutId.Value()
	if checkoutID == nil {
		logrus.WithFields(logrus.Fields{
			"gateway":  req.Gateway,
			"order_id": req.OrderId,
		}).Error("(WebhookCreateOrder) checkout_id is null, unable to process gateway order")
		return
	}

	checkoutIDStr := strconv.FormatInt(checkoutID.(int64), 10)
	gc, err := findCompletedGatewayCheckoutByReference(checkoutIDStr, h)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":       err.Error(),
			"gateway":     req.Gateway,
			"checkout_id": checkoutIDStr,
		}).Error("(WebhookCreateOrder) search for gateway checkout produced error")
		return
	}

	if gc.BreadVersion == BreadPlatform {
		return
	}

	if gc.TransactionID == "" {
		logrus.WithFields(logrus.Fields{
			"gateway":           req.Gateway,
			"orderNumber":       req.OrderNumber,
			"checkoutID":        checkoutIDStr,
			"accountID":         gc.AccountID,
			"gatewayCheckoutID": gc.Id,
		}).Error("(WebhookCreateOrder) gateway checkout contains empty transactionID")
		return
	}

	account, err := findGatewayAccountById(gc.AccountID, h)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":       err.Error(),
			"gateway":     req.Gateway,
			"checkout_id": checkoutIDStr,
			"account_id":  gc.AccountID,
		}).Error("(WebhookCreateOrder) search for gateway account produced error")
		return
	}
	var apiKey string
	var sharedSecret string
	var host string
	if gc.Test {
		apiKey = account.SandboxApiKey
		sharedSecret = account.SandboxSharedSecret
		host = appConfig.HostConfig.BreadHostDevelopment
	} else {
		apiKey = account.ApiKey
		sharedSecret = account.SharedSecret
		host = appConfig.HostConfig.BreadHost
	}
	bc := bread.NewClient(apiKey, sharedSecret)

	updateRequest := &bread.TransactionActionRequest{
		MerchantOrderId: strconv.Itoa(req.OrderNumber),
	}
	_, err = bc.UpdateTransaction(gc.TransactionID, host, updateRequest)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":         err.Error(),
			"orderNumber":   req.OrderNumber,
			"gateway":       req.Gateway,
			"checkoutID":    checkoutIDStr,
			"account":       account,
			"transactionID": gc.TransactionID,
		}).Error("(WebhookCreateOrder) updating transaction with order number produced an error")
	}
	return
}

func processPOSOrder(req createOrderRequest, shopName string, h *Handlers) {
	// Pull order note attributes
	var transactionId string
	var ok bool
	for _, n := range req.NoteAttributes {
		if n.Name == "breadTxId" {
			transactionId, ok = n.Value.(string)
			if !ok {
				logrus.WithFields(logrus.Fields{
					"transactionId": n.Value,
					"shopName":      shopName,
					"request":       req,
				}).Error("(WebhookCreateOrder) unable to type assert transactionID to string")
			}
		}
	}

	shop, err := findShopByName(shopName, h)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":    err.Error(),
			"shopName": shopName,
			"request":  req,
		}).Error("(WebhookCreateOrder) could not find shop")
		return
	}

	if transactionId == "" {
		// Proactively void Bread POS orders that are accidentally submitted without a Bread transaction
		voidShopifyPOSOrder(req.Id, shop, missingTransactionNote)
		logrus.WithFields(logrus.Fields{
			"shopName": shopName,
			"request":  req,
		}).Error("(WebhookCreateOrder) no Bread transaction ID, canceling Shopify order")
		return
	}

	// Initialize Bread HTTP client
	bc := bread.NewClient(shop.GetAPIKeys())

	// Authorize Bread transaction
	authorizeRequest := &bread.TransactionActionRequest{
		Type:            "authorize",
		MerchantOrderId: strconv.Itoa(req.OrderNumber),
	}
	_, err = bc.AuthorizeTransaction(transactionId, shop.BreadHost(), authorizeRequest)
	if err != nil {
		// Retry authorization once
		_, err = bc.AuthorizeTransaction(transactionId, shop.BreadHost(), authorizeRequest)
		if err != nil {
			voidShopifyPOSOrder(req.Id, shop, failedAuthorizationNote)
			logrus.WithFields(logrus.Fields{
				"error":         err.Error(),
				"request":       req,
				"shop":          shop,
				"transactionId": transactionId,
			}).Error("(WebhookCreateOrder) authorizing transaction produced error, canceling Shopify order")
			return
		}
	}

	// Auto-settle Bread transaction
	settleRequest := &bread.TransactionActionRequest{
		Type: "settle",
	}
	_, err = bc.SettleTransaction(transactionId, shop.BreadHost(), settleRequest)
	if err != nil {
		// Retry settle once
		_, err = bc.SettleTransaction(transactionId, shop.BreadHost(), settleRequest)
		if err != nil {
			voidShopifyPOSOrder(req.Id, shop, failedSettlementNote)
			logrus.WithFields(logrus.Fields{
				"error":         err.Error(),
				"request":       req,
				"shop":          shop,
				"transactionId": transactionId,
			}).Error("(WebhookCreateOrder) settling transaction produced error, canceling Shopify order")
			return
		}
	}

	// Add pos-order to Milton order -> transaction lookup
	_, err = createOrder(shop, transactionId, req.Id, h)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":         err.Error(),
			"request":       req,
			"shop":          shop,
			"transactionId": transactionId,
		}).Error("(WebhookCreateOrder) adding order to Milton lookup table produced error")
		return
	}
}

func voidShopifyPOSOrder(orderId int, shop types.Shop, orderNote string) {
	sc := shopify.NewClient(shop.Shop, shop.AccessToken)

	// Query Shopify order for associated transactions
	var tr shopify.SearchTransactionsResponse
	err := sc.QueryTransactions(strconv.Itoa(orderId), &tr)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":   err.Error(),
			"orderId": orderId,
			"shop":    shop.Shop,
		}).Error("(WebhookCreateOrder) void transaction failed, unable to query Shopify transactions")
		return
	}

	// Find transaction where "kind":"sale" and return id
	var saleTransaction shopify.Transaction
	for _, t := range tr.Transactions {
		if t.Kind == "sale" {
			saleTransaction = t
			break
		}
	}

	if strconv.Itoa(saleTransaction.ID) == "" {
		logrus.WithFields(logrus.Fields{
			"orderId": orderId,
			"shop":    shop.Shop,
		}).Error("(WebhookCreateOrder) void transaction failed, no sale transaction found")
	} else {
		// Create a void transaction and POST to Shopify order
		voidRequest := &shopify.CreateTransactionRequest{
			Transaction: shopify.Transaction{
				Kind:     "void",
				Status:   "success",
				Currency: "USD",
				Gateway:  "Bread POS",
				ParentID: saleTransaction.ID,
				Test:     saleTransaction.Test,
				Message:  "",
			},
		}
		var voidResponse shopify.CreateTransactionResponse
		err = sc.CreateTransaction(orderId, voidRequest, &voidResponse)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error":   err.Error(),
				"orderId": orderId,
				"shop":    shop.Shop,
				"req":     voidRequest,
			}).Error("(WebhookCreateOrder) void transaction failed, unable to create void transaction")
		}
	}

	// Cancel Shopify order
	var cancelOrderResponse shopify.CreateOrderResponse
	err = sc.CancelOrder(orderId, &cancelOrderResponse)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":   err.Error(),
			"orderId": orderId,
			"shop":    shop.Shop,
		}).Error("(WebhookCreateOrder) cancel order failed, unable to mark order canceled")
	}

	// Add note to Shopify order
	updateRequest := &shopify.UpdateOrderRequest{
		Order: shopify.UpdateOrder{
			ID:   orderId,
			Note: orderNote,
		},
	}
	var updateOrderResponse shopify.CreateOrderResponse
	err = sc.UpdateOrder(orderId, updateRequest, &updateOrderResponse)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":   err.Error(),
			"orderId": orderId,
			"shop":    shop.Shop,
			"req":     updateRequest,
		}).Error("(WebhookCreateOrder) update order failed, unable to add void note to order")
	}
	return
}
