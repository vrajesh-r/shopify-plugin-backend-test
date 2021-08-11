// Code generated by https://github.com/getbread/breadkit/zeus/tree/master/generators/updater. DO NOT EDIT.

package update

import (
	"encoding/json"
	"fmt"
	"strings"

	zeus "github.com/getbread/breadkit/zeus/types"
)

type DraftOrderCartCheckoutUpdateField int

const (
	DraftOrderCartCheckoutUpdate_UpdatedTime DraftOrderCartCheckoutUpdateField = iota
)

func (s DraftOrderCartCheckoutUpdateField) MarshalText() ([]byte, error) {

	var data string

	switch s {
	case DraftOrderCartCheckoutUpdate_UpdatedTime:
		data = "updatedAt"

	default:
		return nil, fmt.Errorf("Cannot marshal text '%v'", s)
	}

	return []byte(data), nil

}

func (s *DraftOrderCartCheckoutUpdateField) UnmarshalText(b []byte) error {
	str := strings.Trim(string(b), `"`)

	switch str {
	case "updatedAt":
		*s = DraftOrderCartCheckoutUpdate_UpdatedTime

	default:
		return fmt.Errorf("Cannot unmarshal text '%s'", str)
	}

	return nil

}

func (s DraftOrderCartCheckoutUpdateField) String() string {
	switch s {
	case DraftOrderCartCheckoutUpdate_UpdatedTime:
		return "updated_at"

	}
	return ""
}

type DraftOrderCartCheckoutUpdateRequest struct {
	Id      zeus.Uuid                                         `json:"id"`
	Updates map[DraftOrderCartCheckoutUpdateField]interface{} `json:"updates"`
}

func (ur DraftOrderCartCheckoutUpdateRequest) MarshalText() ([]byte, error) {
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

func (ur DraftOrderCartCheckoutUpdateRequest) MarshalBinary() ([]byte, error) {
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

func (ur *DraftOrderCartCheckoutUpdateRequest) UnmarshalText(b []byte) error {
	var i map[string]interface{}

	err := json.Unmarshal(b, &i)

	if err != nil {
		return err
	}

	ur.Id = zeus.Uuid(i["id"].(string))

	var j = make(map[DraftOrderCartCheckoutUpdateField]interface{})

	for key, value := range i["updates"].(map[string]interface{}) {
		var p DraftOrderCartCheckoutUpdateField
		err = p.UnmarshalText([]byte(key))

		if err != nil {
			return err
		}
		j[p] = value
	}

	ur.Updates = j

	return nil
}

func (ur *DraftOrderCartCheckoutUpdateRequest) UnmarshalBinary(b []byte) error {
	var i map[string]interface{}

	err := json.Unmarshal(b, &i)

	if err != nil {
		return err
	}

	ur.Id = zeus.Uuid(i["id"].(string))

	var j = make(map[DraftOrderCartCheckoutUpdateField]interface{})

	for key, value := range i["updates"].(map[string]interface{}) {
		var p DraftOrderCartCheckoutUpdateField
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
func (ur *DraftOrderCartCheckoutUpdateRequest) GetId() zeus.Uuid {
	return ur.Id
}

func (ur *DraftOrderCartCheckoutUpdateRequest) GetTableName() string {
	return "shopify_shops_draft_order_cart_checkouts"
}

func (ur *DraftOrderCartCheckoutUpdateRequest) GetUpdates() map[string]interface{} {
	updates := make(map[string]interface{})

	for field, value := range ur.Updates {
		updates[field.String()] = value
	}

	return updates
}

func (ur *DraftOrderCartCheckoutUpdateRequest) AddUpdate(field DraftOrderCartCheckoutUpdateField, value interface{}) {
	if ur.Updates == nil {
		ur.Updates = make(map[DraftOrderCartCheckoutUpdateField]interface{})
	}

	ur.Updates[field] = value
}