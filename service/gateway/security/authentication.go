package security

import (
	"strings"
)

func GatewayRequestAuthentic(requestLookup map[string]string, secret, control string) bool {
	// Filter out un neccessary fields
	signatureInputs := map[string]string{}
	for key, _ := range requestLookup {
		if strings.HasPrefix(key, "x_") && key != "x_signature" {
			signatureInputs[key] = requestLookup[key]
		}
	}

	// Generate test
	test := GenerateGatewaySignature(signatureInputs, secret)

	// Test generated HMAC against request signature
	return test == control
}
