package types

import (
	"time"

	zeus "github.com/getbread/breadkit/zeus/types"
)

//go:generate go run github.com/getbread/breadkit/zeus/generators/searcher -type=PlusGatewayCheckout -table=shopify_plus_gateway_checkouts
//go:generate go run github.com/getbread/breadkit/zeus/generators/creator -type=PlusGatewayCheckout -table=shopify_plus_gateway_checkouts
//go:generate go run github.com/getbread/breadkit/zeus/generators/updater -type=PlusGatewayCheckout -table=shopify_plus_gateway_checkouts
type PlusGatewayCheckout struct {
	Id            zeus.Uuid `json:"id" db:"id" zeus:"search"`
	CheckoutID    string    `json:"checkoutID" db:"checkout_id" zeus:"create,search"`
	TransactionID string    `json:"transactionID" db:"transaction_id" zeus:"create,search"`
	CreatedAt     time.Time `json:"created_at" db:"created_at" zeus:"search"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at" zeus:"update,search"`
}
