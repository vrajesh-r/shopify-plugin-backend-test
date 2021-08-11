package types

import (
	"time"

	zeus "github.com/getbread/breadkit/zeus/types"
)

//go:generate go run github.com/getbread/breadkit/zeus/generators/searcher -type=GatewaySession -table=shopify_gateway_sessions
//go:generate go run github.com/getbread/breadkit/zeus/generators/creator -type=GatewaySession -table=shopify_gateway_sessions
//go:generate go run github.com/getbread/breadkit/zeus/generators/updater -type=GatewaySession -table=shopify_gateway_sessions
type GatewaySession struct {
	Id               zeus.Uuid `json:"id" db:"id" zeus:"search"`
	GatewayAccountID zeus.Uuid `json:"gatewayAccountId" db:"gateway_account_id" zeus:"create,search"`
	Expiration       int64     `json:"expiration" db:"expiration" zeus:"create,search,update"`
	CreatedAt        time.Time `json:"createdAt" db:"created_at" zeus:"search"`
	UpdatedAt        time.Time `json:"updatedAt" db:"updated_at" zeus:"search,update"`
}

func GenerateGatewaySessionExpiration() int64 {
	life := int64(60 * 60) // 1 hour expiration
	return time.Now().Unix() + life
}
