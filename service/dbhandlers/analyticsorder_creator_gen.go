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
type AnalyticsOrderCreator interface {
	Create(c types.AnalyticsOrder) (zeus.Uuid, error)
	TxCreate(tx *sqlx.Tx, c types.AnalyticsOrder) (zeus.Uuid, error)
}

// implement SQL based creator
type sqlAnalyticsOrderCreator struct {
	db *sqlx.DB
}

func newSqlAnalyticsOrderCreator(db *sqlx.DB) AnalyticsOrderCreator {
	return &sqlAnalyticsOrderCreator{db: db}
}

func NewSqlAnalyticsOrderCreator(db *sqlx.DB) AnalyticsOrderCreator {
	return &sqlAnalyticsOrderCreator{db: db}
}

func (r *sqlAnalyticsOrderCreator) Create(c types.AnalyticsOrder) (zeus.Uuid, error) {
	columns := []string{
		"shop_name",
		"order_id",
		"customer_id",
		"customer_email",
		"total_price",
		"gateway",
		"financial_status",
		"fulfillment_status",
		"test",
		"checkout_id",
		"checkout_token",
	}

	return creator.Insert(r.db.DB, "shopify_analytics_orders", columns, c.ShopName, c.OrderID, c.CustomerID, c.CustomerEmail, c.TotalPrice, c.Gateway, c.FinancialStatus, c.FulfillmentStatus, c.Test, c.CheckoutID, c.CheckoutToken)
}

func (r *sqlAnalyticsOrderCreator) TxCreate(tx *sqlx.Tx, c types.AnalyticsOrder) (zeus.Uuid, error) {
	columns := []string{
		"shop_name",
		"order_id",
		"customer_id",
		"customer_email",
		"total_price",
		"gateway",
		"financial_status",
		"fulfillment_status",
		"test",
		"checkout_id",
		"checkout_token",
	}

	return creator.TxInsert(tx.Tx, "shopify_analytics_orders", columns, c.ShopName, c.OrderID, c.CustomerID, c.CustomerEmail, c.TotalPrice, c.Gateway, c.FinancialStatus, c.FulfillmentStatus, c.Test, c.CheckoutID, c.CheckoutToken)
}

// implement Fake creator for testing
type FakeAnalyticsOrderCreator struct {
	fakeResponse       zeus.Uuid
	fakeErr            error
	lastAnalyticsOrder types.AnalyticsOrder
	allAnalyticsOrder  []types.AnalyticsOrder
}

func NewFakeAnalyticsOrderCreatorWithError(fakeErr error) *FakeAnalyticsOrderCreator {
	return &FakeAnalyticsOrderCreator{fakeErr: fakeErr}
}

func NewFakeAnalyticsOrderCreator(fakeResponse zeus.Uuid) *FakeAnalyticsOrderCreator {
	return &FakeAnalyticsOrderCreator{fakeResponse: fakeResponse}
}

func (r *FakeAnalyticsOrderCreator) Create(c types.AnalyticsOrder) (zeus.Uuid, error) {
	logrus.WithField("createRequest", c).Info("returning AnalyticsOrder fake")

	if r.fakeErr != nil {
		return zeus.Uuid(""), r.fakeErr
	}
	r.lastAnalyticsOrder = c
	r.allAnalyticsOrder = append(r.allAnalyticsOrder, c)
	return r.fakeResponse, nil
}

func (r *FakeAnalyticsOrderCreator) TxCreate(tx *sqlx.Tx, c types.AnalyticsOrder) (zeus.Uuid, error) {
	return r.Create(c)
}

func (r *FakeAnalyticsOrderCreator) GetLastCreate() types.AnalyticsOrder {
	return r.lastAnalyticsOrder
}
func (r *FakeAnalyticsOrderCreator) GetAllCreated() []types.AnalyticsOrder {
	return r.allAnalyticsOrder
}

func (r *FakeAnalyticsOrderCreator) GetFakeResponse() zeus.Uuid {
	return r.fakeResponse
}
