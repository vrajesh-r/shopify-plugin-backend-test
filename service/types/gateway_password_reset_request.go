package types

import (
	"time"

	zeus "github.com/getbread/breadkit/zeus/types"
)

//go:generate go run github.com/getbread/breadkit/zeus/generators/searcher -type=GatewayPasswordResetRequest -table=shopify_gateway_password_reset_requests
//go:generate go run github.com/getbread/breadkit/zeus/generators/creator -type=GatewayPasswordResetRequest -table=shopify_gateway_password_reset_requests
//go:generate go run github.com/getbread/breadkit/zeus/generators/updater -type=GatewayPasswordResetRequest -table=shopify_gateway_password_reset_requests
type GatewayPasswordResetRequest struct {
	ID         zeus.Uuid `json:"id" db:"id" zeus:"search"`
	AccountID  zeus.Uuid `json:"accountId" db:"account_id" zeus:"create,search"`
	TokenHash  string    `json:"_" db:"token_hash" zeus:"create,search,update"`
	Expiration int64     `json:"expiration" db:"expiration" zeus:"create,search,update"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at" zeus:"search"`
	UpdatedAt  time.Time `json:"updatedAt" db:"updated_at" zeus:"search,update"`
}

func GenerateResetTokenExpiration() int64 {
	life := int64(60 * 60) // 1 hour expiration
	return time.Now().Unix() + life
}
