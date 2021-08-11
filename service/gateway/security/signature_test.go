package security_test

import (
	"testing"

	"github.com/getbread/shopify_plugin_backend/service/gateway/security"
	"github.com/stretchr/testify/assert"
)

func TestGenerateGatewaySignature(t *testing.T) {
	values := map[string]string{
		"x_account_id":        "Z9s7Yt0Txsqbbx",
		"x_amount":            "89.99",
		"x_currency":          "USD",
		"x_gateway_reference": "123",
		"x_reference":         "19783",
		"x_result":            "completed",
		"x_test":              "true",
		"x_timestamp":         "2014-03-24T12:15:41Z",
	}
	secret := "iU44RWxeik"

	expectedSignature := "49d3166063b4d881b50af0b4648c1244bfa9890a53ed6bce6d2386404b610777"

	actualSignature := security.GenerateGatewaySignature(values, secret)
	assert.Equal(t, expectedSignature, actualSignature)
}
