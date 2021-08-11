
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

ALTER TABLE shopify_shops ALTER COLUMN acts_as_label SET DEFAULT false;
ALTER TABLE shopify_shops ALTER COLUMN as_low_as SET DEFAULT true;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

ALTER TABLE shopify_shops ALTER COLUMN acts_as_label SET DEFAULT true;
ALTER TABLE shopify_shops ALTER COLUMN as_low_as SET DEFAULT false;