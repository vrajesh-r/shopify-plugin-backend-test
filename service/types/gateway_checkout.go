package types

import (
	"time"

	zeus "github.com/getbread/breadkit/zeus/types"
)

//go:generate go run github.com/getbread/breadkit/zeus/generators/searcher -type=GatewayCheckout -table=shopify_gateway_checkouts
//go:generate go run github.com/getbread/breadkit/zeus/generators/creator -type=GatewayCheckout -table=shopify_gateway_checkouts
//go:generate go run github.com/getbread/breadkit/zeus/generators/updater -type=GatewayCheckout -table=shopify_gateway_checkouts
type GatewayCheckout struct {
	Id            zeus.Uuid `json:"id" db:"id" zeus:"search"`
	AccountID     zeus.Uuid `json:"account_id" db:"account_id" zeus:"create,search"`
	TransactionID string    `json:"transaction_id" db:"transaction_id" zeus:"create,search,update"`
	Test          bool      `json:"test" db:"test" zeus:"create,search"`
	Reference     string    `json:"reference" db:"reference" zeus:"create,search"`
	Currency      string    `json:"currency" db:"currency" zeus:"create,search"`
	Amount        float64   `json:"amount" db:"amount" zeus:"create,search"`
	CallbackUrl   string    `json:"redirect_url" db:"callback_url" zeus:"create"`
	CompleteUrl   string    `json:"complete_url" db:"complete_url" zeus:"create"`
	CancelUrl     string    `json:"cancel_url" db:"cancel_url" zeus:"create"`
	Completed     bool      `json:"completed" db:"completed" zeus:"create,update,search"`
	Errored       bool      `json:"errored" db:"errored" zeus:"create,update,search"`
	CreatedAt     time.Time `json:"created_at" db:"created_at" zeus:"search"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at" zeus:"update,search"`
	AmountStr     string    `json:"amount_str" db:"amount_str" zeus:"create,search"`
	BreadVersion  string    `json:"bread_version" db:"bread_version" zeus:"create,search"`
	MerchantId    string    `json:"merchant_id" db:"merchant_id" zeus:"update,search"`
}
