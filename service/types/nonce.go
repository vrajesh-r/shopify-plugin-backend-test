package types

import (
	"math/rand"
	"strconv"
	"time"

	zeus "github.com/getbread/breadkit/zeus/types"
)

//go:generate go run github.com/getbread/breadkit/zeus/generators/searcher -type=Nonce -table=shopify_shops_nonces
//go:generate go run github.com/getbread/breadkit/zeus/generators/creator -type=Nonce -table=shopify_shops_nonces
//go:generate go run github.com/getbread/breadkit/zeus/generators/updater -type=Nonce -table=shopify_shops_nonces
type Nonce struct {
	Id        zeus.Uuid `json:"id" db:"id" zeus:"search"`
	ShopId    zeus.Uuid `json:"shop_id" db:"shop_id" zeus:"create,search"`
	Nonce     string    `json:"nonce" db:"nonce" zeus:"create,update,search"`
	CreatedAt time.Time `json:"created_at" db:"created_at" zeus:"search"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" zeus:"update,search"`
}

func GenerateNonceValue() string {
	return strconv.FormatInt(rand.Int63(), 16)
}
