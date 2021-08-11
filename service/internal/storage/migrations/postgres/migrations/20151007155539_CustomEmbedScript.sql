
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

ALTER TABLE shopify_shops ADD COLUMN manual_embed_script boolean NOT NULL DEFAULT false;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

ALTER TABLE shopify_shops DROP COLUMN IF EXISTS manual_embed_script;

