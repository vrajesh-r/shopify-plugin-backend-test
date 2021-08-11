
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE shopify_analytics_orders
(
	id									uuid PRIMARY KEY UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
	order_id						bigint NOT NULL,
	shop_name						varchar(128),
	customer_id					bigint,
	customer_email			varchar(128),
	total_price					varchar(128),
	gateway							varchar(128),
	financial_status		varchar(128) DEFAULT '',
	fulfillment_status	varchar(128) DEFAULT '',
	test								boolean NOT NULL,
	redacted						boolean NOT NULL DEFAULT false,
	created_at					timestamp with time zone NOT NULL DEFAULT now(),
	updated_at					timestamp with time zone NOT NULL DEFAULT now()
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE IF EXISTS shopify_analytics_orders;
