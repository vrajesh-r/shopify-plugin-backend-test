package shopify

type OAuthExchangeRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code"`
}

type RegisterWebhookRequest struct {
	Webhook MiniWebhook `json:"webhook"`
}

type CreateCustomerRequest struct {
	Customer        Customer          `json:"customer"`
	SendEmailInvite bool              `json:"send_email_invite"`
	MetaFields      map[string]string `json:"meta_fields"`
}

type CreateOrderRequest struct {
	Order RequestOrder `json:"order"`
}

type UpdateOrderRequest struct {
	Order UpdateOrder `json:"order"`
}

type CreateDraftOrderRequest struct {
	DraftOrder DraftOrderRequest `json:"draft_order"`
}

type CreateTransactionRequest struct {
	Transaction Transaction `json:"transaction"`
}

type AddToCartRequest struct {
	Quantity int `json:"quantity"`
	Id       int `json:"id"`
}

type CartTaxCheckRequest struct {
	Method            string `url:"_method"`
	FirstName         string `url:"checkout[shipping_address][first_name]"`
	LastName          string `url:"checkout[shipping_address][last_name]"`
	Company           string `url:"checkout[shipping_address][company]"`
	Address1          string `url:"checkout[shipping_address][address1]"`
	Address2          string `url:"checkout[shipping_address][address2]"`
	City              string `url:"checkout[shipping_address][city]"`
	Country           string `url:"checkout[shipping_address][country]"`
	Province          string `url:"checkout[shipping_address][province]"`
	Zip               string `url:"checkout[shipping_address][zip]"`
	Phone             string `url:"checkout[shipping_address][phone]"`
	Email             string `url:"checkout[email]"`
	Step              string `url:"step"`
	PreviousStep      string `url:"previous_step"`
	AuthenticityToken string `url:"authenticity_token"`
}

type EmbedScriptRequest struct {
	ScriptTag ScriptTag `json:"script_tag"`
}
type GatewayCheckoutResult string

const (
	ResultComplete GatewayCheckoutResult = "completed"
	ResultFailed   GatewayCheckoutResult = "failed"
	ResultPending  GatewayCheckoutResult = "pending"
)

type GatewayCheckoutTxType string

const (
	TxTypeSale          GatewayCheckoutTxType = "sale"
	TxTypeAuthorization GatewayCheckoutTxType = "authorization"
)

type GatewayCheckoutCompleteRequest struct {
	AccountId        string                `json:"x_account_id"`
	Reference        string                `json:"x_reference"`
	Currency         string                `json:"x_currency"`
	Test             bool                  `json:"x_test"`
	Amount           float64               `json:"x_amount"`
	GatewayReference string                `json:"x_gateway_reference"`
	Timestamp        string                `json:"x_timestamp"`
	Result           GatewayCheckoutResult `json:"x_result"`
	Signature        string                `json:"x_signature"`
	TransactionType  GatewayCheckoutTxType `json:"x_transaction_type"`
	Message          string                `json:"x_message"`
}
