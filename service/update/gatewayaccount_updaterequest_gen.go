// Code generated by https://github.com/getbread/breadkit/zeus/tree/master/generators/updater. DO NOT EDIT.

package update

import (
	"encoding/json"
	"fmt"
	"strings"

	zeus "github.com/getbread/breadkit/zeus/types"
)

type GatewayAccountUpdateField int

const (
	GatewayAccountUpdate_Email GatewayAccountUpdateField = iota
	GatewayAccountUpdate_PasswordHash
	GatewayAccountUpdate_GatewayKey
	GatewayAccountUpdate_GatewaySecret
	GatewayAccountUpdate_ApiKey
	GatewayAccountUpdate_SharedSecret
	GatewayAccountUpdate_SandboxApiKey
	GatewayAccountUpdate_SandboxSharedSecret
	GatewayAccountUpdate_AutoSettle
	GatewayAccountUpdate_HealthcareMode
	GatewayAccountUpdate_TargetedFinancing
	GatewayAccountUpdate_TargetedFinancingID
	GatewayAccountUpdate_TargetedFinancingThreshold
	GatewayAccountUpdate_PlusEmbeddedCheckout
	GatewayAccountUpdate_Production
	GatewayAccountUpdate_UpdatedAt
	GatewayAccountUpdate_RemainderPayAutoCancel
	GatewayAccountUpdate_PlatformApiKey
	GatewayAccountUpdate_PlatformSharedSecret
	GatewayAccountUpdate_PlatformSandboxApiKey
	GatewayAccountUpdate_PlatformSandboxSharedSecret
	GatewayAccountUpdate_PlatformAutoSettle
	GatewayAccountUpdate_ActiveVersion
	GatewayAccountUpdate_IntegrationKey
	GatewayAccountUpdate_SandboxIntegrationKey
)

func (s GatewayAccountUpdateField) MarshalText() ([]byte, error) {

	var data string

	switch s {
	case GatewayAccountUpdate_Email:
		data = "email"
	case GatewayAccountUpdate_PasswordHash:
		data = "_"
	case GatewayAccountUpdate_GatewayKey:
		data = "gatewayKey"
	case GatewayAccountUpdate_GatewaySecret:
		data = "gatewaySecret"
	case GatewayAccountUpdate_ApiKey:
		data = "apiKey"
	case GatewayAccountUpdate_SharedSecret:
		data = "sharedSecret"
	case GatewayAccountUpdate_SandboxApiKey:
		data = "sandboxApiKey"
	case GatewayAccountUpdate_SandboxSharedSecret:
		data = "sandboxSharedSecret"
	case GatewayAccountUpdate_AutoSettle:
		data = "autoSettle"
	case GatewayAccountUpdate_HealthcareMode:
		data = "healthcareMode"
	case GatewayAccountUpdate_TargetedFinancing:
		data = "targetedFinancing"
	case GatewayAccountUpdate_TargetedFinancingID:
		data = "targetedFinancingID"
	case GatewayAccountUpdate_TargetedFinancingThreshold:
		data = "targetedFinancingThreshold"
	case GatewayAccountUpdate_PlusEmbeddedCheckout:
		data = "plusEmbeddedCheckout"
	case GatewayAccountUpdate_Production:
		data = "production"
	case GatewayAccountUpdate_UpdatedAt:
		data = "updatedAt"
	case GatewayAccountUpdate_RemainderPayAutoCancel:
		data = "remainderPayAutoCancel"
	case GatewayAccountUpdate_PlatformApiKey:
		data = "platformApiKey"
	case GatewayAccountUpdate_PlatformSharedSecret:
		data = "platformSharedSecret"
	case GatewayAccountUpdate_PlatformSandboxApiKey:
		data = "platformSandboxApiKey"
	case GatewayAccountUpdate_PlatformSandboxSharedSecret:
		data = "platformSandboxSharedSecret"
	case GatewayAccountUpdate_PlatformAutoSettle:
		data = "platformAutoSettle"
	case GatewayAccountUpdate_ActiveVersion:
		data = "activeVersion"
	case GatewayAccountUpdate_IntegrationKey:
		data = "integrationKey"
	case GatewayAccountUpdate_SandboxIntegrationKey:
		data = "sandboxIntegrationKey"

	default:
		return nil, fmt.Errorf("Cannot marshal text '%v'", s)
	}

	return []byte(data), nil

}

func (s *GatewayAccountUpdateField) UnmarshalText(b []byte) error {
	str := strings.Trim(string(b), `"`)

	switch str {
	case "email":
		*s = GatewayAccountUpdate_Email
	case "_":
		*s = GatewayAccountUpdate_PasswordHash
	case "gatewayKey":
		*s = GatewayAccountUpdate_GatewayKey
	case "gatewaySecret":
		*s = GatewayAccountUpdate_GatewaySecret
	case "apiKey":
		*s = GatewayAccountUpdate_ApiKey
	case "sharedSecret":
		*s = GatewayAccountUpdate_SharedSecret
	case "sandboxApiKey":
		*s = GatewayAccountUpdate_SandboxApiKey
	case "sandboxSharedSecret":
		*s = GatewayAccountUpdate_SandboxSharedSecret
	case "autoSettle":
		*s = GatewayAccountUpdate_AutoSettle
	case "healthcareMode":
		*s = GatewayAccountUpdate_HealthcareMode
	case "targetedFinancing":
		*s = GatewayAccountUpdate_TargetedFinancing
	case "targetedFinancingID":
		*s = GatewayAccountUpdate_TargetedFinancingID
	case "targetedFinancingThreshold":
		*s = GatewayAccountUpdate_TargetedFinancingThreshold
	case "plusEmbeddedCheckout":
		*s = GatewayAccountUpdate_PlusEmbeddedCheckout
	case "production":
		*s = GatewayAccountUpdate_Production
	case "updatedAt":
		*s = GatewayAccountUpdate_UpdatedAt
	case "remainderPayAutoCancel":
		*s = GatewayAccountUpdate_RemainderPayAutoCancel
	case "platformApiKey":
		*s = GatewayAccountUpdate_PlatformApiKey
	case "platformSharedSecret":
		*s = GatewayAccountUpdate_PlatformSharedSecret
	case "platformSandboxApiKey":
		*s = GatewayAccountUpdate_PlatformSandboxApiKey
	case "platformSandboxSharedSecret":
		*s = GatewayAccountUpdate_PlatformSandboxSharedSecret
	case "platformAutoSettle":
		*s = GatewayAccountUpdate_PlatformAutoSettle
	case "activeVersion":
		*s = GatewayAccountUpdate_ActiveVersion
	case "integrationKey":
		*s = GatewayAccountUpdate_IntegrationKey
	case "sandboxIntegrationKey":
		*s = GatewayAccountUpdate_SandboxIntegrationKey

	default:
		return fmt.Errorf("Cannot unmarshal text '%s'", str)
	}

	return nil

}

func (s GatewayAccountUpdateField) String() string {
	switch s {
	case GatewayAccountUpdate_Email:
		return "email"
	case GatewayAccountUpdate_PasswordHash:
		return "password_hash"
	case GatewayAccountUpdate_GatewayKey:
		return "gateway_key"
	case GatewayAccountUpdate_GatewaySecret:
		return "gateway_secret"
	case GatewayAccountUpdate_ApiKey:
		return "api_key"
	case GatewayAccountUpdate_SharedSecret:
		return "shared_secret"
	case GatewayAccountUpdate_SandboxApiKey:
		return "sandbox_api_key"
	case GatewayAccountUpdate_SandboxSharedSecret:
		return "sandbox_shared_secret"
	case GatewayAccountUpdate_AutoSettle:
		return "auto_settle"
	case GatewayAccountUpdate_HealthcareMode:
		return "healthcare_mode"
	case GatewayAccountUpdate_TargetedFinancing:
		return "targeted_financing"
	case GatewayAccountUpdate_TargetedFinancingID:
		return "targeted_financing_id"
	case GatewayAccountUpdate_TargetedFinancingThreshold:
		return "targeted_financing_threshold"
	case GatewayAccountUpdate_PlusEmbeddedCheckout:
		return "plus_embedded_checkout"
	case GatewayAccountUpdate_Production:
		return "production"
	case GatewayAccountUpdate_UpdatedAt:
		return "updated_at"
	case GatewayAccountUpdate_RemainderPayAutoCancel:
		return "remainder_pay_decline_auto_cancel"
	case GatewayAccountUpdate_PlatformApiKey:
		return "api_key_v2"
	case GatewayAccountUpdate_PlatformSharedSecret:
		return "shared_secret_v2"
	case GatewayAccountUpdate_PlatformSandboxApiKey:
		return "sandbox_api_key_v2"
	case GatewayAccountUpdate_PlatformSandboxSharedSecret:
		return "sandbox_shared_secret_v2"
	case GatewayAccountUpdate_PlatformAutoSettle:
		return "auto_settle_v2"
	case GatewayAccountUpdate_ActiveVersion:
		return "active_version"
	case GatewayAccountUpdate_IntegrationKey:
		return "integration_key"
	case GatewayAccountUpdate_SandboxIntegrationKey:
		return "sandbox_integration_key"

	}
	return ""
}

type GatewayAccountUpdateRequest struct {
	Id      zeus.Uuid                                 `json:"id"`
	Updates map[GatewayAccountUpdateField]interface{} `json:"updates"`
}

func (ur GatewayAccountUpdateRequest) MarshalText() ([]byte, error) {
	stringified := make(map[string]interface{})

	for key, value := range ur.Updates {
		s, err := key.MarshalText()

		if err != nil {
			return nil, err
		}

		stringified[string(s)] = value
	}

	result := map[string]interface{}{
		"updates": stringified,
		"id":      ur.Id,
	}

	return json.Marshal(result)
}

func (ur GatewayAccountUpdateRequest) MarshalBinary() ([]byte, error) {
	stringified := make(map[string]interface{})

	for key, value := range ur.Updates {
		s, err := key.MarshalText()

		if err != nil {
			return nil, err
		}

		stringified[string(s)] = value
	}

	result := map[string]interface{}{
		"updates": stringified,
		"id":      ur.Id,
	}

	return json.Marshal(result)
}

func (ur *GatewayAccountUpdateRequest) UnmarshalText(b []byte) error {
	var i map[string]interface{}

	err := json.Unmarshal(b, &i)

	if err != nil {
		return err
	}

	ur.Id = zeus.Uuid(i["id"].(string))

	var j = make(map[GatewayAccountUpdateField]interface{})

	for key, value := range i["updates"].(map[string]interface{}) {
		var p GatewayAccountUpdateField
		err = p.UnmarshalText([]byte(key))

		if err != nil {
			return err
		}
		j[p] = value
	}

	ur.Updates = j

	return nil
}

func (ur *GatewayAccountUpdateRequest) UnmarshalBinary(b []byte) error {
	var i map[string]interface{}

	err := json.Unmarshal(b, &i)

	if err != nil {
		return err
	}

	ur.Id = zeus.Uuid(i["id"].(string))

	var j = make(map[GatewayAccountUpdateField]interface{})

	for key, value := range i["updates"].(map[string]interface{}) {
		var p GatewayAccountUpdateField
		err = p.UnmarshalText([]byte(key))

		if err != nil {
			return err
		}
		j[p] = value
	}

	ur.Updates = j

	return nil
}

// implement updater.UpdateRequest interface
func (ur *GatewayAccountUpdateRequest) GetId() zeus.Uuid {
	return ur.Id
}

func (ur *GatewayAccountUpdateRequest) GetTableName() string {
	return "shopify_gateway_accounts"
}

func (ur *GatewayAccountUpdateRequest) GetUpdates() map[string]interface{} {
	updates := make(map[string]interface{})

	for field, value := range ur.Updates {
		updates[field.String()] = value
	}

	return updates
}

func (ur *GatewayAccountUpdateRequest) AddUpdate(field GatewayAccountUpdateField, value interface{}) {
	if ur.Updates == nil {
		ur.Updates = make(map[GatewayAccountUpdateField]interface{})
	}

	ur.Updates[field] = value
}
