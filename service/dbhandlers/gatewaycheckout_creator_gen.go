// Code generated by https://github.com/getbread/breadkit/zeus/tree/master/generators/creator. DO NOT EDIT.

package dbhandlers

import (
	"github.com/getbread/breadkit/zeus/creator"
	zeus "github.com/getbread/breadkit/zeus/types"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"

	types "github.com/getbread/shopify_plugin_backend/service/types"
)

// interface for this creator
type GatewayCheckoutCreator interface {
	Create(c types.GatewayCheckout) (zeus.Uuid, error)
	TxCreate(tx *sqlx.Tx, c types.GatewayCheckout) (zeus.Uuid, error)
}

// implement SQL based creator
type sqlGatewayCheckoutCreator struct {
	db *sqlx.DB
}

func newSqlGatewayCheckoutCreator(db *sqlx.DB) GatewayCheckoutCreator {
	return &sqlGatewayCheckoutCreator{db: db}
}

func NewSqlGatewayCheckoutCreator(db *sqlx.DB) GatewayCheckoutCreator {
	return &sqlGatewayCheckoutCreator{db: db}
}

func (r *sqlGatewayCheckoutCreator) Create(c types.GatewayCheckout) (zeus.Uuid, error) {
	columns := []string{
		"account_id",
		"transaction_id",
		"test",
		"reference",
		"currency",
		"amount",
		"callback_url",
		"complete_url",
		"cancel_url",
		"completed",
		"errored",
		"amount_str",
		"bread_version",
	}

	return creator.Insert(r.db.DB, "shopify_gateway_checkouts", columns, c.AccountID, c.TransactionID, c.Test, c.Reference, c.Currency, c.Amount, c.CallbackUrl, c.CompleteUrl, c.CancelUrl, c.Completed, c.Errored, c.AmountStr, c.BreadVersion)
}

func (r *sqlGatewayCheckoutCreator) TxCreate(tx *sqlx.Tx, c types.GatewayCheckout) (zeus.Uuid, error) {
	columns := []string{
		"account_id",
		"transaction_id",
		"test",
		"reference",
		"currency",
		"amount",
		"callback_url",
		"complete_url",
		"cancel_url",
		"completed",
		"errored",
		"amount_str",
		"bread_version",
	}

	return creator.TxInsert(tx.Tx, "shopify_gateway_checkouts", columns, c.AccountID, c.TransactionID, c.Test, c.Reference, c.Currency, c.Amount, c.CallbackUrl, c.CompleteUrl, c.CancelUrl, c.Completed, c.Errored, c.AmountStr, c.BreadVersion)
}

// implement Fake creator for testing
type FakeGatewayCheckoutCreator struct {
	fakeResponse        zeus.Uuid
	fakeErr             error
	lastGatewayCheckout types.GatewayCheckout
	allGatewayCheckout  []types.GatewayCheckout
}

func NewFakeGatewayCheckoutCreatorWithError(fakeErr error) *FakeGatewayCheckoutCreator {
	return &FakeGatewayCheckoutCreator{fakeErr: fakeErr}
}

func NewFakeGatewayCheckoutCreator(fakeResponse zeus.Uuid) *FakeGatewayCheckoutCreator {
	return &FakeGatewayCheckoutCreator{fakeResponse: fakeResponse}
}

func (r *FakeGatewayCheckoutCreator) Create(c types.GatewayCheckout) (zeus.Uuid, error) {
	logrus.WithField("createRequest", c).Info("returning GatewayCheckout fake")

	if r.fakeErr != nil {
		return zeus.Uuid(""), r.fakeErr
	}
	r.lastGatewayCheckout = c
	r.allGatewayCheckout = append(r.allGatewayCheckout, c)
	return r.fakeResponse, nil
}

func (r *FakeGatewayCheckoutCreator) TxCreate(tx *sqlx.Tx, c types.GatewayCheckout) (zeus.Uuid, error) {
	return r.Create(c)
}

func (r *FakeGatewayCheckoutCreator) GetLastCreate() types.GatewayCheckout {
	return r.lastGatewayCheckout
}
func (r *FakeGatewayCheckoutCreator) GetAllCreated() []types.GatewayCheckout {
	return r.allGatewayCheckout
}

func (r *FakeGatewayCheckoutCreator) GetFakeResponse() zeus.Uuid {
	return r.fakeResponse
}
