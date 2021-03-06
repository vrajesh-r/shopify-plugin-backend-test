// Code generated by https://github.com/getbread/breadkit/zeus/tree/master/generators/updater. DO NOT EDIT.

package update

import (
	"encoding/json"
	"fmt"
	"strings"

	zeus "github.com/getbread/breadkit/zeus/types"
)

type GiftCardOrderUpdateField int

const (
	GiftCardOrderUpdate_UpdatedAt GiftCardOrderUpdateField = iota
)

func (s GiftCardOrderUpdateField) MarshalText() ([]byte, error) {

	var data string

	switch s {
	case GiftCardOrderUpdate_UpdatedAt:
		data = "updatedAt"

	default:
		return nil, fmt.Errorf("Cannot marshal text '%v'", s)
	}

	return []byte(data), nil

}

func (s *GiftCardOrderUpdateField) UnmarshalText(b []byte) error {
	str := strings.Trim(string(b), `"`)

	switch str {
	case "updatedAt":
		*s = GiftCardOrderUpdate_UpdatedAt

	default:
		return fmt.Errorf("Cannot unmarshal text '%s'", str)
	}

	return nil

}

func (s GiftCardOrderUpdateField) String() string {
	switch s {
	case GiftCardOrderUpdate_UpdatedAt:
		return "updated_at"

	}
	return ""
}

type GiftCardOrderUpdateRequest struct {
	Id      zeus.Uuid                                `json:"id"`
	Updates map[GiftCardOrderUpdateField]interface{} `json:"updates"`
}

func (ur GiftCardOrderUpdateRequest) MarshalText() ([]byte, error) {
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

func (ur GiftCardOrderUpdateRequest) MarshalBinary() ([]byte, error) {
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

func (ur *GiftCardOrderUpdateRequest) UnmarshalText(b []byte) error {
	var i map[string]interface{}

	err := json.Unmarshal(b, &i)

	if err != nil {
		return err
	}

	ur.Id = zeus.Uuid(i["id"].(string))

	var j = make(map[GiftCardOrderUpdateField]interface{})

	for key, value := range i["updates"].(map[string]interface{}) {
		var p GiftCardOrderUpdateField
		err = p.UnmarshalText([]byte(key))

		if err != nil {
			return err
		}
		j[p] = value
	}

	ur.Updates = j

	return nil
}

func (ur *GiftCardOrderUpdateRequest) UnmarshalBinary(b []byte) error {
	var i map[string]interface{}

	err := json.Unmarshal(b, &i)

	if err != nil {
		return err
	}

	ur.Id = zeus.Uuid(i["id"].(string))

	var j = make(map[GiftCardOrderUpdateField]interface{})

	for key, value := range i["updates"].(map[string]interface{}) {
		var p GiftCardOrderUpdateField
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
func (ur *GiftCardOrderUpdateRequest) GetId() zeus.Uuid {
	return ur.Id
}

func (ur *GiftCardOrderUpdateRequest) GetTableName() string {
	return "shopify_gift_card_orders"
}

func (ur *GiftCardOrderUpdateRequest) GetUpdates() map[string]interface{} {
	updates := make(map[string]interface{})

	for field, value := range ur.Updates {
		updates[field.String()] = value
	}

	return updates
}

func (ur *GiftCardOrderUpdateRequest) AddUpdate(field GiftCardOrderUpdateField, value interface{}) {
	if ur.Updates == nil {
		ur.Updates = make(map[GiftCardOrderUpdateField]interface{})
	}

	ur.Updates[field] = value
}
