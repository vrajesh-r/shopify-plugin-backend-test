
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

ALTER TABLE shopify_gateway_checkouts ADD COLUMN amount_str varchar(32) DEFAULT '';
ALTER TABLE shopify_gateway_checkouts ADD COLUMN bread_version varchar(32) DEFAULT '';

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

ALTER TABLE shopify_gateway_checkouts DROP COLUMN IF EXISTS amount_str;
ALTER TABLE shopify_gateway_checkouts DROP COLUMN IF EXISTS bread_version;