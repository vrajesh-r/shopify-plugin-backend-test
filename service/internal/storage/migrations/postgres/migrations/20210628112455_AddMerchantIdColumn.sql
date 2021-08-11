
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE shopify_gateway_checkouts ADD COLUMN merchant_id varchar(64) DEFAULT '';

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE shopify_gateway_checkouts DROP COLUMN IF EXISTS merchant_id;
