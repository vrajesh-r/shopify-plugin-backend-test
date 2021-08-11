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
type DraftOrderCartCheckoutSearcher interface {
	Search(searchRequest search.DraftOrderCartCheckoutSearchRequest) ([]types.DraftOrderCartCheckout, error)
	Count(searchRequest search.DraftOrderCartCheckoutSearchRequest) (searcher.CountResult, error)
	ById(id zeus.Uuid) (types.DraftOrderCartCheckout, error)

	TxSearch(tx *sqlx.Tx, searchRequest search.DraftOrderCartCheckoutSearchRequest) ([]types.DraftOrderCartCheckout, error)
	TxCount(tx *sqlx.Tx, searchRequest search.DraftOrderCartCheckoutSearchRequest) (searcher.CountResult, error)
	TxById(tx *sqlx.Tx, id zeus.Uuid) (types.DraftOrderCartCheckout, error)
}

// implement SQL based searcher
type sqlDraftOrderCartCheckoutSearcher struct {
	db *sqlx.DB
}

func newSqlDraftOrderCartCheckoutSearcher(db *sqlx.DB) DraftOrderCartCheckoutSearcher {
	return &sqlDraftOrderCartCheckoutSearcher{db}
}

func NewSqlDraftOrderCartCheckoutSearcher(db *sqlx.DB) DraftOrderCartCheckoutSearcher {
	return &sqlDraftOrderCartCheckoutSearcher{db}
}

func (r *sqlDraftOrderCartCheckoutSearcher) Search(searchRequest search.DraftOrderCartCheckoutSearchRequest) ([]types.DraftOrderCartCheckout, error) {
	results := []types.DraftOrderCartCheckout{}

	sqlStr, values, err := searcher.GetSelectSql(&searchRequest)

	if err != nil {
		return nil, fmt.Errorf("Error generating search SQL for DraftOrderCartCheckoutSearcher : %s", err.Error())
	}

	err = r.db.Select(&results, sqlStr, values.([]interface{})...)

	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return results, nil
		}
	}

	return results, err
}

func (r *sqlDraftOrderCartCheckoutSearcher) TxSearch(tx *sqlx.Tx, searchRequest search.DraftOrderCartCheckoutSearchRequest) ([]types.DraftOrderCartCheckout, error) {
	results := []types.DraftOrderCartCheckout{}

	sqlStr, values, err := searcher.GetSelectSql(&searchRequest)

	if err != nil {
		return nil, fmt.Errorf("Error generating search SQL for DraftOrderCartCheckoutSearcher : %s", err.Error())
	}

	err = tx.Select(&results, sqlStr, values.([]interface{})...)

	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return results, nil
		}
	}

	return results, err
}

func (r *sqlDraftOrderCartCheckoutSearcher) Count(searchRequest search.DraftOrderCartCheckoutSearchRequest) (searcher.CountResult, error) {
	result := searcher.CountResult{}

	sqlStr, values, err := searcher.GetCountSql(&searchRequest)

	if err != nil {
		return result, fmt.Errorf("Error generating count SQL for DraftOrderCartCheckoutSearcher : %s", err.Error())
	}

	err = r.db.Get(&result, sqlStr, values.([]interface{})...)

	return result, err
}

func (r *sqlDraftOrderCartCheckoutSearcher) TxCount(tx *sqlx.Tx, searchRequest search.DraftOrderCartCheckoutSearchRequest) (searcher.CountResult, error) {
	result := searcher.CountResult{}

	sqlStr, values, err := searcher.GetCountSql(&searchRequest)

	if err != nil {
		return result, fmt.Errorf("Error generating count SQL for DraftOrderCartCheckoutSearcher : %s", err.Error())
	}

	err = tx.Get(&result, sqlStr, values.([]interface{})...)

	return result, err
}

func (r *sqlDraftOrderCartCheckoutSearcher) ById(id zeus.Uuid) (types.DraftOrderCartCheckout, error) {
	result := types.DraftOrderCartCheckout{}

	fieldsStr := searcher.GetSelectFieldsAvailable(&search.DraftOrderCartCheckoutSearchRequest{})

	err := r.db.Get(&result, fmt.Sprintf("SELECT %s FROM shopify_shops_draft_order_cart_checkouts WHERE id=$1", fieldsStr), string(id))

	return result, err
}

func (r *sqlDraftOrderCartCheckoutSearcher) TxById(tx *sqlx.Tx, id zeus.Uuid) (types.DraftOrderCartCheckout, error) {
	result := types.DraftOrderCartCheckout{}

	fieldsStr := searcher.GetSelectFieldsAvailable(&search.DraftOrderCartCheckoutSearchRequest{})

	err := tx.Get(&result, fmt.Sprintf("SELECT %s FROM shopify_shops_draft_order_cart_checkouts WHERE id=$1", fieldsStr), string(id))

	return result, err
}

// implement Fake searcher for testing
type FakeDraftOrderCartCheckoutSearcher struct {
	fakeResponse []types.DraftOrderCartCheckout
	requests     []search.DraftOrderCartCheckoutSearchRequest

	fakeIdMap map[zeus.Uuid]types.DraftOrderCartCheckout
	fakeError error

	// On the nth occurrence of a call to this searcher, return the slice, otherwise default to `fakeResponse`
	onOccurrenceReturns map[int][]types.DraftOrderCartCheckout

	onOccurrenceCountReturns map[int]int
}

func NewFakeDraftOrderCartCheckoutSearcher(fakeResponse []types.DraftOrderCartCheckout) DraftOrderCartCheckoutSearcher {
	return &FakeDraftOrderCartCheckoutSearcher{
		fakeResponse:             fakeResponse,
		requests:                 []search.DraftOrderCartCheckoutSearchRequest{},
		fakeIdMap:                map[zeus.Uuid]types.DraftOrderCartCheckout{},
		fakeError:                nil,
		onOccurrenceReturns:      map[int][]types.DraftOrderCartCheckout{},
		onOccurrenceCountReturns: map[int]int{},
	}
}

func (r *FakeDraftOrderCartCheckoutSearcher) Search(searchRequest search.DraftOrderCartCheckoutSearchRequest) ([]types.DraftOrderCartCheckout, error) {
	r.requests = append(r.requests, searchRequest)

	if len(r.onOccurrenceReturns) > 0 {
		if occ, ok := r.onOccurrenceReturns[len(r.requests)]; ok {
			return occ, nil
		}
	}

	return r.fakeResponse, r.fakeError
}

func (r *FakeDraftOrderCartCheckoutSearcher) GetRequests() []search.DraftOrderCartCheckoutSearchRequest {
	return r.requests
}

func (r *FakeDraftOrderCartCheckoutSearcher) TxSearch(tx *sqlx.Tx, searchRequest search.DraftOrderCartCheckoutSearchRequest) ([]types.DraftOrderCartCheckout, error) {
	return r.Search(searchRequest)
}

func (r *FakeDraftOrderCartCheckoutSearcher) Count(searchRequest search.DraftOrderCartCheckoutSearchRequest) (searcher.CountResult, error) {
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

func (r *FakeDraftOrderCartCheckoutSearcher) TxCount(tx *sqlx.Tx, searchRequest search.DraftOrderCartCheckoutSearchRequest) (searcher.CountResult, error) {
	return r.Count(searchRequest)
}

func (r *FakeDraftOrderCartCheckoutSearcher) ById(id zeus.Uuid) (types.DraftOrderCartCheckout, error) {
	if len(r.fakeIdMap) == 0 {
		return r.fakeResponse[0], nil
	}

	entity, ok := r.fakeIdMap[id]
	if !ok {
		return types.DraftOrderCartCheckout{}, errors.New("Not able to find DraftOrderCartCheckout by id: " + string(id))
	}

	return entity, r.fakeError
}

func (r *FakeDraftOrderCartCheckoutSearcher) TxById(tx *sqlx.Tx, id zeus.Uuid) (types.DraftOrderCartCheckout, error) {
	return r.ById(id)
}

func (r *FakeDraftOrderCartCheckoutSearcher) SetByIdResponses(resp map[zeus.Uuid]types.DraftOrderCartCheckout) {
	r.fakeIdMap = resp
}

func (r *FakeDraftOrderCartCheckoutSearcher) SetError(err error) {
	r.fakeError = err
}

func (r *FakeDraftOrderCartCheckoutSearcher) GetFakeError() error {
	return r.fakeError
}

func (r *FakeDraftOrderCartCheckoutSearcher) GetOnOccurrenceReturns() map[int][]types.DraftOrderCartCheckout {
	return r.onOccurrenceReturns
}

func (r *FakeDraftOrderCartCheckoutSearcher) SetOnOccurrenceReturns(val map[int][]types.DraftOrderCartCheckout) {
	r.onOccurrenceReturns = val
}

func (r *FakeDraftOrderCartCheckoutSearcher) GetOnOccurrenceCountReturns() map[int]int {
	return r.onOccurrenceCountReturns
}

func (r *FakeDraftOrderCartCheckoutSearcher) SetOnOccurrenceCountReturns(val map[int]int) {
	r.onOccurrenceCountReturns = val
}
