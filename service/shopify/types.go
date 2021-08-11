package shopify

import (
	"time"

	"github.com/getbread/shopify_plugin_backend/service/types"
)

type DraftOrder struct {
	ID              int              `json:"id"`
	OrderID         int              `json:"order_id"`
	Name            string           `json:"name"`
	CustomerID      int              `json:"customer_id"`
	Customer        Customer         `json:"customer"`
	ShippingAddress Address          `json:"shipping_address"` // TODO: `Address` type is missing fields
	BillingAddress  Address          `json:"billing_address"`
	Note            string           `json:"note"`
	NoteAttributes  []NoteAttribute  `json:"node_attributes"`
	Email           string           `json:"email"`
	Currency        string           `json:"currency"`
	InvoiceSentAt   string           `json:"invoice_sent_at"`
	InvoiceUrl      string           `json:"invoice_url"`
	LineItems       []LineItem       `json:"line_items"`
	ShippingLine    ShippingLine     `json:"shipping_line"`
	Tags            string           `json:"tags"`
	TaxExempt       bool             `json:"tax_exempt"`
	TaxLines        []TaxLine        `json:"tax_lines"`
	AppliedDiscount AppliedDiscount  `json:"applied_discount"`
	TaxesIncluded   bool             `json:"taxes_included"`
	TotalTax        string           `json:"total_tax"`
	SubtotalPrice   string           `json:"subtotal_price"` // TODO: verify this property is a string
	TotalPrice      string           `json:"total_price"`
	CompletedAt     string           `json:"completed_at"`
	CreatedAt       string           `json:"created_at"`
	UpdatedAt       string           `json:"updated_at"`
	Status          DraftOrderStatus `json:"status"`
}

type NoteAttribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type AppliedDiscount struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Value       string    `json:"value"`
	ValueType   ValueType `json:"value_type"`
	Amount      string    `json:"amount"`
}

type ValueType string

const (
	ValueTypePercentage  ValueType = "percentage"
	ValueTypeFixedAmount ValueType = "fixed_amount"
)

type DraftOrderStatus string

const (
	DraftOrderStatusOpen        DraftOrderStatus = "open"
	DraftOrderStatusInvoiceSent DraftOrderStatus = "invoice_sent"
	DraftOrderStatusCompleted   DraftOrderStatus = "completed"
)

type Order struct {
	BuyerAcceptsMarketing bool                     `json:"buyer_accepts_marketing"`
	CancelReason          string                   `json:"cancel_reason"`
	CancelledAt           string                   `json:"cancelled_at"`
	CartToken             string                   `json:"cart_token"`
	CheckoutToken         string                   `json:"checkout_token"`
	ClosedAt              string                   `json:"closed_at"`
	Confirmed             bool                     `json:"confirmed"`
	CreatedAt             string                   `json:"created_at"`
	Currency              string                   `json:"currency"`
	DeviceID              int                      `json:"device_id"`
	Email                 string                   `json:"email"`
	FinancialStatus       string                   `json:"financial_status"`
	FulfillmentStatus     string                   `json:"fulfillment_status"`
	ID                    int                      `json:"id"`
	LandingSite           string                   `json:"landing_site"`
	LocationID            int                      `json:"location_id"`
	Name                  string                   `json:"name"`
	Note                  string                   `json:"note"`
	Number                int                      `json:"number"`
	ProcessedAt           string                   `json:"processed_at"`
	Reference             string                   `json:"reference"`
	ReferringSite         string                   `json:"referring_site"`
	SourceIdentifier      string                   `json:"source_identifier"`
	SourceUrl             string                   `json:"source_url"`
	SubtotalPrice         string                   `json:"subtotal_price"`
	TaxesIncluded         bool                     `json:"taxes_included"`
	Test                  bool                     `json:"test"`
	Token                 string                   `json:"token"`
	TotalDiscounts        string                   `json:"total_discounts"`
	TotalLineItemsPrice   string                   `json:"total_line_items_price"`
	TotalPrice            string                   `json:"total_price"`
	TotalPriceUSD         string                   `json:"total_price_usd"`
	TotalTax              string                   `json:"total_tax"`
	TotalWeight           int                      `json:"total_weight"`
	UpdatedAt             string                   `json:"updated_at"`
	UserID                int                      `json:"user_id"`
	BrowserIP             string                   `json:"browser_ip"`
	LandingSiteRef        string                   `json:"landing_site_ref"`
	OrderNumber           int                      `json:"order_number"`
	DiscountCodes         []DiscountCode           `json:"discount_codes"`
	NoteAttributes        []map[string]interface{} `json:"note_attributes"`
	ProcessingMethod      string                   `json:"processing_method"`
	Source                string                   `json:"source"`
	CheckoutID            int                      `json:"checkout_id"`
	SourceName            string                   `json:"source_name"`
	TaxLines              []TaxLine                `json:"tax_lines"`
	Tags                  string                   `json:"tags"`
	LineItems             []LineItem               `json:"line_items"`
	ShippingLines         []ShippingLine           `json:"shipping_lines"`
	BillingAddress        Address                  `json:"billing_address"`
	ShippingAddress       Address                  `json:"shipping_address"`
	Fulfillments          []Fulfillment            `json:"fulfillments"`
	Refunds               []Refund                 `json:"refunds"`
	PaymentDetails        map[string]interface{}   `json:"payment_details"`
	Customer              Customer                 `json:"customer"`
	ClientDetails         map[string]interface{}   `json:"client_details"`
	Gateway               string                   `json:"gateway"`
}

type DiscountCode struct {
	Code   string `json:"code"`
	Amount string `json:"amount"`
	Type   string `json:"type"`
}

type TaxLine struct {
	Price string  `json:"price"`
	Rate  float32 `json:"rate"`
	Title string  `json:"title"`
}

type LineItem struct {
	FulfillmentService         string                 `json:"fulfillment_service,omitempty"`
	FulfillmentStatus          string                 `json:"fulfillment_status,omitempty"`
	GiftCard                   bool                   `json:"gift_card,omitempty"`
	Grams                      int                    `json:"grams,omitempty"`
	ID                         int                    `json:"id,omitempty"`
	Price                      string                 `json:"price,omitempty"`
	ProductID                  int                    `json:"product_id,omitempty"`
	Quantity                   int                    `json:"quantity,omitempty"`
	RequiresShipping           bool                   `json:"requires_shipping,omitempty"`
	Sku                        string                 `json:"sku,omitempty"`
	Taxable                    bool                   `json:"taxable"`
	Title                      string                 `json:"title,omitempty"`
	VariantID                  int                    `json:"variant_id,omitempty"`
	VariantTitle               string                 `json:"variant_title,omitempty"`
	Vendor                     string                 `json:"vendor,omitempty"`
	Name                       string                 `json:"name,omitempty"`
	VariantInventoryManagement string                 `json:"variant_inventory_management,omitempty"`
	ProductExists              bool                   `json:"product_exists,omitempty"`
	FulfillableQuantity        int                    `json:"fulfillable_quantity,omitempty"`
	TotalDiscount              string                 `json:"total_discount,omitempty"`
	TaxLines                   []TaxLine              `json:"tax_lines,omitempty"`
	Fulfillments               []Fulfillment          `json:"fulfillments,omitempty"`
	ClientDetails              map[string]interface{} `json:"client_details,omitempty"`
}

type Refund struct {
	CreatedAt       string           `json:"created_at"`
	ID              int              `json:"id"`
	Note            string           `json:"note"`
	OrderID         int              `json:"order_id"`
	Restock         bool             `json:"refund"`
	UserID          int              `json:"user_id"`
	RefundLineItems []RefundLineItem `json:"refund_line_items"`
	Transactions    []Transaction    `json:"transactions"`
}

type RefundLineItem struct {
	ID                  int       `json:"id"`
	LineItemID          int       `json:"line_item_id"`
	Quantity            int       `json:"quantity"`
	LineItem            LineItem  `json:"line_item"`
	ProductExists       bool      `json:"product_exists"`
	FulfillableQuantity int       `json:"fulfillable_quantity"`
	TotalDiscount       string    `json:"total_discount"`
	TaxLines            []TaxLine `json:"tax_lines"`
}

type Transaction struct {
	Amount            string                 `json:"amount,omitempty"`
	Authorization     string                 `json:"authorization,omitempty"`
	CreatedAt         string                 `json:"created_at,omitempty"`
	Currency          string                 `json:"currency,omitempty"`
	Gateway           string                 `json:"gateway,omitempty"`
	ID                int                    `json:"id,omitempty"`
	Kind              string                 `json:"kind,omitempty"`
	LocationID        int                    `json:"location_id,omitempty"`
	Message           string                 `json:"message,omitempty"`
	OrderID           int                    `json:"order_id,omitempty"`
	ParentID          int                    `json:"parent_id,omitempty"`
	Status            string                 `json:"status,omitempty"`
	Test              bool                   `json:"test,omitempty"`
	UserID            int                    `json:"user_id,omitempty"`
	DeviceID          int                    `json:"device_id,omitempty"`
	Receipt           map[string]interface{} `json:"receipt,omitempty"`
	ErrorCode         string                 `json:"error_code,omitempty"`
	SourceName        string                 `json:"source_name,omitempty"`
	ProcessedAt       string                 `json:"processed_at,omitempty"`
	AdminGraphQLAPIID string                 `json:"admin_graphql_api_id,omitempty"`
}

type CustomShippingLine struct {
	Handle string  `json:"handle"`
	Price  float32 `json:"price"`
	Title  string  `json:"title"`
}

type ShippingLine struct {
	Code     string    `json:"code"`
	Price    string    `json:"price"`
	Source   string    `json:"source"`
	Title    string    `json:"title"`
	TaxLines []TaxLine `json:"tax_lines"`
}

type ShippingRate struct {
	Code         string `json:"code"`
	Title        string `json:"title"`
	Name         string `json:"name"`
	Price        string `json:"price"`
	DeliveryDate string `json:"delivery_date"`
	Source       string `json:"source"`
}

type Fulfillment struct {
	CreatedAt       string                 `json:"created_at"`
	ID              int                    `json:"id"`
	OrderID         int                    `json:"order_id"`
	Service         string                 `json:"service"`
	Status          string                 `json:"status"`
	TrackingCompany string                 `json:"tracking_company"`
	UpdatedAt       string                 `json:"updated_at"`
	TrackingNumber  string                 `json:"tracking_number"`
	TrackingNumbers []string               `json:"tracking_numbers"`
	TrackingUrl     string                 `json:"tracking_url"`
	TrackingUrls    []string               `json:"tracking_urls"`
	Receipt         map[string]interface{} `json:"receipt"`
	LineItems       []LineItem             `json:"line_items"`
}

type Customer struct {
	AcceptsMarketing     bool      `json:"accepts_marketing,omitempty"`
	CreatedAt            time.Time `json:"created_at,omitempty"`
	UpdatedAt            time.Time `json:"updated_at,omitempty"`
	Email                string    `json:"email,omitempty"`
	FirstName            string    `json:"first_name,omitempty"`
	ID                   int       `json:"id,omitempty"`
	LastName             string    `json:"last_name,omitempty"`
	LastOrderID          int       `json:"last_order_id,omitempty"`
	MultiPassIntentifier string    `json:"multipass_identifier,omitempty"`
	Note                 string    `json:"note,omitempty"`
	OrdersCount          int       `json:"orders_count,omitempty"`
	State                string    `json:"state,omitempty"`
	TaxExempt            bool      `json:"tax_exempt,omitempty"`
	TotalSpent           string    `json:"total_spent,omitempty"`
	LastOrderName        string    `json:"last_order_name,omitempty"`
	VerifiedEmail        bool      `json:"verified_email,omitempty"`
	Tags                 string    `json:"tags,omitempty"`
	DefaultAddress       Address   `json:"default_address,omitempty"`
	Addresses            []Address `json:"address,omitempty"`
}

type CustomerRef struct {
	ID int `json:"id"`
}

type Address struct {
	Address1     string `json:"address1"`
	Address2     string `json:"address2,omitempty"`
	City         string `json:"city"`
	Company      string `json:"company,omitempty"`
	Country      string `json:"country"`
	FirstName    string `json:"first_name"`
	ID           int    `json:"id,omitempty"`
	LastName     string `json:"last_name"`
	Phone        string `json:"phone"`
	Province     string `json:"province,omitempty"`
	Zip          string `json:"zip"`
	Name         string `json:"name"`
	ProvinceCode string `json:"province_code,omitempty"`
	CountryCode  string `json:"country_code,omitempty"`
	CountryName  string `json:"country_name,omitempty"`
	Default      bool   `json:"default,omitempty"`
}

type Product struct {
	BodyHTML       string                   `json:"body_html"`
	CreatedAt      string                   `json:"created_at"`
	Handle         string                   `json:"handle"`
	ID             int                      `json:"id"`
	Image          []ProductImage           `json:"images"`
	Options        []map[string]interface{} `json:"options"`
	ProductType    string                   `json:"product_type"`
	PublishedAt    string                   `json:"published_at"`
	PublishedScope string                   `json:"published_scope"`
	Tags           string                   `json:"tags"`
	TemplateSuffix string                   `json:"template_suffix"`
	Title          string                   `json:"title"`
	UpdatedAt      string                   `json:"updated_at"`
	Variants       []ProductVariant         `json:"variants"`
	Vendor         string                   `json:"vendor"`
}

type ProductVariant struct {
	Barcode              string  `json:"barcode"`
	CompareAtPrice       string  `json:"compare_at_price"`
	CreatedAt            string  `json:"created_at"`
	FulfillmentService   string  `json:"fulfillment_service"`
	Grams                int     `json:"grams"`
	ID                   int     `json:"id"`
	InventoryManagement  string  `json:"inventory_management"`
	InventoryPolicy      string  `json:"inventory_policy"`
	InventoryQuantity    int     `json:"inventory_quantity"`
	OldInventoryQuantity int     `json:"old_inventory_quantity"`
	Key                  string  `json:"key"`
	Value                int     `json:"value"`
	ValueType            string  `json:"value_type"`
	Namespace            string  `json:"namespace"`
	Option1              string  `json:"option1"`
	Option2              string  `json:"option2"`
	Option3              string  `json:"option3"`
	Position             int     `json:"position"`
	Price                string  `json:"price"`
	ProductID            int     `json:"product_id"`
	RequiresShipping     bool    `json:"required_shipping"`
	Sku                  string  `json:"sku"`
	Taxable              bool    `json:"taxable"`
	Title                string  `json:"title"`
	UpdatedAt            string  `json:"updated_at"`
	Weight               float32 `json:"weight"`
	WeightUnit           string  `json:"weight_unit"`
	ImageID              int     `json:"image_id"`
}

type ProductImage struct {
	CreatedAt  string `json:"created_at"`
	ID         int    `json:"id"`
	Position   int    `json:"position"`
	ProductID  int    `json:"product_id"`
	UpdatedAt  string `json:"updated_at"`
	Src        string `json:"src"`
	VariantIDs []int  `json:"variant_ids"`
}

type Webhook struct {
	Address        string   `json:"address"`
	Fields         []string `json:"fields"`
	Format         string   `json:"format"`
	Id             int      `json:"id"`
	MetaNamespaces []string `json:"metafield_namespaces"`
	Topic          string   `json:"topic"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
}

type MiniWebhook struct {
	Topic   string   `json:"topic"`
	Address string   `json:"address"`
	Format  string   `json:"format"`
	Fields  []string `json:"fields"`
}

type DraftOrderRequest struct {
	LineItems       []LineItem         `json:"line_items"`
	ShippingLine    CustomShippingLine `json:"shipping_line"`
	ShippingAddress Address            `json:"shipping_address"`
}

type RequestOrder struct {
	ShippingLines          []ShippingLine `json:"shipping_lines"`
	BillingAddress         Address        `json:"billing_address"`
	ShippingAddress        Address        `json:"shipping_address"`
	Email                  string         `json:"email"`
	TotalPrice             string         `json:"total_price"`
	TaxesIncluded          string         `json:"taxes_included,omitempty"`
	Transactions           []Transaction  `json:"transactions"`
	TaxLines               []TaxLine      `json:"tax_lines,omitempty"`
	TotalTax               string         `json:"total_tax,omitempty"`
	Currency               string         `json:"currency"`
	FinancialStatus        string         `json:"financial_status"`
	SendWebhooks           bool           `json:"send_webhooks"`
	SendReceipt            bool           `json:"send_receipt"`
	SendFulfillmentReceipt bool           `json:"send_fulfillment_receipt"`
	LineItems              []LineItem     `json:"line_items"`
	Customer               Customer       `json:"customer"`
	DiscountCodes          []DiscountCode `json:"discount_codes"`
	TotalDiscounts         string         `json:"total_discounts"`
	InventoryBehaviour     string         `json:"inventory_behaviour"`
}

type UpdateOrder struct {
	ID   int    `json:"id"`
	Note string `json:"note"`
}

type Cart struct {
	Token                       string                 `json:"token"`
	Note                        string                 `json:"note"`
	Attributes                  map[string]interface{} `json:"attributes"`
	TotalPrice                  int                    `json:"total_price"`
	TotalWeight                 int                    `json:"total_weight"`
	ItemCount                   int                    `json:"item_count"`
	Items                       []CartItem             `json:"items"`
	RequiresShipping            bool                   `json:"requires_shippings"`
	Currency                    string                 `json:"currency"`
	ItemsSubTotalPrice          types.Cents            `json:"items_subtotal_price"`
	CartLevelDiscoutApplication struct {
		TotalAllocatedAmount types.Cents `json:"total_allocated_amount"`
	}
}

type CartItem struct {
	Id                 int         `json:"id"`
	Title              string      `json:"title"`
	Price              int         `json:"price"`
	LinePrice          int         `json:"line_price"`
	Quantity           int         `json:"quantity"`
	Sku                string      `json:"sku"`
	Grams              int         `json:"grams"`
	Vendor             string      `json:"vendor"`
	Properties         string      `json:"properties"`
	VariantId          int         `json:"variant_id"`
	GiftCard           bool        `json:"gift_card"`
	Url                string      `json:"url"`
	Image              string      `json:"image"`
	Handle             string      `json:"handle"`
	RequiresShipping   bool        `json:"requies_shipping"`
	ProductTitle       string      `json:"product_title"`
	ProductDescription string      `json:"product_description"`
	ProductType        string      `json:"product_type"`
	VariantTitle       string      `json:"variant_title"`
	VariantOptions     []string    `json:"variant_options"`
	FinalPrice         types.Cents `json:"final_price"`
}

type ScriptTag struct {
	Event     string `json:"event,omitempty"`
	Id        int    `json:"id,omitempty"`
	Src       string `json:"src,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

type Location struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	LocationType string `json:"location_type"`
	Address1     string `json:"address1"`
	Address2     string `json:"address2"`
	Zip          string `json:"zip"`
	City         string `json:"city"`
	Province     string `json:"province"`
	Country      string `json:"country"`
	Phone        string `json:"phone"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type Shop struct {
	Address1                string  `json:"address1"`
	City                    string  `json:"city"`
	Country                 string  `json:"country"`
	CountryCode             string  `json:"country_code"`
	CountryName             string  `json:"country_name"`
	CreatedAt               string  `json:"created_at"`
	CustomerEmail           string  `json:"customer_email"`
	Currency                string  `json:"currency"`
	Description             string  `json:"description"`
	Domain                  string  `json:"domain"`
	Email                   string  `json:"email"`
	GoogleAppsDomain        string  `json:"google_apps_domain"`
	GoogleAppsLoginEnabled  string  `json:"google_apps_login_enabled"`
	Id                      int     `json:"id"`
	Latitude                float64 `json:"latitude"`
	Longitude               float64 `json:"longitude"`
	MoneyFormat             string  `json:"money_format"`
	MoneyWithCurrencyFormat string  `json:"money_with_currnecy_format"`
	MyshopifyAdmin          string  `json:"myshopify_admin"`
	Name                    string  `json:"name"`
	PlanName                string  `json:"plan_name"`
	PlanDisplayName         string  `json:"plan_display_name"`
	Phone                   string  `json:"phone"`
	PrimaryLocale           string  `json:"primary_locale"`
	Province                string  `json:"province"`
	ProvinceCode            string  `json:"province_code"`
	ShipsToCountries        string  `json:"ships_to_countries"`
	ShopOwner               string  `json:"shop_owner"`
	Source                  string  `json:"source"`
	TaxShipping             bool    `json:"tax_shipping"`
	Timezone                string  `json:"timezone"`
	IanaTimezone            string  `json:"iana_timezone"`
	Zip                     string  `json:"zip"`
	TaxesIncluded           bool    `json:"taxes_included"`
	CountyTaxes             bool    `json:"county_taxes"`
	PasswordEnabled         bool    `json:"password_enabled"`
	HasStorefront           bool    `json:"has_storefront"`
	SetupRequired           bool    `json:"setup_required"`
}
