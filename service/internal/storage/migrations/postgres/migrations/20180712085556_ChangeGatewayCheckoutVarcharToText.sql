
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE shopify_gateway_checkouts
  ALTER COLUMN callback_url TYPE VARCHAR(4096),
  ALTER COLUMN cancel_url TYPE VARCHAR(4096),
  ALTER COLUMN complete_url TYPE VARCHAR(4096);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE shopify_gateway_checkouts
  ALTER COLUMN callback_url TYPE VARCHAR(128),
  ALTER COLUMN cancel_url TYPE VARCHAR(128),
  ALTER COLUMN complete_url TYPE VARCHAR(128);
