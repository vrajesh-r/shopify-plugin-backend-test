package types

import (
	"time"

	zeus "github.com/getbread/breadkit/zeus/types"
)

//go:generate go run github.com/getbread/breadkit/zeus/generators/searcher -type=Session -table=shopify_shops_sessions
//go:generate go run github.com/getbread/breadkit/zeus/generators/creator -type=Session -table=shopify_shops_sessions
//go:generate go run github.com/getbread/breadkit/zeus/generators/updater -type=Session -table=shopify_shops_sessions
type Session struct {
	Id         zeus.Uuid `json:"id" db:"id" zeus:"search"`
	ShopId     zeus.Uuid `json:"shop_id" db:"shop_id" zeus:"create,search"`
	Expiration int64     `json:"expiration" db:"expiration" zeus:"create,search"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at" zeus:"search"`
	UpdatedAt  time.Time `json:"updatedAt" db:"updated_at" zeus:"search,update"`
}

func GenerateSessionExpiration() int64 {
	life := int64(60 * 60) // expires in one hour
	return time.Now().Unix() + life
}
