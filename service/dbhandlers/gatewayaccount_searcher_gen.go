// Code generated by https://github.com/getbread/breadkit/zeus/tree/master/generators/searcher. DO NOT EDIT.

package dbhandlers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/getbread/breadkit/zeus/searcher"
	zeus "github.com/getbread/breadkit/zeus/types"
	"github.com/jmoiron/sqlx"

	search "github.com/getbread/shopify_plugin_backend/service/search"
	types "github.com/getbread/shopify_plugin_backend/service/types"
)

// interface for this searcher
type GatewayAccountSearcher interface {
	Search(searchRequest search.GatewayAccountSearchRequest) ([]types.GatewayAccount, error)
	Count(searchRequest search.GatewayAccountSearchRequest) (searcher.CountResult, error)
	ById(id zeus.Uuid) (types.GatewayAccount, error)

	TxSearch(tx *sqlx.Tx, searchRequest search.GatewayAccountSearchRequest) ([]types.GatewayAccount, error)
	TxCount(tx *sqlx.Tx, searchRequest search.GatewayAccountSearchRequest) (searcher.CountResult, error)
	TxById(tx *sqlx.Tx, id zeus.Uuid) (types.GatewayAccount, error)
}

// implement SQL based searcher
type sqlGatewayAccountSearcher struct {
	db *sqlx.DB
}

func newSqlGatewayAccountSearcher(db *sqlx.DB) GatewayAccountSearcher {
	return &sqlGatewayAccountSearcher{db}
}

func NewSqlGatewayAccountSearcher(db *sqlx.DB) GatewayAccountSearcher {
	return &sqlGatewayAccountSearcher{db}
}

func (r *sqlGatewayAccountSearcher) Search(searchRequest search.GatewayAccountSearchRequest) ([]types.GatewayAccount, error) {
	results := []types.GatewayAccount{}

	sqlStr, values, err := searcher.GetSelectSql(&searchRequest)

	if err != nil {
		return nil, fmt.Errorf("Error generating search SQL for GatewayAccountSearcher : %s", err.Error())
	}

	err = r.db.Select(&results, sqlStr, values.([]interface{})...)

	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return results, nil
		}
	}

	return results, err
}

func (r *sqlGatewayAccountSearcher) TxSearch(tx *sqlx.Tx, searchRequest search.GatewayAccountSearchRequest) ([]types.GatewayAccount, error) {
	results := []types.GatewayAccount{}

	sqlStr, values, err := searcher.GetSelectSql(&searchRequest)

	if err != nil {
		return nil, fmt.Errorf("Error generating search SQL for GatewayAccountSearcher : %s", err.Error())
	}

	err = tx.Select(&results, sqlStr, values.([]interface{})...)

	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return results, nil
		}
	}

	return results, err
}

func (r *sqlGatewayAccountSearcher) Count(searchRequest search.GatewayAccountSearchRequest) (searcher.CountResult, error) {
	result := searcher.CountResult{}

	sqlStr, values, err := searcher.GetCountSql(&searchRequest)

	if err != nil {
		return result, fmt.Errorf("Error generating count SQL for GatewayAccountSearcher : %s", err.Error())
	}

	err = r.db.Get(&result, sqlStr, values.([]interface{})...)

	return result, err
}

func (r *sqlGatewayAccountSearcher) TxCount(tx *sqlx.Tx, searchRequest search.GatewayAccountSearchRequest) (searcher.CountResult, error) {
	result := searcher.CountResult{}

	sqlStr, values, err := searcher.GetCountSql(&searchRequest)

	if err != nil {
		return result, fmt.Errorf("Error generating count SQL for GatewayAccountSearcher : %s", err.Error())
	}

	err = tx.Get(&result, sqlStr, values.([]interface{})...)

	return result, err
}

func (r *sqlGatewayAccountSearcher) ById(id zeus.Uuid) (types.GatewayAccount, error) {
	result := types.GatewayAccount{}

	fieldsStr := searcher.GetSelectFieldsAvailable(&search.GatewayAccountSearchRequest{})

	err := r.db.Get(&result, fmt.Sprintf("SELECT %s FROM shopify_gateway_accounts WHERE id=$1", fieldsStr), string(id))

	return result, err
}

func (r *sqlGatewayAccountSearcher) TxById(tx *sqlx.Tx, id zeus.Uuid) (types.GatewayAccount, error) {
	result := types.GatewayAccount{}

	fieldsStr := searcher.GetSelectFieldsAvailable(&search.GatewayAccountSearchRequest{})

	err := tx.Get(&result, fmt.Sprintf("SELECT %s FROM shopify_gateway_accounts WHERE id=$1", fieldsStr), string(id))

	return result, err
}

// implement Fake searcher for testing
type FakeGatewayAccountSearcher struct {
	fakeResponse []types.GatewayAccount
	requests     []search.GatewayAccountSearchRequest

	fakeIdMap map[zeus.Uuid]types.GatewayAccount
	fakeError error

	// On the nth occurrence of a call to this searcher, return the slice, otherwise default to `fakeResponse`
	onOccurrenceReturns map[int][]types.GatewayAccount

	onOccurrenceCountReturns map[int]int
}

func NewFakeGatewayAccountSearcher(fakeResponse []types.GatewayAccount) GatewayAccountSearcher {
	return &FakeGatewayAccountSearcher{
		fakeResponse:             fakeResponse,
		requests:                 []search.GatewayAccountSearchRequest{},
		fakeIdMap:                map[zeus.Uuid]types.GatewayAccount{},
		fakeError:                nil,
		onOccurrenceReturns:      map[int][]types.GatewayAccount{},
		onOccurrenceCountReturns: map[int]int{},
	}
}

func (r *FakeGatewayAccountSearcher) Search(searchRequest search.GatewayAccountSearchRequest) ([]types.GatewayAccount, error) {
	r.requests = append(r.requests, searchRequest)

	if len(r.onOccurrenceReturns) > 0 {
		if occ, ok := r.onOccurrenceReturns[len(r.requests)]; ok {
			return occ, nil
		}
	}

	return r.fakeResponse, r.fakeError
}

func (r *FakeGatewayAccountSearcher) GetRequests() []search.GatewayAccountSearchRequest {
	return r.requests
}

func (r *FakeGatewayAccountSearcher) TxSearch(tx *sqlx.Tx, searchRequest search.GatewayAccountSearchRequest) ([]types.GatewayAccount, error) {
	return r.Search(searchRequest)
}

func (r *FakeGatewayAccountSearcher) Count(searchRequest search.GatewayAccountSearchRequest) (searcher.CountResult, error) {
	count := searcher.CountResult{
		Count: len(r.fakeResponse),
	}

	r.requests = append(r.requests, searchRequest)

	if len(r.onOccurrenceCountReturns) > 0 {
		if occ, ok := r.onOccurrenceCountReturns[len(r.requests)]; ok {
			count.Count = occ
			return count, nil
		}
	}

	return count, r.fakeError
}

func (r *FakeGatewayAccountSearcher) TxCount(tx *sqlx.Tx, searchRequest search.GatewayAccountSearchRequest) (searcher.CountResult, error) {
	return r.Count(searchRequest)
}

func (r *FakeGatewayAccountSearcher) ById(id zeus.Uuid) (types.GatewayAccount, error) {
	if len(r.fakeIdMap) == 0 {
		return r.fakeResponse[0], nil
	}

	entity, ok := r.fakeIdMap[id]
	if !ok {
		return types.GatewayAccount{}, errors.New("Not able to find GatewayAccount by id: " + string(id))
	}

	return entity, r.fakeError
}

func (r *FakeGatewayAccountSearcher) TxById(tx *sqlx.Tx, id zeus.Uuid) (types.GatewayAccount, error) {
	return r.ById(id)
}

func (r *FakeGatewayAccountSearcher) SetByIdResponses(resp map[zeus.Uuid]types.GatewayAccount) {
	r.fakeIdMap = resp
}

func (r *FakeGatewayAccountSearcher) SetError(err error) {
	r.fakeError = err
}

func (r *FakeGatewayAccountSearcher) GetFakeError() error {
	return r.fakeError
}

func (r *FakeGatewayAccountSearcher) GetOnOccurrenceReturns() map[int][]types.GatewayAccount {
	return r.onOccurrenceReturns
}

func (r *FakeGatewayAccountSearcher) SetOnOccurrenceReturns(val map[int][]types.GatewayAccount) {
	r.onOccurrenceReturns = val
}

func (r *FakeGatewayAccountSearcher) GetOnOccurrenceCountReturns() map[int]int {
	return r.onOccurrenceCountReturns
}

func (r *FakeGatewayAccountSearcher) SetOnOccurrenceCountReturns(val map[int]int) {
	r.onOccurrenceCountReturns = val
}