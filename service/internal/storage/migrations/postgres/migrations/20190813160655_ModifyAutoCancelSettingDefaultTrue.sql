
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

ALTER TABLE shopify_shops ALTER COLUMN splitpay_decline_auto_cancel SET DEFAULT true;
ALTER TABLE shopify_gateway_accounts ALTER COLUMN splitpay_decline_auto_cancel SET DEFAULT true;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

ALTER TABLE shopify_shops ALTER COLUMN splitpay_decline_auto_cancel SET DEFAULT false;
ALTER TABLE shopify_gateway_accounts ALTER COLUMN splitpay_decline_auto_cancel SET DEFAULT false;
