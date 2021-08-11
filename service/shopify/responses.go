package shopify

type OAuthExchangeResponse struct {
	AccessToken string `json:"access_token"`
}

type DeleteWebhookResponse struct{}

type GetDraftOrdersCountResponse struct {
	Count int `json:"count"`
}

type GetDraftOrdersResponse struct {
	DraftOrders []DraftOrder `json:"draft_orders"`
}

type GetDraftOrderResponse struct {
	DraftOrder DraftOrder `json:"draft_order"`
}

type DeleteDraftOrderResponse struct {
	DraftOrder DraftOrder `json:"draft_order"`
}

type QueryWebhooksResponse struct {
	Webhooks []Webhook `json:"webhooks"`
}

type RegisterWebhookResponse struct {
	Webhook Webhook `json:"webhook"`
}

type SearchCustomerResponse struct {
	Customers []Customer `json:"customers"`
}

type SearchOrderResponse struct {
	Order Order `json:"order"`
}

type SearchTransactionsResponse struct {
	Transactions []Transaction `json:"transactions"`
}

type SearchTransactionResponse struct {
	Transaction Transaction `json:"transaction"`
}

type CreateCustomerResponse struct {
	Customer Customer `json:"customer"`
}

type CreateOrderResponse struct {
	Order Order `json:"order"`
}

type CreateTransactionResponse struct {
	Transaction Transaction `json:"transaction"`
}

type SearchProductByIdResponse struct {
	Product Product `json:"product"`
}

type SearchProductVariantByIdResponse struct {
	Variant ProductVariant `json:"variant"`
}

type CartResponse struct {
	Cart
}

type AddToCartResponse struct {
	CartItem
}

type ShippingRatesResponse struct {
	ShippingRates []ShippingRate `json:"shipping_rates"`
}

type EmbedScriptResponse struct {
	ScriptTag ScriptTag `json:"script_tag"`
}

type SearchLocationsResponse struct {
	Locations []Location `json:"locations"`
}

type SearchShopResponse struct {
	Shop Shop `json:"shop"`
}

type SearchEmbeddedScriptResponse struct {
	ScriptTags []ScriptTag `json:"script_tags"`
}

type DeleteEmbeddedScriptResponse struct{}
