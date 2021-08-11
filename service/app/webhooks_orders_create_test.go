package app

import "testing"

func TestIsBreadGateway(t *testing.T) {
	validCases := []string{
		"bread",
		"bread_staging_",
		"bread_sandbox_",
		"bread_development_",
	}

	for _, test := range validCases {
		if !isBreadGateway(test) {
			t.Errorf("Expected %s to return true, but instead got false", test)
		}
	}
}

func TestIsBreadPOSOrder(t *testing.T) {
	validCases := []string{
		"Bread POS",
		"Bread  POS",
		"bread pos",
		"BREAD POS",
		"breadpos",
		"BREADPOS",
		"bread-pos",
		"BREAD-POS",
		"Bread - POS",
	}

	for _, test := range validCases {
		if !isBreadPOSOrder(test) {
			t.Errorf("Expected %s to return true, but instead got false", test)
		}
	}

	invalidCases := []string{
		"Brad POS",
		"bread",
		"bread shopify payments",
		"bread pos shopify",
		"bread pos gateway",
	}

	for _, test := range invalidCases {
		if isBreadPOSOrder(test) {
			t.Errorf("Expected %s to return false, but instead got true", test)
		}
	}
}
