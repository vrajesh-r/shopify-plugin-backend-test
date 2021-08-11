package gateway

import "github.com/getbread/shopify_plugin_backend/service/types"

type gatewayCheckoutRequest struct {
	AccountID                 string  `json:"x_account_id"`
	Currency                  string  `json:"x_currency"`
	Amount                    float64 `json:"x_amount"`
	Reference                 string  `json:"x_reference"`
	ShopCountry               string  `json:"x_shop_country"`
	ShopName                  string  `json:"x_shop_name"`
	TransactionType           string  `json:"x_transaction_type"`
	Description               string  `json:"x_description"`
	Invoice                   string  `json:"x_invoice"`
	Test                      bool    `json:"x_test"`
	CustomerFirstName         string  `json:"x_customer_first_name"`
	CustomerLastName          string  `json:"x_customer_last_name"`
	CustomerEmail             string  `json:"x_customer_email"`
	CustomerPhone             string  `json:"x_customer_phone"`
	CustomerBillingCity       string  `json:"x_customer_billing_city"`
	CustomerBillingCompany    string  `json:"x_customer_billing_company"`
	CustomerBillingAddress1   string  `json:"x_customer_billing_address1"`
	CustomerBillingAddress2   string  `json:"x_customer_billing_address2"`
	CustomerBillingState      string  `json:"x_customer_billing_state"`
	CustomerBillingZip        string  `json:"x_customer_billing_zip"`
	CustomerBillingCountry    string  `json:"x_customer_billing_country"`
	CustomerBillingPhone      string  `json:"x_customer_billing_phone"`
	CustomerShippingFirstName string  `json:"x_customer_shipping_first_name"`
	CustomerShippingLastName  string  `json:"x_customer_shipping_last_name"`
	CustomerShippingCity      string  `json:"x_customer_shipping_city"`
	CustomerShippingCompany   string  `json:"x_customer_shipping_company"`
	CustomerShippingAddress1  string  `json:"x_customer_shipping_address1"`
	CustomerShippingAddress2  string  `json:"x_customer_shipping_address2"`
	CustomerShippingState     string  `json:"x_customer_shipping_state"`
	CustomerShippingZIP       string  `json:"x_customer_shipping_zip"`
	CustomerShippingCountry   string  `json:"x_customer_shipping_country"`
	CustomerShippingPhone     string  `json:"x_customer_shipping_phone"`
	CallbackURL               string  `json:"x_url_callback"`
	CancelURL                 string  `json:"x_url_cancel"`
	CompleteURL               string  `json:"x_url_complete"`
	Timestamp                 string  `json:"x_timestamp"`
	Signature                 string  `json:"x_signature"`
}

type gatewayCheckoutCallbackRequest struct {
	OrderRef      string `json:"orderRef"`
	TransactionId string `json:"transactionId"`
}

type gatewayAccountSignUpRequest struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	PasswordVerify string `json:"passwordVerify"`
}

type gatewayAccountSignInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type updateGatewayAccountRequest struct {
	ApiKey                      string `json:"apiKey"`
	SharedSecret                string `json:"sharedSecret"`
	SandboxApiKey               string `json:"sandboxApiKey"`
	SandboxSharedSecret         string `json:"sandboxSharedSecret"`
	AutoSettle                  bool   `json:"autoSettle"`
	HealthcareMode              bool   `json:"healthcareMode"`
	TargetedFinancing           bool   `json:"targetedFinancing"`
	TargetedFinancingID         string `json:"targetedFinancingID"`
	TargetedFinancingThreshold  int64  `json:"targetedFinancingThreshold"`
	PlusEmbeddedCheckout        bool   `json:"plusEmbeddedCheckout"`
	Production                  bool   `json:"production"`
	RemainderPayAutoCancel      bool   `json:"remainderPayAutoCancel"`
	PlatformApiKey              string `json:"platformApiKey"`
	PlatformSharedSecret        string `json:"platformSharedSecret"`
	PlatformSandboxApiKey       string `json:"platformSandboxApiKey"`
	PlatformSandboxSharedSecret string `json:"platformSandboxSharedSecret"`
	PlatformAutoSettle          bool   `json:"platformAutoSettle"`
	IntegrationKey              string `json:"integrationKey"`
	SandboxIntegrationKey       string `json:"sandboxIntegrationKey"`
}

type updateGatewayVersionRequest struct {
	ActiveVersion string `json:"activeVersion"`
}

type updateGatewayAccountPasswordRequest struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

type resetPasswordRequest struct {
	ResetToken        string `json:"resetToken"`
	NewPassword       string `json:"newPassword"`
	NewPasswordVerify string `json:"newPasswordVerify"`
}

type gatewayOrderManagementRequest struct {
	AccountID        string `json:"x_account_id"`
	Amount           string `json:"x_amount"`
	Reference        int    `json:"x_reference"`
	Currency         string `json:"x_currency"`
	GatewayReference string `json:"x_gateway_reference"`
	Test             bool   `json:"x_test"`
	CallbackURL      string `json:"x_url_callback"`
	ShopifyOrderID   int    `json:"x_shopify_order_id"`
	TransactionType  string `json:"x_transaction_type"`
	Signature        string `json:"x_signature"`
}

type gatewayOrderManagementResponse struct {
	AccountID        string `json:"x_account_id"`
	Amount           string `json:"x_amount"`
	Reference        int    `json:"x_reference"`
	Currency         string `json:"x_currency"`
	GatewayReference string `json:"x_gateway_reference"`
	Test             bool   `json:"x_test"`
	TransactionType  string `json:"x_transaction_type"`
	Signature        string `json:"x_signature"`
	Timestamp        string `json:"x_timestamp"`
	Result           string `json:"x_result"`
	Message          string `json:"x_message"`
}

type PlatformCheckoutConfirmationRequest struct {
	TransactionID string      `json:"transactionID"`
	Amount        types.Cents `json:"amount"`
	MerchantID    string      `json:"merchantID"`
	checkoutID    string      `json:"checkoutID"`
}
