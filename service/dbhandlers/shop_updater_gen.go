// Code generated by https://github.com/getbread/breadkit/zeus/tree/master/generators/updater. DO NOT EDIT.

package dbhandlers

import (
	"fmt"

	zeus "github.com/getbread/breadkit/zeus/types"
	"github.com/getbread/breadkit/zeus/updater"
	update "github.com/getbread/shopify_plugin_backend/service/update"
	"github.com/jmoiron/sqlx"
)

// interface for this updater
type ShopUpdater interface {
	Update(updateRequest update.ShopUpdateRequest) error
	TxUpdate(tx *sqlx.Tx, updateRequest update.ShopUpdateRequest) error
	DeleteById(id zeus.Uuid) error
	TxDeleteById(tx *sqlx.Tx, id zeus.Uuid) error
}

// implement SQL based updater
type sqlShopUpdater struct {
	db *sqlx.DB
}

func newSqlShopUpdater(db *sqlx.DB) ShopUpdater {
	return &sqlShopUpdater{db: db}
}

func NewSqlShopUpdater(db *sqlx.DB) ShopUpdater {
	return &sqlShopUpdater{db: db}
}

func (r *sqlShopUpdater) Update(updateRequest update.ShopUpdateRequest) error {

	sqlStr, values, err := updater.GetUpdateSql(&updateRequest)

	if err != nil {
		return fmt.Errorf("Error generating update SQL for ShopUpdater : %s", err.Error())
	}

	_, err = r.db.Exec(sqlStr, values.([]interface{})...)

	return err
}

func (r *sqlShopUpdater) TxUpdate(tx *sqlx.Tx, updateRequest update.ShopUpdateRequest) error {

	sqlStr, values, err := updater.GetUpdateSql(&updateRequest)

	if err != nil {
		return fmt.Errorf("Error generating update SQL for ShopUpdater : %s", err.Error())
	}

	_, err = tx.Exec(sqlStr, values.([]interface{})...)

	return err
}

func (r *sqlShopUpdater) DeleteById(id zeus.Uuid) error {
	ur := update.ShopUpdateRequest{
		Id: id,
	}

	sqlStr, values, err := updater.GetDeleteSql(&ur)

	if err != nil {
		return fmt.Errorf("Error generating delete SQL for ShopUpdater : %s", err.Error())
	}

	_, err = r.db.Exec(sqlStr, values.([]interface{})...)

	return err
}

func (r *sqlShopUpdater) TxDeleteById(tx *sqlx.Tx, id zeus.Uuid) error {
	ur := update.ShopUpdateRequest{
		Id: id,
	}

	sqlStr, values, err := updater.GetDeleteSql(&ur)

	if err != nil {
		return fmt.Errorf("Error generating delete SQL for ShopUpdater : %s", err.Error())
	}

	_, err = tx.Exec(sqlStr, values.([]interface{})...)

	return err
}

// implement Fake updater for testing
type FakeShopUpdater struct {
	fakeResponse     error
	collectedUpdates []update.ShopUpdateRequest
	collectedDeletes []zeus.Uuid
}

func NewFakeShopUpdater(fakeResponse error) ShopUpdater {
	return &FakeShopUpdater{fakeResponse: fakeResponse}
}

func (r *FakeShopUpdater) Update(updateRequest update.ShopUpdateRequest) error {
	r.collectedUpdates = append(r.collectedUpdates, updateRequest)
	return r.fakeResponse
}

func (r *FakeShopUpdater) TxUpdate(tx *sqlx.Tx, updateRequest update.ShopUpdateRequest) error {
	return r.Update(updateRequest)
}

func (r *FakeShopUpdater) DeleteById(id zeus.Uuid) error {
	r.collectedDeletes = append(r.collectedDeletes, id)
	return nil
}

func (r *FakeShopUpdater) TxDeleteById(tx *sqlx.Tx, id zeus.Uuid) error {
	return r.DeleteById(id)
}

func (r *FakeShopUpdater) GetCollectedUpdates() []update.ShopUpdateRequest {
	return r.collectedUpdates
}

func (r *FakeShopUpdater) GetCollectedDeletes() []zeus.Uuid {
	return r.collectedDeletes
}
