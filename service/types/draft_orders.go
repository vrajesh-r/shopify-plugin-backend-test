package types

import (
	"time"

	"github.com/getbread/breadkit/zeus/types"
)

//go:generate go run github.com/getbread/breadkit/zeus/generators/searcher -type=DraftOrderCart -table=shopify_shops_draft_order_carts
//go:generate go run github.com/getbread/breadkit/zeus/generators/creator -type=DraftOrderCart -table=shopify_shops_draft_order_carts
//go:generate go run github.com/getbread/breadkit/zeus/generators/updater -type=DraftOrderCart -table=shopify_shops_draft_order_carts
type DraftOrderCart struct {
	ID                   types.Uuid `json:"id" db:"id" zeus:"create,search"`
	ShopID               types.Uuid `json:"shopId" db:"shop_id" zeus:"create,search"`
	DraftOrderID         int        `json:"draftOrderId" db:"draft_order_id" zeus:"create,search,update"`
	CartID               types.Uuid `json:"cartId" db:"cart_id" zeus:"create,search,update"`
	CartURL              string     `json:"cartUrl" db:"cart_url" zeus:"create,search,update"`
	IsProduction         bool       `json:"isProduction" db:"is_production" zeus:"create,search,update"`
	IsDeleted            bool       `json:"isDeleted" db:"is_deleted" zeus:"create,search,update"`
	UseDraftOrderAsOrder bool       `json:"useDraftOrderAsOrder" db:"use_draft_order_as_order" zeus:"search,update"`
	CreatedAt            time.Time  `json:"createdAt" db:"created_at" zeus:"search"`
	UpdatedAt            time.Time  `json:"updatedAt" db:"updated_at" zeus:"search,update"`
}

//go:generate go run github.com/getbread/breadkit/zeus/generators/searcher -type=DraftOrderCartCheckout -table=shopify_shops_draft_order_cart_checkouts
//go:generate go run github.com/getbread/breadkit/zeus/generators/creator -type=DraftOrderCartCheckout -table=shopify_shops_draft_order_cart_checkouts
//go:generate go run github.com/getbread/breadkit/zeus/generators/updater -type=DraftOrderCartCheckout -table=shopify_shops_draft_order_cart_checkouts
type DraftOrderCartCheckout struct {
	ID               types.Uuid `json:"id" db:"id" zeus:"search"`
	ShopID           types.Uuid `json:"shopId" db:"shop_id" zeus:"create,search"`
	TxID             types.Uuid `json:"txId" db:"tx_id" zeus:"create,search"`
	DraftOrderCartID types.Uuid `json:"draftOrderCartId" db:"draft_order_cart_id" zeus:"create,search"`
	OrderID          int        `json:"orderId" db:"order_id" zeus:"create,search"`
	IsProduction     bool       `json:"isProduction" db:"is_production" zeus:"create,search"`
	Completed        bool       `json:"completed" db:"completed" zeus:"create,search"`
	Errored          bool       `json:"errored" db:"errored" zeus:"create,search"`
	CreatedAt        time.Time  `json:"createdAt" db:"created_at" zeus:"search"`
	UpdatedTime      time.Time  `json:"updatedAt" db:"updated_at" zeus:"search,update"`
}
