
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE shopify_shops
(
	id			 uuid PRIMARY KEY UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
	shop		         varchar(128) UNIQUE NOT NULL,
	access_token	 	 varchar(128),
	bread_api_key		 varchar(128),	
	bread_secret_key 	 varchar(128),
	production 		 boolean DEFAULT false,
	auto_authorize		 boolean DEFAULT true,
	create_customers	 boolean DEFAULT false,
	auto_settle              boolean DEFAULT true,
	acts_as_label		 boolean DEFAULT true,
	css			 text,
	created_at 	 	 timestamp with time zone NOT NULL DEFAULT now(),
	updated_at 		 timestamp with time zone NOT NULL DEFAULT now()
);

CREATE TABLE shopify_shops_nonces
(
	id		uuid PRIMARY KEY UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
	shop_id 	uuid NOT NULL references shopify_shops,
	nonce 		varchar(128) NOT NULL,
	created_at 	timestamp with time zone NOT NULL DEFAULT now(),
	updated_at 	timestamp with time zone NOT NULL DEFAULT now()
);

CREATE TABLE shopify_shops_sessions
(
	id		uuid PRIMARY KEY UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
	shop_id 	uuid NOT NULL references shopify_shops,
	expiration	bigint NOT NULL,
	created_at 	timestamp with time zone NOT NULL DEFAULT now(),
	updated_at	timestamp with time zone NOT NULL DEFAULT now()
);

CREATE TABLE shopify_shops_orders
(
	id		uuid PRIMARY KEY UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
	shop_id      	uuid NOT NULL references shopify_shops,
	order_id  	bigint NOT NULL,
	tx_id  	  	uuid NOT NULL,
	production	boolean DEFAULT false
);


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE IF EXISTS shopify_shops_orders;
DROP TABLE IF EXISTS shopify_shops_nonces;
DROP TABLE IF EXISTS shopify_shops_sessions;
DROP TABLE IF EXISTS shopify_shops;

