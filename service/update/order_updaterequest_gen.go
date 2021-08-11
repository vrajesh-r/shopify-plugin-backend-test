// Code generated by https://github.com/getbread/breadkit/zeus/tree/master/generators/updater. DO NOT EDIT.

package update

import (
	"encoding/json"
	"fmt"
	"strings"

	zeus "github.com/getbread/breadkit/zeus/types"
)

type OrderUpdateField int

const (
	OrderUpdate_UpdatedAt OrderUpdateField = iota
)

func (s OrderUpdateField) MarshalText() ([]byte, error) {

	var data string

	switch s {
	case OrderUpdate_UpdatedAt:
		data = "updatedAt"

	default:
		return nil, fmt.Errorf("Cannot marshal text '%v'", s)
	}

	return []byte(data), nil

}

func (s *OrderUpdateField) UnmarshalText(b []byte) error {
	str := strings.Trim(string(b), `"`)

	switch str {
	case "updatedAt":
		*s = OrderUpdate_UpdatedAt

	default:
		return fmt.Errorf("Cannot unmarshal text '%s'", str)
	}

	return nil

}

func (s OrderUpdateField) String() string {
	switch s {
	case OrderUpdate_UpdatedAt:
		return "updated_at"

	}
	return ""
}

type OrderUpdateRequest struct {
	Id      zeus.Uuid                        `json:"id"`
	Updates map[OrderUpdateField]interface{} `json:"updates"`
}

func (ur OrderUpdateRequest) MarshalText() ([]byte, error) {
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

func (ur OrderUpdateRequest) MarshalBinary() ([]byte, error) {
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

func (ur *OrderUpdateRequest) UnmarshalText(b []byte) error {
	var i map[string]interface{}

	err := json.Unmarshal(b, &i)

	if err != nil {
		return err
	}

	ur.Id = zeus.Uuid(i["id"].(string))

	var j = make(map[OrderUpdateField]interface{})

	for key, value := range i["updates"].(map[string]interface{}) {
		var p OrderUpdateField
		err = p.UnmarshalText([]byte(key))

		if err != nil {
			return err
		}
		j[p] = value
	}

	ur.Updates = j

	return nil
}

func (ur *OrderUpdateRequest) UnmarshalBinary(b []byte) error {
	var i map[string]interface{}

	err := json.Unmarshal(b, &i)

	if err != nil {
		return err
	}

	ur.Id = zeus.Uuid(i["id"].(string))

	var j = make(map[OrderUpdateField]interface{})

	for key, value := range i["updates"].(map[string]interface{}) {
		var p OrderUpdateField
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
func (ur *OrderUpdateRequest) GetId() zeus.Uuid {
	return ur.Id
}

func (ur *OrderUpdateRequest) GetTableName() string {
	return "shopify_shops_orders"
}

func (ur *OrderUpdateRequest) GetUpdates() map[string]interface{} {
	updates := make(map[string]interface{})

	for field, value := range ur.Updates {
		updates[field.String()] = value
	}

	return updates
}

func (ur *OrderUpdateRequest) AddUpdate(field OrderUpdateField, value interface{}) {
	if ur.Updates == nil {
		ur.Updates = make(map[OrderUpdateField]interface{})
	}

	ur.Updates[field] = value
}
