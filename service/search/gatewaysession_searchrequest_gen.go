// Code generated by https://github.com/getbread/breadkit/zeus/tree/master/generators/searcher. DO NOT EDIT.

package search

import (
	"fmt"
	"strings"

	"github.com/getbread/breadkit/zeus/searcher"

	zeus "github.com/getbread/breadkit/zeus/types"
)

type GatewaySessionSearchField int

const (
	GatewaySessionSearch_Id GatewaySessionSearchField = iota
	GatewaySessionSearch_GatewayAccountID
	GatewaySessionSearch_Expiration
	GatewaySessionSearch_CreatedAt
	GatewaySessionSearch_UpdatedAt
)

func (s GatewaySessionSearchField) MarshalText() ([]byte, error) {
	var data string

	switch s {
	case GatewaySessionSearch_Id:
		data = "id"
	case GatewaySessionSearch_GatewayAccountID:
		data = "gatewayAccountId"
	case GatewaySessionSearch_Expiration:
		data = "expiration"
	case GatewaySessionSearch_CreatedAt:
		data = "createdAt"
	case GatewaySessionSearch_UpdatedAt:
		data = "updatedAt"

	default:
		return nil, fmt.Errorf("Cannot marshal text '%v'", s)
	}
	return []byte(data), nil
}

func (s GatewaySessionSearchField) MarshalBinary() ([]byte, error) {
	var data string

	switch s {
	case GatewaySessionSearch_Id:
		data = "id"
	case GatewaySessionSearch_GatewayAccountID:
		data = "gatewayAccountId"
	case GatewaySessionSearch_Expiration:
		data = "expiration"
	case GatewaySessionSearch_CreatedAt:
		data = "createdAt"
	case GatewaySessionSearch_UpdatedAt:
		data = "updatedAt"

	default:
		return nil, fmt.Errorf("Cannot marshal binary '%v'", s)
	}
	return []byte(data), nil
}

func (s *GatewaySessionSearchField) UnmarshalText(b []byte) error {
	str := strings.Trim(string(b), `"`)

	switch str {
	case "id":
		*s = GatewaySessionSearch_Id
	case "gatewayAccountId":
		*s = GatewaySessionSearch_GatewayAccountID
	case "expiration":
		*s = GatewaySessionSearch_Expiration
	case "createdAt":
		*s = GatewaySessionSearch_CreatedAt
	case "updatedAt":
		*s = GatewaySessionSearch_UpdatedAt

	default:
		return fmt.Errorf("Cannot unmarshal text '%s'", str)
	}
	return nil
}

func (s *GatewaySessionSearchField) UnmarshalBinary(b []byte) error {
	str := strings.Trim(string(b), `"`)

	switch str {
	case "id":
		*s = GatewaySessionSearch_Id
	case "gatewayAccountId":
		*s = GatewaySessionSearch_GatewayAccountID
	case "expiration":
		*s = GatewaySessionSearch_Expiration
	case "createdAt":
		*s = GatewaySessionSearch_CreatedAt
	case "updatedAt":
		*s = GatewaySessionSearch_UpdatedAt

	default:
		return fmt.Errorf("Cannot unmarshal binary '%s'", str)
	}
	return nil
}

func (s GatewaySessionSearchField) DbFieldName() string {
	switch s {
	case GatewaySessionSearch_Id:
		return "id"
	case GatewaySessionSearch_GatewayAccountID:
		return "gateway_account_id"
	case GatewaySessionSearch_Expiration:
		return "expiration"
	case GatewaySessionSearch_CreatedAt:
		return "created_at"
	case GatewaySessionSearch_UpdatedAt:
		return "updated_at"

	}
	return ""
}

type GatewaySessionSearchRequest struct {
	searcher.SearchRequestFields

	Filters     []GatewaySessionSearchFilter `json:"filters"`
	FilterGroup searcher.FilterGroup         `json:"filterGroup"`
	OrderBy     GatewaySessionOrderBy        `json:"orderBy"`
	OrderBys    []GatewaySessionOrderBy      `json:"orderBys"`
	Fields      []GatewaySessionSearchField  `json:"fields"`
	IsByID      bool                         `json:"isById"`
}

type GatewaySessionSearchFilter struct {
	Field     GatewaySessionSearchField `json:"field"`
	Value     interface{}               `json:"value"`
	Operator  searcher.FilterOperator   `json:"operator"`
	Condition searcher.FilterCondition  `json:"condition"`
}

type GatewaySessionOrderBy struct {
	Field      GatewaySessionSearchField `json:"field"`
	Descending bool                      `json:"desc"`
}

/*
GatewaySessionByID constructs a GatewaySessionSearchRequest to pull
a GatewaySession by it's ID.

You can add additional options using functions.

Handlers may choose to return (*GatewaySession, error) by checking the
IsSearchByID() function.
*/
func GatewaySessionByID(ID zeus.Uuid, options ...func(*GatewaySessionSearchRequest)) GatewaySessionSearchRequest {
	var searchRequest GatewaySessionSearchRequest

	searchRequest.AddFilter(
		GatewaySessionSearch_Id,
		ID,
		searcher.Operator_EQ,
		searcher.Condition_AND)

	searchRequest.Limit = 1
	searchRequest.IsByID = true

	for _, f := range options {
		f(&searchRequest)
	}

	return searchRequest
}

// implement searcher.SearchRequest interface
func (sr *GatewaySessionSearchRequest) GetTableName() string {
	return "shopify_gateway_sessions"
}

func (sr *GatewaySessionSearchRequest) GetFilters() []searcher.Filter {
	filters := []searcher.Filter{}

	for _, f := range sr.Filters {
		filter := searcher.Filter{
			Field:     f.Field,
			Value:     f.Value,
			Operator:  f.Operator,
			Condition: f.Condition,
		}
		filters = append(filters, filter)
	}

	return filters
}

func (sr *GatewaySessionSearchRequest) GetFilterGroup() searcher.FilterGroup {
	return sr.FilterGroup
}

func (sr *GatewaySessionSearchRequest) GetOrderBy() searcher.OrderBy {
	return searcher.OrderBy{
		Field:      sr.OrderBy.Field,
		Descending: sr.OrderBy.Descending,
	}
}

func (sr *GatewaySessionSearchRequest) GetOrderBys() []searcher.OrderBy {
	orderBys := make([]searcher.OrderBy, len(sr.OrderBys))
	for i, value := range sr.OrderBys {
		orderBys[i] = searcher.OrderBy{
			Field:      value.Field,
			Descending: value.Descending,
		}
	}
	return orderBys
}

func (sr *GatewaySessionSearchRequest) GetLimit() int {
	return sr.Limit
}

func (sr *GatewaySessionSearchRequest) GetOffset() int {
	return sr.Offset
}

func (sr *GatewaySessionSearchRequest) IsSearchByID() bool {
	return sr.IsByID
}

func (sr *GatewaySessionSearchRequest) AddFilter(field GatewaySessionSearchField, value interface{}, operator searcher.FilterOperator, condition searcher.FilterCondition) {
	if len(sr.FilterGroup.Filters) > 0 || len(sr.FilterGroup.FilterGroups) > 0 {
		panic("Filters cannot be used with FilterGroups")
	}
	f := GatewaySessionSearchFilter{
		Field:     field,
		Value:     value,
		Operator:  operator,
		Condition: condition,
	}
	sr.Filters = append(sr.Filters, f)
}

func (sr *GatewaySessionSearchRequest) SetFilterGroup(fg searcher.FilterGroup) {
	if len(sr.Filters) > 0 {
		panic("FilterGroups cannot be used with Filters")
	}
	sr.FilterGroup = fg
}

func (sr *GatewaySessionSearchRequest) SetOrderBy(field GatewaySessionSearchField, isDescending bool) {
	sr.OrderBy = GatewaySessionOrderBy{
		Field:      field,
		Descending: isDescending,
	}

	// Set this primary order by as the first in the slice
	sr.OrderBys = []GatewaySessionOrderBy{sr.OrderBy}
}

func (sr *GatewaySessionSearchRequest) SetOrderBys(orderBys ...GatewaySessionOrderBy) {
	sr.OrderBys = append([]GatewaySessionOrderBy{}, orderBys...)
}

func (sr *GatewaySessionSearchRequest) GetAllFields() []string {
	return []string{
		"id",
		"gateway_account_id",
		"expiration",
		"created_at",
		"updated_at",
	}
}

func (sr *GatewaySessionSearchRequest) GetFields() []string {
	fields := []string{}

	for _, f := range sr.Fields {
		fields = append(fields, f.DbFieldName())
	}

	return fields
}