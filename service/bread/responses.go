package bread

type TransactionResponse struct {
	CreatedAt          string         `json:"createdAt"`
	BillingContact     Contact        `json:"billingContact"`
	ShippingContact    Contact        `json:"shippingContact"`
	BreadTransactionId string         `json:"breadTransactionId"`
	LineItems          []LineItem     `json:"lineItems"`
	MerchantOrderId    string         `json:"merchantOrderId"`
	ShippingCost       int            `json:"shippingCost"`
	ShippingMethodCode string         `json:"shippingMethodCode"`
	ShippingMethodName string         `json:"shippingMethodName"`
	Status             string         `json:"status"`
	Total              int            `json:"total"`
	AdjustedTotal      int            `json:"adjustedTotal"`
	TotalTax           int            `json:"totalTax"`
	Discounts          []OptsDiscount `json:"discounts"`
}

type CreateOffsiteCheckoutResponse struct {
	RedirectUrl string `json:"redirectUrl"`
}

type MessageResponse struct {
	Msg string `json:"message"`
}

type TrxResponse struct {
	AdjustedAmount            TrxAmount  `json:"adjustedAmount"`
	ApplicationID             string     `json:"applicationID"`
	BillingContact            TrxContact `json:"billingContact"`
	BuyerID                   string     `json:"buyerID"`
	CreatedAt                 string     `json:"createdAt"`
	Description               string     `json:"description"`
	DiscountAmount            TrxAmount  `json:"discountAmount"`
	ExternalID                string     `json:"externalID"`
	FulfillmentCarrier        string     `json:"fulfillmentCarrier"`
	FulfillmentTrackingNumber string     `json:"fulfillmentTrackingNumber"`
	ID                        string     `json:"id"`
	MerchantID                string     `json:"merchantID"`
	Nonce                     string     `json:"nonce"`
	PaymentAgreementID        string     `json:"paymentAgreementID"`
	ProductType               string     `json:"productType"`
	ProgramID                 string     `json:"programID"`
	SettledAmount             TrxAmount  `json:"settledAmount"`
	ShippingAmount            TrxAmount  `json:"shippingAmount"`
	ShippingContact           TrxContact `json:"shippingContact"`
	Status                    string     `json:"status"`
	TaxAmount                 TrxAmount  `json:"taxAmount"`
	TenantID                  string     `json:"tenantID"`
	TotalAmount               TrxAmount  `json:"totalAmount"`
}

type TrxErrorResponse struct {
	Domain   string           `json:"domain"`
	Message  string           `json:"message"`
	Metadata TrxErrorMetadata `json:"metadata"`
	Reason   string           `json:"reason"`
}

type TrxAuthTokenResponse struct {
	Token string `json:"token"`
}
