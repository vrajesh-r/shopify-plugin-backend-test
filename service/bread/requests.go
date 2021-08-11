package bread

type TransactionActionRequest struct {
	Type            string     `json:"type,omitempty"`
	Amount          int        `json:"amount,omitempty"`
	MerchantOrderId string     `json:"merchantOrderId,omitempty"`
	LineItems       []LineItem `json:"lineItems,omitempty"`
}

type CreateOffsiteCheckoutRequest struct {
	Redirect        bool             `json:"redirect,omitempty"`
	ApiKey          string           `json:"apiKey"`
	OrderRef        string           `json:"orderRef"`
	CallbackURL     string           `json:"callbackUrl"`
	CompleteURL     string           `json:"completeUrl"`
	ErrorURL        string           `json:"errorUrl"`
	ShippingType    string           `json:"shippingType,omitempty"`
	ShippingTypeID  string           `json:"shippingTypeId,omitempty"`
	ShippingCost    int              `json:"shippingCost,omitempty"`
	ShippingOptions []ShippingOption `json:"shippingOptions,omitempty"`
	ShippingContact ShippingAddress  `json:"shippingContact"`
	Tax             int              `json:"tax"`
	Items           []Item           `json:"items"`
	CustomTotal     int              `json:"customTotal"`
	Total           int              `json:"total"`
}

type SendCartEmailRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type SendCartTextRequest struct {
	Phone string `json:"phone"`
}

type TransactionShipmentRequest struct {
	CarrierName    string `json:"carrierName"`
	TrackingNumber string `json:"trackingNumber"`
}

type TrxRequest struct {
	Amount             TrxAmount `json:"amount"`
	MerchantOfRecordID string    `json:"merchantOfRecordID"`
}

type TrxAuthTokenRequest struct {
	ApiKey string `json:"apiKey"`
	Secret string `json:"secret"`
}

