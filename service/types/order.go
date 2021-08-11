package types

import (
	"time"

	zeus "github.com/getbread/breadkit/zeus/types"
)

//go:generate go run github.com/getbread/breadkit/zeus/generators/searcher -type=Order -table=shopify_shops_orders
//go:generate go run github.com/getbread/breadkit/zeus/generators/creator -type=Order -table=shopify_shops_orders
//go:generate go run github.com/getbread/breadkit/zeus/generators/updater -type=Order -table=shopify_shops_orders
type Order struct {
	Id         zeus.Uuid `json:"id" db:"id" zeus:"search"`
	ShopId     zeus.Uuid `json:"shop_id" db:"shop_id" zeus:"create,search"`
	OrderId    int       `json:"order_id" db:"order_id" zeus:"create,search"`
	TxId       zeus.Uuid `json:"tx_id" db:"tx_id" zeus:"create,search"`
	Production bool      `json:"production" db:"production" zeus:"create,search"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at" zeus:"search"`
	UpdatedAt  time.Time `json:"updatedAt" db:"updated_at" zeus:"search,update"`
}

func (o Order) BreadHost() string {
	if o.Production {
		return typesConfig.HostConfig.BreadHost
	} else {
		return typesConfig.HostConfig.BreadHostDevelopment
	}
}
