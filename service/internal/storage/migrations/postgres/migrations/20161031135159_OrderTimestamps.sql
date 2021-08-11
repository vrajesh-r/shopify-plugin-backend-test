
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE IF EXISTS shopify_shops_orders ADD COLUMN created_at timestamp with time zone NOT NULL DEFAULT now();
ALTER TABLE IF EXISTS shopify_shops_orders ADD COLUMN updated_at timestamp with time zone NOT NULL DEFAULT now();

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

ALTER TABLE shopify_shops_orders DROP COLUMN IF EXISTS created_at;
ALTER TABLE shopify_shops_orders DROP COLUMN IF EXISTS updated_at;

