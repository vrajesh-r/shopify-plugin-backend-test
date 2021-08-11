
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

ALTER TABLE shopify_shops RENAME COLUMN splitpay_decline_auto_cancel TO remainder_pay_decline_auto_cancel;
ALTER TABLE shopify_gateway_accounts RENAME COLUMN splitpay_decline_auto_cancel TO remainder_pay_decline_auto_cancel;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

ALTER TABLE shopify_shops RENAME COLUMN remainder_pay_decline_auto_cancel TO splitpay_decline_auto_cancel;
ALTER TABLE shopify_gateway_accounts RENAME COLUMN remainder_pay_decline_auto_cancel TO splitpay_decline_auto_cancel;
