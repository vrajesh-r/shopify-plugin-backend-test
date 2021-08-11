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
type PlusGatewayCheckoutCreator interface {
	Create(c types.PlusGatewayCheckout) (zeus.Uuid, error)
	TxCreate(tx *sqlx.Tx, c types.PlusGatewayCheckout) (zeus.Uuid, error)
}

// implement SQL based creator
type sqlPlusGatewayCheckoutCreator struct {
	db *sqlx.DB
}

func newSqlPlusGatewayCheckoutCreator(db *sqlx.DB) PlusGatewayCheckoutCreator {
	return &sqlPlusGatewayCheckoutCreator{db: db}
}

func NewSqlPlusGatewayCheckoutCreator(db *sqlx.DB) PlusGatewayCheckoutCreator {
	return &sqlPlusGatewayCheckoutCreator{db: db}
}

func (r *sqlPlusGatewayCheckoutCreator) Create(c types.PlusGatewayCheckout) (zeus.Uuid, error) {
	columns := []string{
		"checkout_id",
		"transaction_id",
	}

	return creator.Insert(r.db.DB, "shopify_plus_gateway_checkouts", columns, c.CheckoutID, c.TransactionID)
}

func (r *sqlPlusGatewayCheckoutCreator) TxCreate(tx *sqlx.Tx, c types.PlusGatewayCheckout) (zeus.Uuid, error) {
	columns := []string{
		"checkout_id",
		"transaction_id",
	}

	return creator.TxInsert(tx.Tx, "shopify_plus_gateway_checkouts", columns, c.CheckoutID, c.TransactionID)
}

// implement Fake creator for testing
type FakePlusGatewayCheckoutCreator struct {
	fakeResponse            zeus.Uuid
	fakeErr                 error
	lastPlusGatewayCheckout types.PlusGatewayCheckout
	allPlusGatewayCheckout  []types.PlusGatewayCheckout
}

func NewFakePlusGatewayCheckoutCreatorWithError(fakeErr error) *FakePlusGatewayCheckoutCreator {
	return &FakePlusGatewayCheckoutCreator{fakeErr: fakeErr}
}

func NewFakePlusGatewayCheckoutCreator(fakeResponse zeus.Uuid) *FakePlusGatewayCheckoutCreator {
	return &FakePlusGatewayCheckoutCreator{fakeResponse: fakeResponse}
}

func (r *FakePlusGatewayCheckoutCreator) Create(c types.PlusGatewayCheckout) (zeus.Uuid, error) {
	logrus.WithField("createRequest", c).Info("returning PlusGatewayCheckout fake")

	if r.fakeErr != nil {
		return zeus.Uuid(""), r.fakeErr
	}
	r.lastPlusGatewayCheckout = c
	r.allPlusGatewayCheckout = append(r.allPlusGatewayCheckout, c)
	return r.fakeResponse, nil
}

func (r *FakePlusGatewayCheckoutCreator) TxCreate(tx *sqlx.Tx, c types.PlusGatewayCheckout) (zeus.Uuid, error) {
	return r.Create(c)
}

func (r *FakePlusGatewayCheckoutCreator) GetLastCreate() types.PlusGatewayCheckout {
	return r.lastPlusGatewayCheckout
}
func (r *FakePlusGatewayCheckoutCreator) GetAllCreated() []types.PlusGatewayCheckout {
	return r.allPlusGatewayCheckout
}

func (r *FakePlusGatewayCheckoutCreator) GetFakeResponse() zeus.Uuid {
	return r.fakeResponse
}