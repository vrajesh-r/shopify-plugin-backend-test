
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

ALTER TABLE shopify_shops ADD COLUMN splitpay_decline_auto_cancel boolean NOT NULL DEFAULT false;
ALTER TABLE shopify_gateway_accounts ADD COLUMN splitpay_decline_auto_cancel boolean NOT NULL DEFAULT false;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

ALTER TABLE shopify_shops DROP COLUMN IF EXISTS splitpay_decline_auto_cancel;
ALTER TABLE shopify_gateway_accounts DROP COLUMN IF EXISTS splitpay_decline_auto_cancel;
