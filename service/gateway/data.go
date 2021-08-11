package gateway

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/getbread/breadkit/featureflags"
	"github.com/getbread/breadkit/zeus/searcher"
	zeus "github.com/getbread/breadkit/zeus/types"
	"github.com/getbread/shopify_plugin_backend/service/bread"
	"github.com/getbread/shopify_plugin_backend/service/gateway/security"
	"github.com/getbread/shopify_plugin_backend/service/search"
	"github.com/getbread/shopify_plugin_backend/service/types"
	"github.com/getbread/shopify_plugin_backend/service/update"

	"github.com/sirupsen/logrus"
)

func findPlusGatewayCheckoutByCheckoutID(checkoutID string, h *Handlers) (types.PlusGatewayCheckout, error) {
	pgcsr := search.PlusGatewayCheckoutSearchRequest{}
	pgcsr.AddFilter(search.PlusGatewayCheckoutSearch_CheckoutID, checkoutID, searcher.Operator_EQ, searcher.Condition_AND)
	pgcsr.SetOrderBy(search.PlusGatewayCheckoutSearch_CreatedAt, zeus.IS_DESC)
	pgcsr.Limit = 1
	checkouts, err := h.PlusGatewayCheckoutSearcher.Search(pgcsr)
	if err != nil {
		return types.PlusGatewayCheckout{}, err
	}
	if len(checkouts) == 0 {
		return types.PlusGatewayCheckout{}, fmt.Errorf("PlusGatewayCheckout not found")
	}
	return checkouts[0], nil
}

func findGatewayCheckoutById(gcID zeus.Uuid, h *Handlers) (types.GatewayCheckout, error) {
	gcsr := search.GatewayCheckoutSearchRequest{}
	gcsr.AddFilter(search.GatewayCheckoutSearch_Id, gcID, searcher.Operator_EQ, searcher.Condition_AND)
	gcsr.Limit = 1
	checkouts, err := h.GatewayCheckoutSearcher.Search(gcsr)
	if err != nil {
		return types.GatewayCheckout{}, err
	}
	if len(checkouts) == 0 {
		return types.GatewayCheckout{}, fmt.Errorf("GatewayCheckout not found")
	}
	return checkouts[0], nil
}

func findGatewayCheckoutByTxId(breadTransactionId string, h *Handlers) (types.GatewayCheckout, error) {
	gcsr := search.GatewayCheckoutSearchRequest{}
	gcsr.AddFilter(search.GatewayCheckoutSearch_TransactionID, breadTransactionId, searcher.Operator_EQ, searcher.Condition_AND)
	gcsr.Limit = 1
	checkouts, err := h.GatewayCheckoutSearcher.Search(gcsr)
	if err != nil {
		return types.GatewayCheckout{}, err
	}
	if len(checkouts) == 0 {
		return types.GatewayCheckout{}, fmt.Errorf("GatewayCheckout not found")
	}
	return checkouts[0], nil
}

func completeGatewayCheckout(checkoutID zeus.Uuid, breadTransactionID, merchantID string, h *Handlers) error {
	gcur := update.GatewayCheckoutUpdateRequest{
		Id:      checkoutID,
		Updates: map[update.GatewayCheckoutUpdateField]interface{}{},
	}
	gcur.Updates[update.GatewayCheckoutUpdate_TransactionID] = breadTransactionID
	gcur.Updates[update.GatewayCheckoutUpdate_MerchantId] = merchantID
	gcur.Updates[update.GatewayCheckoutUpdate_Completed] = true
	return h.GatewayCheckoutUpdater.Update(gcur)
}

func queryTransaction(breadTransactionId string, account types.GatewayAccount, checkout types.GatewayCheckout) (*bread.TransactionResponse, error) {
	var apiKey string
	var sharedSecret string
	var host string
	if checkout.Test {
		apiKey = account.SandboxApiKey
		sharedSecret = account.SandboxSharedSecret
		host = gatewayConfig.HostConfig.BreadHostDevelopment
	} else {
		apiKey = account.ApiKey
		sharedSecret = account.SharedSecret
		host = gatewayConfig.HostConfig.BreadHost
	}
	bc := bread.NewClient(apiKey, sharedSecret)

	breadTx, err := bc.QueryTransaction(breadTransactionId, host)
	if err != nil {
		return nil, err
	}
	return breadTx, nil
}

func authorizeTransaction(breadTransactionId string, account types.GatewayAccount, checkout types.GatewayCheckout) error {
	var apiKey string
	var sharedSecret string
	var host string
	if checkout.Test {
		apiKey = account.SandboxApiKey
		sharedSecret = account.SandboxSharedSecret
		host = gatewayConfig.HostConfig.BreadHostDevelopment
	} else {
		apiKey = account.ApiKey
		sharedSecret = account.SharedSecret
		host = gatewayConfig.HostConfig.BreadHost
	}
	bc := bread.NewClient(apiKey, sharedSecret)

	authorizeRequest := &bread.TransactionActionRequest{
		Type: "authorize",
	}
	_, err := bc.AuthorizeTransaction(breadTransactionId, host, authorizeRequest)
	if err != nil {
		return err
	}
	return nil
}

func settleTransaction(breadTransactionId string, account types.GatewayAccount, checkout types.GatewayCheckout) error {
	var apiKey string
	var sharedSecret string
	var host string
	if checkout.Test {
		apiKey = account.SandboxApiKey
		sharedSecret = account.SandboxSharedSecret
		host = gatewayConfig.HostConfig.BreadHostDevelopment
	} else {
		apiKey = account.ApiKey
		sharedSecret = account.SharedSecret
		host = gatewayConfig.HostConfig.BreadHost
	}
	bc := bread.NewClient(apiKey, sharedSecret)
	settleRequest := &bread.TransactionActionRequest{
		Type: "settle",
	}
	_, err := bc.SettleTransaction(breadTransactionId, host, settleRequest)
	if err != nil {
		return err
	}
	return nil
}

func cancelTransaction(breadTransactionId string, account types.GatewayAccount, checkout types.GatewayCheckout, amount int) error {
	var apiKey string
	var sharedSecret string
	var host string
	if checkout.Test {
		apiKey = account.SandboxApiKey
		sharedSecret = account.SandboxSharedSecret
		host = gatewayConfig.HostConfig.BreadHostDevelopment
	} else {
		apiKey = account.ApiKey
		sharedSecret = account.SharedSecret
		host = gatewayConfig.HostConfig.BreadHost
	}
	bc := bread.NewClient(apiKey, sharedSecret)
	settleRequest := &bread.TransactionActionRequest{
		Type: "cancel",
	}
	if amount > 0 {
		settleRequest.Amount = amount
	}
	_, err := bc.CancelTransaction(breadTransactionId, host, settleRequest)
	if err != nil {
		return err
	}
	return nil
}

func refundTransaction(breadTransactionId string, account types.GatewayAccount, checkout types.GatewayCheckout, amount int) error {
	var apiKey string
	var sharedSecret string
	var host string
	if checkout.Test {
		apiKey = account.SandboxApiKey
		sharedSecret = account.SandboxSharedSecret
		host = gatewayConfig.HostConfig.BreadHostDevelopment
	} else {
		apiKey = account.ApiKey
		sharedSecret = account.SharedSecret
		host = gatewayConfig.HostConfig.BreadHost
	}
	bc := bread.NewClient(apiKey, sharedSecret)
	refundRequest := &bread.TransactionActionRequest{
		Type:   "refund",
		Amount: amount,
	}
	_, err := bc.RefundTransaction(breadTransactionId, host, refundRequest)
	if err != nil {
		return err
	}
	return nil
}

func postGatewayOrderResponse(req gatewayOrderManagementRequest, account types.GatewayAccount, result string, msg string) error {
	// Create response object
	res := gatewayOrderManagementResponse{
		AccountID:        req.AccountID,
		Amount:           req.Amount,
		Reference:        req.Reference,
		Currency:         req.Currency,
		GatewayReference: req.GatewayReference,
		Test:             req.Test,
		TransactionType:  req.TransactionType,
		Timestamp:        time.Now().UTC().Format(time.RFC3339),
		Result:           result,
		Message:          msg,
	}

	// Generate and assign response signature
	signatureInputs := map[string]string{
		"x_account_id":        res.AccountID,
		"x_reference":         strconv.Itoa(res.Reference),
		"x_currency":          res.Currency,
		"x_test":              strconv.FormatBool(res.Test),
		"x_amount":            res.Amount,
		"x_gateway_reference": res.GatewayReference,
		"x_timestamp":         res.Timestamp,
		"x_result":            string(res.Result),
		"x_transaction_type":  string(res.TransactionType),
		"x_message":           res.Message,
	}
	res.Signature = security.GenerateGatewaySignature(signatureInputs, account.GatewaySecret)

	// Form response object and POST request for x_url_callback
	form := url.Values{}
	form.Set("x_account_id", res.AccountID)
	form.Set("x_reference", strconv.Itoa(res.Reference))
	form.Set("x_currency", res.Currency)
	form.Set("x_test", strconv.FormatBool(res.Test))
	form.Set("x_amount", res.Amount)
	form.Set("x_gateway_reference", res.GatewayReference)
	form.Set("x_result", res.Result)
	form.Set("x_transaction_type", res.TransactionType)
	form.Set("x_message", res.Message)
	form.Set("x_timestamp", res.Timestamp)
	form.Set("x_signature", res.Signature)

	// POST response to Shopify callback url
	err := HTTPFormRequest("POST", req.CallbackURL, form, struct{}{})

	if featureflags.GetBool("milton-retry-callback", false) {
		attempts := 0
		// Retry callback up to 5 times at 60s interval
		for err != nil && attempts < 5 {
			time.Sleep(time.Second * 60)
			err = HTTPFormRequest("POST", req.CallbackURL, form, struct{}{})
			attempts++
		}

		// Log successful retries
		if err == nil && attempts > 0 {
			logrus.Infof("Shopify callback retry succeeded after %d attempts", attempts)
		}
	}

	return err
}
