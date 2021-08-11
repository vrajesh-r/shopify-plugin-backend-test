package gateway

import (
	"encoding/json"
	"io/ioutil"
	"strconv"

	"github.com/getbread/shopify_plugin_backend/service/types"
	"github.com/gin-gonic/gin"
)

func processGatewayOrderManagementRequest(c *gin.Context, h *Handlers) (gatewayOrderManagementRequest, types.GatewayAccount, error) {
	var req gatewayOrderManagementRequest

	// Read request body
	bb, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		return req, types.GatewayAccount{}, err
	}

	// Unmarshal request
	err = json.Unmarshal(bb, &req)
	if err != nil {
		return req, types.GatewayAccount{}, err
	}

	// Query shop by BreadApiKey (x_account_id in request)
	account, err := findGatewayAccountByGatewayKey(req.AccountID, h)
	if err != nil {
		return req, types.GatewayAccount{}, err
	}
	return req, account, nil
}

func createRequestSignatureMap(req gatewayOrderManagementRequest) map[string]string {
	return map[string]string{
		"x_account_id":        req.AccountID,
		"x_amount":            req.Amount,
		"x_reference":         strconv.Itoa(req.Reference),
		"x_currency":          req.Currency,
		"x_gateway_reference": req.GatewayReference,
		"x_test":              strconv.FormatBool(req.Test),
		"x_url_callback":      req.CallbackURL,
		"x_shopify_order_id":  strconv.Itoa(req.ShopifyOrderID),
		"x_transaction_type":  string(req.TransactionType),
		"x_signature":         req.Signature,
	}
}
