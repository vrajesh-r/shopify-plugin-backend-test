
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

ALTER TABLE shopify_shops ADD COLUMN allow_checkout_pdp boolean NOT NULL DEFAULT true;
ALTER TABLE shopify_shops ADD COLUMN allow_checkout_cart boolean NOT NULL DEFAULT true;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

ALTER TABLE shopify_shops DROP COLUMN IF EXISTS allow_checkout_pdp;
ALTER TABLE shopify_shops DROP COLUMN IF EXISTS allow_checkout_cart;
