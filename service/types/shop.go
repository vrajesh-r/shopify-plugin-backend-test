package types

import (
	"time"

	zeus "github.com/getbread/breadkit/zeus/types"
)

//go:generate go run github.com/getbread/breadkit/zeus/generators/searcher -type=Shop -table=shopify_shops
//go:generate go run github.com/getbread/breadkit/zeus/generators/creator -type=Shop -table=shopify_shops
//go:generate go run github.com/getbread/breadkit/zeus/generators/updater -type=Shop -table=shopify_shops
type Shop struct {
	Id                           zeus.Uuid `json:"id" db:"id" zeus:"search"`
	Shop                         string    `json:"shop" db:"shop" zeus:"create,search"`
	AccessToken                  string    `json:"accessToken" db:"access_token" zeus:"create,update"`
	BreadApiKey                  string    `json:"breadApiKey" db:"bread_api_key" zeus:"create,update,search"`
	BreadSecretKey               string    `json:"breadSecretKey" db:"bread_secret_key" zeus:"create,update"`
	BreadSandboxApiKey           string    `json:"breadSandboxApiKey" db:"bread_sandbox_api_key" zeus:"create,update,search"`
	BreadSandboxSecretKey        string    `json:"breadSandboxSecretKey" db:"bread_sandbox_secret_key" zeus:"create,update"`
	Production                   bool      `json:"production" db:"production" zeus:"update"`
	AutoAuthorize                bool      `json:"autoAuthorize" db:"auto_authorize" zeus:"update"`
	CreateCustomers              bool      `json:"createCustomers" db:"create_customers" zeus:"update"`
	AutoSettle                   bool      `json:"autoSettle" db:"auto_settle" zeus:"update"`
	ActsAsLabel                  bool      `json:"actsAsLabel" db:"acts_as_label" zeus:"update"`
	CSS                          string    `json:"css" db:"css" zeus:"create,update"`
	CSSCart                      string    `json:"cssCart" db:"css_cart" zeus:"create,update"`
	ManualEmbedScript            bool      `json:"manualEmbedScript" db:"manual_embed_script" zeus:"update"`
	AsLowAs                      bool      `json:"asLowAs" db:"as_low_as" zeus:"update"`
	EnableOrderWebhooks          bool      `json:"enableOrderWebhooks" db:"enable_order_webhooks" zeus:"update"` // Whether to enable order-based webhooks.
	CreatedAt                    time.Time `json:"createdAt" db:"created_at" zeus:"search"`
	UpdatedAt                    time.Time `json:"updatedAt" db:"updated_at" zeus:"search,update"`
	AllowCheckoutPDP             bool      `json:"allowCheckoutPDP" db:"allow_checkout_pdp" zeus:"update"`
	EnableAddToCart              bool      `json:"enableAddToCart" db:"enable_add_to_cart" zeus:"update"`
	AllowCheckoutCart            bool      `json:"allowCheckoutCart" db:"allow_checkout_cart" zeus:"update"`
	OAuthPermissionsUpToDate     bool      `json:"oauthPermissionsUpToDate" db:"oauth_permissions_up_to_date" zeus:"create,search,update"`
	HealthcareMode               bool      `json:"healthcareMode" db:"healthcare_mode" zeus:"update"`
	TargetedFinancing            bool      `json:"targetedFinancing" db:"targeted_financing" zeus:"update"`
	TargetedFinancingID          string    `json:"targetedFinancingID" db:"targeted_financing_id" zeus:"update"`
	TargetedFinancingThreshold   int64     `json:"targetedFinancingThreshold" db:"targeted_financing_threshold" zeus:"update"`
	DraftOrderTax                bool      `json:"draftOrderTax" db:"draft_order_tax" zeus:"update"`
	AcceleratedCheckoutPermitted bool      `json:"acceleratedCheckoutPermitted" db:"accelerated_checkout_permitted" zeus:"update"`
	POSAccess                    bool      `json:"posAccess" db:"pos_access" zeus:"update"`
	RemainderPayAutoCancel       bool      `json:"remainderPayAutoCancel" db:"remainder_pay_decline_auto_cancel" zeus:"update"`
	IntegrationKey               string    `json:"integrationKey" db:"integration_key" zeus:"create,update,search"`
	SandboxIntegrationKey        string    `json:"sandboxIntegrationKey" db:"sandbox_integration_key" zeus:"create,update"`
	PlatformProduction           bool      `json:"platformProduction" db:"production_v2" zeus:"update"`
	ActiveVersion                string    `json:"activeVersion" db:"active_version" zeus:"create,search,update"`
}

func (s Shop) GetAPIKeys() (string, string) {
	if s.Production {
		return s.BreadApiKey, s.BreadSecretKey
	} else {
		return s.BreadSandboxApiKey, s.BreadSandboxSecretKey
	}
}

func (s Shop) BreadHost() string {
	if s.Production {
		return typesConfig.HostConfig.BreadHost
	} else {
		return typesConfig.HostConfig.BreadHost
	}
}

func (s Shop) CheckoutHost() string {
	if s.Production {
		return typesConfig.HostConfig.CheckoutHost
	} else {
		return typesConfig.HostConfig.CheckoutHostDevelopment
	}
}

func (s Shop) PlatformCheckoutHost() string {
	if s.PlatformProduction {
		return typesConfig.HostConfig.PlatformCheckoutHost
	} else {
		return typesConfig.HostConfig.PlatformCheckoutHostDevelopment
	}
}

func (s Shop) GetIntegrationKey() string {
	if s.PlatformProduction {
		return s.IntegrationKey
	} else {
		return s.SandboxIntegrationKey
	}
}
