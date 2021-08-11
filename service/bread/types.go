package bread

import (
	"time"

	zeus "github.com/getbread/breadkit/zeus/types"
	"github.com/getbread/shopify_plugin_backend/service/types"
)

type LineItem struct {
	BreadLineItemId string  `json:"breadLineItemId"`
	Price           int     `json:"price"`
	Product         Product `json:"product"`
	Quantity        int     `json:"quantity"`
}

type Product struct {
	Name      string `json:"name"`
	Sku       string `json:"sku"`
	ImageUrl  string `json:"imageUrl"`
	DetailUrl string `json:"detailUrl"`
}

type Contact struct {
	FullName  string `json:"fullName"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Address   string `json:"address"`
	Address2  string `json:"address2"`
	Zip       string `json:"zip"`
	City      string `json:"city"`
	State     string `json:"state"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
}

type ShippingAddress struct {
	FullName string `json:"fullName"`
	Address  string `json:"address"`
	Address2 string `json:"address2"`
	City     string `json:"city"`
	State    string `json:"state"`
	Zip      string `json:"zip"`
	Phone    string `json:"phone"`
}

type ShippingOption struct {
	Cost   int    `json:"cost"`
	Type   string `json:"type"`
	TypeID string `json:"typeId"`
}

type Item struct {
	ImageUrl  string `json:"imageUrl"`
	DetailUrl string `json:"detailUrl"`
	Name      string `json:"name"`
	Price     int    `json:"price"`
	Quantity  int    `json:"quantity"`
	Sku       string `json:"sku"`
}

type Cart struct {
	Id            zeus.Uuid   `json:"id,omitempty"`
	MerchantId    zeus.Uuid   `json:"merchantId,omitempty"`
	Expiration    string      `json:"expiration,omitempty"`
	Url           string      `json:"url"`
	PromoLegalese string      `json:"promoLegalese,omitempty"`
	Options       CartOptions `json:"options"`
	CartOrigin    string      `json:"cartOrigin"`
	CreatedAt     time.Time   `json:"createdAt"`
	UpdatedAt     time.Time   `json:"updatedAt"`
}

type CartOptions struct {
	ApiKey              string               `json:"apiKey"`                    // Populated using merchantId
	OrderRef            string               `json:"orderRef"`                  // Merchants [cart, order] identifier
	CallbackUrl         string               `json:"callbackUrl"`               // POST url for Bread checkout page on success
	CompleteUrl         string               `json:"completeUrl"`               // GET url for Bread checkout page on completion
	ErrorUrl            string               `json:"errorUrl"`                  // GET url for Bread checkout page on error
	ShippingOptions     []OptsShippingOption `json:"shippingOptions,omitempty"` // Standard checkout opts field
	Tax                 types.Cents          `json:"tax,omitempty"`             // Standard checkout opts field
	Items               []OptsItem           `json:"items,omitempty,omitempty"` // Standard checkout opts field
	CustomTotal         types.Cents          `json:"customTotal,omitempty"`     // Standard checkout opts field
	Discounts           []OptsDiscount       `json:"discounts,omitempty"`       // Standard checkout opts field
	ShippingContact     OptsContact          `json:"shippingContact,omitempty"` // Standard checkout opts field
	BillingContact      OptsContact          `json:"billingContact,omitempty"`  // Standard checkout opts field
	FinancingProgramID  string               `json:"financingProgramId"`
	DisableEditShipping bool                 `json:"disableEditShipping"`
}

type OptsShippingOption struct {
	Cost   types.Cents `json:"cost"`
	Type   string      `json:"type"`
	TypeID string      `json:"typeId"`
}

type OptsItem struct {
	ImageUrl  string      `json:"imageUrl"`
	DetailUrl string      `json:"detailUrl"`
	Name      string      `json:"name"`
	Price     types.Cents `json:"price"`
	Quantity  uint32      `json:"quantity"`
	Sku       string      `json:"sku"`
}

type OptsContact struct {
	FullName  string `json:"fullName,omitempty"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Email     string `json:"email,omitempty"`
	Address   string `json:"address,omitempty"`
	Address2  string `json:"address2,omitempty"`
	City      string `json:"city,omitempty"`
	State     string `json:"state,omitempty"`
	Zip       string `json:"zip,omitempty"`
	Phone     string `json:"phone,omitempty"`
}

type OptsDiscount struct {
	Amount      types.Cents `json:"amount"`
	Description string      `json:"description"`
}

type TrxAmount struct {
	Currency string      `json:"currency"`
	Value    types.Cents `json:"value"`
}

type TrxName struct {
	AdditionalName string `json:"additionalName"`
	FamilyName     string `json:"familyName"`
	GivenName      string `json:"givenName"`
}

type TrxAddress struct {
	Address1   string `json:"address1"`
	Address2   string `json:"address2"`
	Country    string `json:"country"`
	Locality   string `json:"locality"`
	PostalCode string `json:"postalCode"`
	Region     string `json:"region"`
}

type TrxContact struct {
	Address TrxAddress `json:"address"`
	Email   string     `json:"email"`
	Name    TrxName    `json:"name"`
	Phone   string     `json:"phone"`
}

type TrxErrorMetadata struct {
	AdditionalProp1 string `json:"additionalProp1"`
	AdditionalProp2 string `json:"additionalProp2"`
	AdditionalProp3 string `json:"additionalProp3"`
}

type PlatformSetupAddress struct {
	Address1   string `json:"address1"`
	Address2   string `json:"address2"`
	Country    string `json:"country"`
	Region     string `json:"region"`
	Locality   string `json:"locality"`
	PostalCode string `json:"postalCode"`
}

type PlatformSetupBuyer struct {
	GivenName       string               `json:"givenName"`
	FamilyName      string               `json:"familyName"`
	AdditionalName  string               `json:"additionalName,omitempty"`
	Email           string               `json:"email,omitempty"`
	Phone           string               `json:"phone,omitempty"`
	BillingAddress  PlatformSetupAddress `json:"billingAddress"`
	ShippingAddress PlatformSetupAddress `json:"shippingAddress"`
}

type PlatformSetup struct {
	IntegrationKey string             `json:"integrationKey"`
	LoyaltyID      string             `json:"loyaltyID,omitempty"`
	Buyer          PlatformSetupBuyer `json:"buyer"`
}

type PlatformOrderItem struct {
	Name                string    `json:"name"`
	Sku                 string    `json:"sku"`
	UnitPrice           TrxAmount `json:"unitPrice"`
	ShippingCost        TrxAmount `json:"shippingCost"`
	ShippingDescription string    `json:"shippingDescription"`
	UnitTax             TrxAmount `json:"unitTax"`
	Brand               string    `json:"brand,omitempty"`
	Currency            string    `json:"currency,omitempty"`
	Quantity            int       `json:"quantity,omitempty"`
}

type PlatformOrder struct {
	Items          []PlatformOrderItem `json:"items"`
	SubTotal       TrxAmount           `json:"subTotal"`
	TotalTax       TrxAmount           `json:"totalTax"`
	TotalShipping  TrxAmount           `json:"totalShipping"`
	TotalDiscounts TrxAmount           `json:"totalDiscounts"`
	TotalPrice     TrxAmount           `json:"totalPrice"`
}

type PlatformPlacement struct {
	FinancingType string        `json:"financingType,omitempty"`
	LocationType  string        `json:"locationType,omitempty"`
	DomID         string        `json:"domID"`
	AllowCheckout bool          `json:"allowCheckout,omitempty"`
	Order         PlatformOrder `json:"order,omitempty"`
}
