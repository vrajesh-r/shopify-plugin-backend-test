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
type DraftOrderCartSearcher interface {
	Search(searchRequest search.DraftOrderCartSearchRequest) ([]types.DraftOrderCart, error)
	Count(searchRequest search.DraftOrderCartSearchRequest) (searcher.CountResult, error)
	ById(id zeus.Uuid) (types.DraftOrderCart, error)

	TxSearch(tx *sqlx.Tx, searchRequest search.DraftOrderCartSearchRequest) ([]types.DraftOrderCart, error)
	TxCount(tx *sqlx.Tx, searchRequest search.DraftOrderCartSearchRequest) (searcher.CountResult, error)
	TxById(tx *sqlx.Tx, id zeus.Uuid) (types.DraftOrderCart, error)
}

// implement SQL based searcher
type sqlDraftOrderCartSearcher struct {
	db *sqlx.DB
}

func newSqlDraftOrderCartSearcher(db *sqlx.DB) DraftOrderCartSearcher {
	return &sqlDraftOrderCartSearcher{db}
}

func NewSqlDraftOrderCartSearcher(db *sqlx.DB) DraftOrderCartSearcher {
	return &sqlDraftOrderCartSearcher{db}
}

func (r *sqlDraftOrderCartSearcher) Search(searchRequest search.DraftOrderCartSearchRequest) ([]types.DraftOrderCart, error) {
	results := []types.DraftOrderCart{}

	sqlStr, values, err := searcher.GetSelectSql(&searchRequest)

	if err != nil {
		return nil, fmt.Errorf("Error generating search SQL for DraftOrderCartSearcher : %s", err.Error())
	}

	err = r.db.Select(&results, sqlStr, values.([]interface{})...)

	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return results, nil
		}
	}

	return results, err
}

func (r *sqlDraftOrderCartSearcher) TxSearch(tx *sqlx.Tx, searchRequest search.DraftOrderCartSearchRequest) ([]types.DraftOrderCart, error) {
	results := []types.DraftOrderCart{}

	sqlStr, values, err := searcher.GetSelectSql(&searchRequest)

	if err != nil {
		return nil, fmt.Errorf("Error generating search SQL for DraftOrderCartSearcher : %s", err.Error())
	}

	err = tx.Select(&results, sqlStr, values.([]interface{})...)

	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return results, nil
		}
	}

	return results, err
}

func (r *sqlDraftOrderCartSearcher) Count(searchRequest search.DraftOrderCartSearchRequest) (searcher.CountResult, error) {
	result := searcher.CountResult{}

	sqlStr, values, err := searcher.GetCountSql(&searchRequest)

	if err != nil {
		return result, fmt.Errorf("Error generating count SQL for DraftOrderCartSearcher : %s", err.Error())
	}

	err = r.db.Get(&result, sqlStr, values.([]interface{})...)

	return result, err
}

func (r *sqlDraftOrderCartSearcher) TxCount(tx *sqlx.Tx, searchRequest search.DraftOrderCartSearchRequest) (searcher.CountResult, error) {
	result := searcher.CountResult{}

	sqlStr, values, err := searcher.GetCountSql(&searchRequest)

	if err != nil {
		return result, fmt.Errorf("Error generating count SQL for DraftOrderCartSearcher : %s", err.Error())
	}

	err = tx.Get(&result, sqlStr, values.([]interface{})...)

	return result, err
}

func (r *sqlDraftOrderCartSearcher) ById(id zeus.Uuid) (types.DraftOrderCart, error) {
	result := types.DraftOrderCart{}

	fieldsStr := searcher.GetSelectFieldsAvailable(&search.DraftOrderCartSearchRequest{})

	err := r.db.Get(&result, fmt.Sprintf("SELECT %s FROM shopify_shops_draft_order_carts WHERE id=$1", fieldsStr), string(id))

	return result, err
}

func (r *sqlDraftOrderCartSearcher) TxById(tx *sqlx.Tx, id zeus.Uuid) (types.DraftOrderCart, error) {
	result := types.DraftOrderCart{}

	fieldsStr := searcher.GetSelectFieldsAvailable(&search.DraftOrderCartSearchRequest{})

	err := tx.Get(&result, fmt.Sprintf("SELECT %s FROM shopify_shops_draft_order_carts WHERE id=$1", fieldsStr), string(id))

	return result, err
}

// implement Fake searcher for testing
type FakeDraftOrderCartSearcher struct {
	fakeResponse []types.DraftOrderCart
	requests     []search.DraftOrderCartSearchRequest

	fakeIdMap map[zeus.Uuid]types.DraftOrderCart
	fakeError error

	// On the nth occurrence of a call to this searcher, return the slice, otherwise default to `fakeResponse`
	onOccurrenceReturns map[int][]types.DraftOrderCart

	onOccurrenceCountReturns map[int]int
}

func NewFakeDraftOrderCartSearcher(fakeResponse []types.DraftOrderCart) DraftOrderCartSearcher {
	return &FakeDraftOrderCartSearcher{
		fakeResponse:             fakeResponse,
		requests:                 []search.DraftOrderCartSearchRequest{},
		fakeIdMap:                map[zeus.Uuid]types.DraftOrderCart{},
		fakeError:                nil,
		onOccurrenceReturns:      map[int][]types.DraftOrderCart{},
		onOccurrenceCountReturns: map[int]int{},
	}
}

func (r *FakeDraftOrderCartSearcher) Search(searchRequest search.DraftOrderCartSearchRequest) ([]types.DraftOrderCart, error) {
	r.requests = append(r.requests, searchRequest)

	if len(r.onOccurrenceReturns) > 0 {
		if occ, ok := r.onOccurrenceReturns[len(r.requests)]; ok {
			return occ, nil
		}
	}

	return r.fakeResponse, r.fakeError
}

func (r *FakeDraftOrderCartSearcher) GetRequests() []search.DraftOrderCartSearchRequest {
	return r.requests
}

func (r *FakeDraftOrderCartSearcher) TxSearch(tx *sqlx.Tx, searchRequest search.DraftOrderCartSearchRequest) ([]types.DraftOrderCart, error) {
	return r.Search(searchRequest)
}

func (r *FakeDraftOrderCartSearcher) Count(searchRequest search.DraftOrderCartSearchRequest) (searcher.CountResult, error) {
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

func (r *FakeDraftOrderCartSearcher) TxCount(tx *sqlx.Tx, searchRequest search.DraftOrderCartSearchRequest) (searcher.CountResult, error) {
	return r.Count(searchRequest)
}

func (r *FakeDraftOrderCartSearcher) ById(id zeus.Uuid) (types.DraftOrderCart, error) {
	if len(r.fakeIdMap) == 0 {
		return r.fakeResponse[0], nil
	}

	entity, ok := r.fakeIdMap[id]
	if !ok {
		return types.DraftOrderCart{}, errors.New("Not able to find DraftOrderCart by id: " + string(id))
	}

	return entity, r.fakeError
}

func (r *FakeDraftOrderCartSearcher) TxById(tx *sqlx.Tx, id zeus.Uuid) (types.DraftOrderCart, error) {
	return r.ById(id)
}

func (r *FakeDraftOrderCartSearcher) SetByIdResponses(resp map[zeus.Uuid]types.DraftOrderCart) {
	r.fakeIdMap = resp
}

func (r *FakeDraftOrderCartSearcher) SetError(err error) {
	r.fakeError = err
}

func (r *FakeDraftOrderCartSearcher) GetFakeError() error {
	return r.fakeError
}

func (r *FakeDraftOrderCartSearcher) GetOnOccurrenceReturns() map[int][]types.DraftOrderCart {
	return r.onOccurrenceReturns
}

func (r *FakeDraftOrderCartSearcher) SetOnOccurrenceReturns(val map[int][]types.DraftOrderCart) {
	r.onOccurrenceReturns = val
}

func (r *FakeDraftOrderCartSearcher) GetOnOccurrenceCountReturns() map[int]int {
	return r.onOccurrenceCountReturns
}

func (r *FakeDraftOrderCartSearcher) SetOnOccurrenceCountReturns(val map[int]int) {
	r.onOccurrenceCountReturns = val
}
