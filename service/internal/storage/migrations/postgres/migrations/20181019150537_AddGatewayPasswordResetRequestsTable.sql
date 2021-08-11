
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE IF NOT EXISTS shopify_gateway_password_reset_requests (
  id            uuid PRIMARY KEY UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
  account_id    uuid NOT NULL references shopify_gateway_accounts,
  token_hash    varchar(128) NOT NULL DEFAULT '',
  expiration    bigint NOT NULL,
  created_at    timestamp with time zone NOT NULL DEFAULT now(),
  updated_at    timestamp with time zone NOT NULL DEFAULT now()
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE IF EXISTS shopify_gateway_password_reset_requests;