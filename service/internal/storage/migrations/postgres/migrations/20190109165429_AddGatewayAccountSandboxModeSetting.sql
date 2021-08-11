
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

ALTER TABLE shopify_gateway_accounts ADD COLUMN production boolean NOT NULL DEFAULT true;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

ALTER TABLE shopify_gateway_accounts DROP COLUMN IF EXISTS production;
