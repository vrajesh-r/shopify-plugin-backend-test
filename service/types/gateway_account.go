package types

import (
	"time"

	zeus "github.com/getbread/breadkit/zeus/types"
	"github.com/getbread/shopify_plugin_backend/service/spb_config"
)

var typesConfig spb_config.ShopifyPluginBackendConfig

func TypesConfigInit(globalConfig spb_config.ShopifyPluginBackendConfig) {
	typesConfig = globalConfig
}

//go:generate go run github.com/getbread/breadkit/zeus/generators/searcher -type=GatewayAccount -table=shopify_gateway_accounts
//go:generate go run github.com/getbread/breadkit/zeus/generators/creator -type=GatewayAccount -table=shopify_gateway_accounts
//go:generate go run github.com/getbread/breadkit/zeus/generators/updater -type=GatewayAccount -table=shopify_gateway_accounts
type GatewayAccount struct {
	Id                          zeus.Uuid `json:"id" db:"id" zeus:"search"`
	Email                       string    `json:"email" db:"email" zeus:"create,search,update"`
	PasswordHash                string    `json:"_" db:"password_hash" zeus:"create,search,update"`
	GatewayKey                  string    `json:"gatewayKey" db:"gateway_key" zeus:"create,search,update"`
	GatewaySecret               string    `json:"gatewaySecret" db:"gateway_secret" zeus:"create,search,update"`
	ApiKey                      string    `json:"apiKey" db:"api_key" zeus:"create,search,update"`
	SharedSecret                string    `json:"sharedSecret" db:"shared_secret" zeus:"create,search,update"`
	SandboxApiKey               string    `json:"sandboxApiKey" db:"sandbox_api_key" zeus:"create,search,update"`
	SandboxSharedSecret         string    `json:"sandboxSharedSecret" db:"sandbox_shared_secret" zeus:"create,search,update"`
	AutoSettle                  bool      `json:"autoSettle" db:"auto_settle" zeus:"create,search,update"`
	HealthcareMode              bool      `json:"healthcareMode" db:"healthcare_mode" zeus:"create,search,update"`
	TargetedFinancing           bool      `json:"targetedFinancing" db:"targeted_financing" zeus:"update"`
	TargetedFinancingID         string    `json:"targetedFinancingID" db:"targeted_financing_id" zeus:"update"`
	TargetedFinancingThreshold  int64     `json:"targetedFinancingThreshold" db:"targeted_financing_threshold" zeus:"update"`
	PlusEmbeddedCheckout        bool      `json:"plusEmbeddedCheckout" db:"plus_embedded_checkout" zeus:"update"`
	Production                  bool      `json:"production" db:"production" zeus:"update"`
	CreatedAt                   time.Time `json:"createdAt" db:"created_at" zeus:"search"`
	UpdatedAt                   time.Time `json:"updatedAt" db:"updated_at" zeus:"search,update"`
	RemainderPayAutoCancel      bool      `json:"remainderPayAutoCancel" db:"remainder_pay_decline_auto_cancel" zeus:"update"`
	PlatformApiKey              string    `json:"platformApiKey" db:"api_key_v2" zeus:"create,search,update"`
	PlatformSharedSecret        string    `json:"platformSharedSecret" db:"shared_secret_v2" zeus:"create,search,update"`
	PlatformSandboxApiKey       string    `json:"platformSandboxApiKey" db:"sandbox_api_key_v2" zeus:"create,search,update"`
	PlatformSandboxSharedSecret string    `json:"platformSandboxSharedSecret" db:"sandbox_shared_secret_v2" zeus:"create,search,update"`
	IntegrationKey              string    `json:"integrationKey" db:"integration_key" zeus:"create,search,update"`
	SandboxIntegrationKey       string    `json:"sandboxIntegrationKey" db:"sandbox_integration_key" zeus:"create,search,update"`
	PlatformAutoSettle          bool      `json:"platformAutoSettle" db:"auto_settle_v2" zeus:"create,search,update"`
	ActiveVersion               string    `json:"activeVersion" db:"active_version" zeus:"create,search,update"`
}

func (a GatewayAccount) GetAPIKeys() (string, string) {
	if a.Production {
		return a.ApiKey, a.SharedSecret
	} else {
		return a.SandboxApiKey, a.SandboxSharedSecret
	}
}

func (a GatewayAccount) BreadHost() string {
	if a.Production {
		return typesConfig.HostConfig.BreadHost
	} else {
		return typesConfig.HostConfig.BreadHost
	}
}

func (a GatewayAccount) CheckoutHost() string {
	if a.Production {
		return typesConfig.HostConfig.CheckoutHost
	} else {
		return typesConfig.HostConfig.CheckoutHostDevelopment
	}
}
