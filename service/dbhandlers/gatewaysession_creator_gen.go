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
type GatewaySessionCreator interface {
	Create(c types.GatewaySession) (zeus.Uuid, error)
	TxCreate(tx *sqlx.Tx, c types.GatewaySession) (zeus.Uuid, error)
}

// implement SQL based creator
type sqlGatewaySessionCreator struct {
	db *sqlx.DB
}

func newSqlGatewaySessionCreator(db *sqlx.DB) GatewaySessionCreator {
	return &sqlGatewaySessionCreator{db: db}
}

func NewSqlGatewaySessionCreator(db *sqlx.DB) GatewaySessionCreator {
	return &sqlGatewaySessionCreator{db: db}
}

func (r *sqlGatewaySessionCreator) Create(c types.GatewaySession) (zeus.Uuid, error) {
	columns := []string{
		"gateway_account_id",
		"expiration",
	}

	return creator.Insert(r.db.DB, "shopify_gateway_sessions", columns, c.GatewayAccountID, c.Expiration)
}

func (r *sqlGatewaySessionCreator) TxCreate(tx *sqlx.Tx, c types.GatewaySession) (zeus.Uuid, error) {
	columns := []string{
		"gateway_account_id",
		"expiration",
	}

	return creator.TxInsert(tx.Tx, "shopify_gateway_sessions", columns, c.GatewayAccountID, c.Expiration)
}

// implement Fake creator for testing
type FakeGatewaySessionCreator struct {
	fakeResponse       zeus.Uuid
	fakeErr            error
	lastGatewaySession types.GatewaySession
	allGatewaySession  []types.GatewaySession
}

func NewFakeGatewaySessionCreatorWithError(fakeErr error) *FakeGatewaySessionCreator {
	return &FakeGatewaySessionCreator{fakeErr: fakeErr}
}

func NewFakeGatewaySessionCreator(fakeResponse zeus.Uuid) *FakeGatewaySessionCreator {
	return &FakeGatewaySessionCreator{fakeResponse: fakeResponse}
}

func (r *FakeGatewaySessionCreator) Create(c types.GatewaySession) (zeus.Uuid, error) {
	logrus.WithField("createRequest", c).Info("returning GatewaySession fake")

	if r.fakeErr != nil {
		return zeus.Uuid(""), r.fakeErr
	}
	r.lastGatewaySession = c
	r.allGatewaySession = append(r.allGatewaySession, c)
	return r.fakeResponse, nil
}

func (r *FakeGatewaySessionCreator) TxCreate(tx *sqlx.Tx, c types.GatewaySession) (zeus.Uuid, error) {
	return r.Create(c)
}

func (r *FakeGatewaySessionCreator) GetLastCreate() types.GatewaySession {
	return r.lastGatewaySession
}
func (r *FakeGatewaySessionCreator) GetAllCreated() []types.GatewaySession {
	return r.allGatewaySession
}

func (r *FakeGatewaySessionCreator) GetFakeResponse() zeus.Uuid {
	return r.fakeResponse
}