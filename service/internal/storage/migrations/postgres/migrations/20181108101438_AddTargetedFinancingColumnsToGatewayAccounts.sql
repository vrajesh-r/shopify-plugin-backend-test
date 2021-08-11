
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

ALTER TABLE shopify_gateway_accounts ADD COLUMN targeted_financing boolean NOT NULL DEFAULT false;
ALTER TABLE shopify_gateway_accounts ADD COLUMN targeted_financing_id varchar(128) DEFAULT '';
ALTER TABLE shopify_gateway_accounts ADD COLUMN targeted_financing_threshold bigint DEFAULT 0;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

ALTER TABLE shopify_gateway_accounts DROP COLUMN IF EXISTS targeted_financing;
ALTER TABLE shopify_gateway_accounts DROP COLUMN IF EXISTS targeted_financing_id;
ALTER TABLE shopify_gateway_accounts DROP COLUMN IF EXISTS targeted_financing_threshold;
