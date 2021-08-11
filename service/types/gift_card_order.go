package types

import (
	"time"

	zeus "github.com/getbread/breadkit/zeus/types"
)

//go:generate go run github.com/getbread/breadkit/zeus/generators/searcher -type=GiftCardOrder -table=shopify_gift_card_orders
//go:generate go run github.com/getbread/breadkit/zeus/generators/creator -type=GiftCardOrder -table=shopify_gift_card_orders
//go:generate go run github.com/getbread/breadkit/zeus/generators/updater -type=GiftCardOrder -table=shopify_gift_card_orders
type GiftCardOrder struct {
	ID                   zeus.Uuid `json:"id" db:"id" zeus:"search"`
	OrderID              int64     `json:"orderID" db:"order_id" zeus:"create,search"`
	ShopName             string    `json:"shopName" db:"shop_name" zeus:"create,search"`
	Gateway              string    `json:"gateway" db:"gateway" zeus:"create,search"`
	Test                 bool      `json:"test" db:"test" zeus:"create,search"`
	ItemName             string    `json:"itemName" db:"item_name" zeus:"create,search"`
	ItemPrice            string    `json:"itemPrice" db:"item_price" zeus:"create,search"`
	Quantity             int64     `json:"quantity" db:"quantity" zeus:"create"`
	RequiresShipping     bool      `json:"requiresShipping" db:"requires_shipping" zeus:"create,search"`
	IsShopifyGiftCard    bool      `json:"isShopifyGiftCard" db:"is_shopify_gift_card" zeus:"create,search"`
	NameContainsGiftOnly bool      `json:"nameContainsGiftOnly" db:"name_contains_gift_only" zeus:"create,search"`
	NameContainsGiftCard bool      `json:"nameContainsGiftCard" db:"name_contains_gift_card" zeus:"create,search"`
	CreatedAt            time.Time `json:"createdAt" db:"created_at" zeus:"search"`
	UpdatedAt            time.Time `json:"updatedAt" db:"updated_at" zeus:"search,update"`
}
