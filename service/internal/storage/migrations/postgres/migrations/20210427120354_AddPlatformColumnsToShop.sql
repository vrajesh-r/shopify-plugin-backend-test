
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE shopify_shops ADD COLUMN production_v2 boolean NOT NULL DEFAULT false;
ALTER TABLE shopify_shops ADD COLUMN bread_api_key_v2 varchar(128) NOT NULL DEFAULT '';
ALTER TABLE shopify_shops ADD COLUMN bread_secret_key_v2 varchar(128) NOT NULL DEFAULT '';
ALTER TABLE shopify_shops ADD COLUMN bread_sandbox_api_key_v2 varchar(128) NOT NULL DEFAULT '';
ALTER TABLE shopify_shops ADD COLUMN bread_sandbox_secret_key_v2 varchar(128) NOT NULL DEFAULT '';
ALTER TABLE shopify_shops ADD COLUMN active_version varchar(32) NOT NULL DEFAULT '';



-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE shopify_shops DROP COLUMN IF EXISTS production_v2;
ALTER TABLE shopify_shops DROP COLUMN IF EXISTS bread_api_key_v2;
ALTER TABLE shopify_shops DROP COLUMN IF EXISTS bread_secret_key_v2;
ALTER TABLE shopify_shops DROP COLUMN IF EXISTS bread_sandbox_api_key_v2;
ALTER TABLE shopify_shops DROP COLUMN IF EXISTS bread_sandbox_secret_key_v2;
ALTER TABLE shopify_shops DROP COLUMN IF EXISTS active_version;