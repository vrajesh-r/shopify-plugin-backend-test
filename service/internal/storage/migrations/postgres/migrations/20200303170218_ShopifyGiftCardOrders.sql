
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE shopify_gift_card_orders
(
	id                       uuid PRIMARY KEY UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
	order_id                 bigint NOT NULL,
	shop_name                varchar(128) NOT NULL,
	gateway                  varchar(128) NOT NULL,
	test                     boolean NOT NULL,
	item_name                varchar(256) NOT NULL,
	item_price               varchar(128) NOT NULL,
	quantity                 smallint NOT NULL,
	requires_shipping        boolean NOT NULL,
	is_shopify_gift_card     boolean NOT NULL,
	name_contains_gift_only  boolean NOT NULL,
	name_contains_gift_card  boolean NOT NULL,
	created_at               timestamp with time zone NOT NULL DEFAULT now(),
	updated_at               timestamp with time zone NOT NULL DEFAULT now()
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE IF EXISTS shopify_gift_card_orders;
