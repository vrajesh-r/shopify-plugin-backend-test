
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

ALTER TABLE shopify_shops ADD COLUMN bread_sandbox_api_key varchar(128) DEFAULT '';
ALTER TABLE shopify_shops ADD COLUMN bread_sandbox_secret_key varchar(128) DEFAULT '';

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

ALTER TABLE shopify_shops DROP COLUMN IF EXISTS bread_sandbox_api_key;
ALTER TABLE shopify_shops DROP COLUMN IF EXISTS bread_sandbox_secret_key;
