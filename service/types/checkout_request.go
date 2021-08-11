package types

type LineItem struct {
	ProductId int    `json:"product_id"`
	Title     string `json:"title"`
	LinePrice string `json:"line_price"`
	Price     string `json:"price"`
	Quantity  int    `json:"quantity"`
	Sku       string `json:"sku"`
}

type ShippingLine struct {
	Title string `json:"title"`
	Price string `json:"price"`
}

type CreateCheckoutRequest struct {
	Id                  int            `json:"id"`
	Token               string         `json:"token"`
	CartToken           string         `json:"cart_token"`
	LineItems           []LineItem     `json:"line_items"`
	PresentmentCurrency string         `json:"presentment_currency"` // USD
	TotalDiscounts      string         `json:"total_discounts"`
	TotalLineItemsPrice string         `json:"total_line_items_price"`
	TotalPrice          string         `json:"total_price"`
	TotalTax            string         `json:"total_tax"`
	SubtotalPrice       string         `json:"subtotal_price"`
	ShippingLines       []ShippingLine `json:"shipping_lines"`
}
