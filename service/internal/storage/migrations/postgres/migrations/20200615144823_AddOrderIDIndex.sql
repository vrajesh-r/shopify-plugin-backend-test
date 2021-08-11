
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE INDEX IF NOT EXISTS shopify_analytics_orders_order_id_idx ON shopify_analytics_orders (order_id);
CREATE INDEX IF NOT EXISTS shopify_analytics_orders_shop_name_idx ON shopify_analytics_orders (shop_name);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP INDEX IF EXISTS shopify_analytics_orders_order_id_idx;
DROP INDEX IF EXISTS shopify_analytics_orders_shop_name_idx;
