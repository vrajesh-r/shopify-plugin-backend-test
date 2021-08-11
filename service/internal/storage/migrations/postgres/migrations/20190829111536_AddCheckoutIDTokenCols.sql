
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

ALTER TABLE shopify_analytics_orders ADD COLUMN checkout_id bigint;
ALTER TABLE shopify_analytics_orders ADD COLUMN checkout_token varchar(128);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

ALTER TABLE shopify_analytics_orders DROP COLUMN IF EXISTS checkout_id;
ALTER TABLE shopify_analytics_orders DROP COLUMN IF EXISTS checkout_token;
