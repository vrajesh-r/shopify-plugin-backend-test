
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE shopify_gateway_accounts
(
	id			     uuid PRIMARY KEY UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
	email 			     varchar(128) UNIQUE NOT NULL,
	password_hash 		     text NOT NULL,	
	gateway_key		     varchar(128) NOT NULL DEFAULT '',
	gateway_secret 		     varchar(128) NOT NULL DEFAULT '',
	api_key		     	     varchar(128)  DEFAULT '',
	shared_secret 		     varchar(128)  DEFAULT '',
	sandbox_api_key 	     varchar(128)  DEFAULT '',
	sandbox_shared_secret 	     varchar(128)  DEFAULT '',
	production 		     boolean NOT NULL DEFAULT false,
	created_at		     timestamp with time zone NOT NULL DEFAULT now(),
	updated_at		     timestamp with time zone NOT NULL DEFAULT now()
);

CREATE TABLE shopify_gateway_checkouts
(
	id		uuid PRIMARY KEY UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
	account_id	uuid NOT NULL references shopify_gateway_accounts,
	transaction_id  varchar(128) NOT NULL DEFAULT '',
	test		boolean NOT NULL,
	reference	varchar(128) NOT NULL,
	currency	varchar(128) NOT NULL,
	amount		decimal NOT NULL,
	complete_url    varchar(128) NOT NULL,
	callback_url 	varchar(128) NOT NULL,
	cancel_url	varchar(128) NOT NULL,
	completed 	boolean NOT NULL DEFAULT false,
	errored		boolean NOT NULL DEFAULT false,
	created_at 	timestamp with time zone NOT NULL DEFAULT now(),
	updated_at 	timestamp with time zone NOT NULL DEFAULT now()
);

CREATE TABLE shopify_gateway_sessions
(
	id				uuid PRIMARY KEY UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
	gateway_account_id  	   	uuid NOT NULL references shopify_gateway_accounts,
	expiration 			bigint NOT NULL,
	created_at			timestamp with time zone NOT NULL DEFAULT now(),
	updated_at			timestamp with time zone NOT NULL DEFAULT now()
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE IF EXISTS shopify_gateway_sessions;
DROP TABLE IF EXISTS shopify_gateway_checkouts;
DROP TABLE IF EXISTS shopify_gateway_accounts;
