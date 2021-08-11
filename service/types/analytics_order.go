package types

import (
	"time"

	zeus "github.com/getbread/breadkit/zeus/types"
)

//go:generate go run github.com/getbread/breadkit/zeus/generators/searcher -type=AnalyticsOrder -table=shopify_analytics_orders
//go:generate go run github.com/getbread/breadkit/zeus/generators/creator -type=AnalyticsOrder -table=shopify_analytics_orders
//go:generate go run github.com/getbread/breadkit/zeus/generators/updater -type=AnalyticsOrder -table=shopify_analytics_orders
type AnalyticsOrder struct {
	ID                zeus.Uuid       `json:"id" db:"id" zeus:"search"`
	ShopName          string          `json:"shopName" db:"shop_name" zeus:"create,search"`
	OrderID           int64           `json:"orderID" db:"order_id" zeus:"create,search"`
	CustomerID        int64           `json:"customerID" db:"customer_id" zeus:"create,search"`
	CustomerEmail     zeus.NullString `json:"customerEmail" db:"customer_email" zeus:"create,search,update"`
	TotalPrice        zeus.NullString `json:"totalPrice" db:"total_price" zeus:"create,search,update"`
	Gateway           zeus.NullString `json:"gateway" db:"gateway" zeus:"create,search,update"`
	FinancialStatus   zeus.NullString `json:"financialStatus" db:"financial_status" zeus:"create,search,update"`
	FulfillmentStatus zeus.NullString `json:"fulfillmentStatus" db:"fulfillment_status" zeus:"create,search,update"`
	Test              bool            `json:"test" db:"test" zeus:"create,search"`
	Redacted          bool            `json:"redacted" db:"redacted" zeus:"search,update"`
	CreatedAt         time.Time       `json:"createdAt" db:"created_at" zeus:"search"`
	UpdatedAt         time.Time       `json:"updatedAt" db:"updated_at" zeus:"search,update"`
	CheckoutID        zeus.NullInt64  `json:"checkoutID" db:"checkout_id" zeus:"create,search"`
	CheckoutToken     zeus.NullString `json:"checkoutToken" db:"checkout_token" zeus:"create,search"`
}
