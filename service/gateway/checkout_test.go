package gateway

import (
	"testing"
)

func TestStateInValidity(t *testing.T) {
	tests := []struct {
		BillingStateCode  string
		ShippingStateCode string
		Expected          bool
	}{
		//al and ar are valid state codes, aa is invalid
		{BillingStateCode: "aa", ShippingStateCode: "al", Expected: true},
		{BillingStateCode: "ar", ShippingStateCode: "aa", Expected: true},
		{BillingStateCode: "al", ShippingStateCode: "ar", Expected: false},
	}

	for _, test := range tests {
		actual := stateIsNotValid(test.BillingStateCode, test.ShippingStateCode)
		if actual != test.Expected {
			t.Errorf("Expected %v, got %v", test.Expected, actual)
		}
	}
}

/*func TestAuthAndSettleTransaction(t *testing.T) {
	bc := new(bmocks.IBreadClient)
	transactionID := "trxID"
	gatewayAccount := samples.NewGatewayAccount()
	host := "http://184.123.98.89"

	t.Run("Authorize but dont settle if auto-settling is false", func(tt *testing.T) {
		gatewayAccount.AutoSettle = false

		bc.On("AuthorizeTransaction", transactionID, host, mock.Anything).Return(mock.Anything, nil)
		bc.On("SetleTransaction", transactionID, host, mock.Anything).Return(mock.Anything, nil).Times(1)

		authAndSettleTransaction(bc, transactionID, gatewayAccount, host)
	})
}*/
