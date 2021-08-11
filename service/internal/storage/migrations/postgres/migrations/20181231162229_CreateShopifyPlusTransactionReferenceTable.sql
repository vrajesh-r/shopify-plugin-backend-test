
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE IF NOT EXISTS shopify_plus_gateway_checkouts (
  id              uuid PRIMARY KEY UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
  checkout_id     varchar(128) NOT NULL,
  transaction_id  varchar(128) NOT NULL,
  created_at      timestamp with time zone NOT NULL DEFAULT now(),
  updated_at      timestamp with time zone NOT NULL DEFAULT now()
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE IF EXISTS shopify_plus_gateway_checkouts;
