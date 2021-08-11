
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE IF NOT EXISTS shopify_shops_draft_order_carts (
       id                       uuid                     PRIMARY KEY UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
       shop_id                	uuid		         NOT NULL,
       draft_order_id         	bigint    	         NOT NULL,
       cart_id                  uuid                     NOT NULL,
       cart_url                 text                     NOT NULL,
       is_production          	boolean                  NOT NULL,
       is_deleted 	        boolean     	         NOT NULL,
       use_draft_order_as_order boolean 	         NOT NULL DEFAULT true,
       created_at             	timestamp with time zone NOT NULL DEFAULT now(),
       updated_at             	timestamp with time zone NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS shopify_shops_draft_order_cart_checkouts (
       id                     uuid                     PRIMARY KEY UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
       shop_id		      uuid 		       NOT NULL,
       tx_id		      uuid                     NOT NULL,
       draft_order_cart_id    uuid                     NOT NULL,
       order_id		      bigint		       NOT NULL,
       completed 	      boolean		       NOT NULL DEFAULT FALSE,
       errored		      boolean		       NOT NULL DEFAULT FALSE,
       is_production 	      boolean 		       NOT NULL,
       created_at             timestamp with time zone NOT NULL DEFAULT now(),
       updated_at             timestamp with time zone NOT NULL DEFAULT now()
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE IF EXISTS shopify_shops_draft_order_cart_checkouts;
DROP TABLE IF EXISTS shopify_shops_draft_order_carts;
