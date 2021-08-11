
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

ALTER TABLE shopify_gateway_accounts ADD COLUMN api_key_v2 varchar(128) DEFAULT '';
ALTER TABLE shopify_gateway_accounts ADD COLUMN shared_secret_v2 varchar(128) DEFAULT '';
ALTER TABLE shopify_gateway_accounts ADD COLUMN sandbox_api_key_v2 varchar(128) DEFAULT '';
ALTER TABLE shopify_gateway_accounts ADD COLUMN sandbox_shared_secret_v2 varchar(128) DEFAULT '';
ALTER TABLE shopify_gateway_accounts ADD COLUMN auto_settle_v2 boolean NOT NULL DEFAULT false;
ALTER TABLE shopify_gateway_accounts ADD COLUMN active_version varchar(32) NOT NULL DEFAULT '';


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

ALTER TABLE shopify_gateway_accounts DROP COLUMN IF EXISTS api_key_v2;
ALTER TABLE shopify_gateway_accounts DROP COLUMN IF EXISTS shared_secret_v2;
ALTER TABLE shopify_gateway_accounts DROP COLUMN IF EXISTS sandbox_api_key_v2;
ALTER TABLE shopify_gateway_accounts DROP COLUMN IF EXISTS sandbox_shared_secret_v2;
ALTER TABLE shopify_gateway_accounts DROP COLUMN IF EXISTS auto_settle_v2;
ALTER TABLE shopify_gateway_accounts DROP COLUMN IF EXISTS active_version;
