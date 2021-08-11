
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE shopify_gateway_accounts ADD COLUMN integration_key varchar(128) DEFAULT '';
ALTER TABLE shopify_gateway_accounts ADD COLUMN sandbox_integration_key varchar(128) DEFAULT '';

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

ALTER TABLE shopify_gateway_accounts DROP COLUMN IF EXISTS integration_key;
ALTER TABLE shopify_gateway_accounts DROP COLUMN IF EXISTS sandbox_integration_key;