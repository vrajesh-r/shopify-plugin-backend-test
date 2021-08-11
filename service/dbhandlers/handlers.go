package dbhandlers

import (
	"github.com/garyburd/redigo/redis"
	"github.com/getbread/breadkit/desmond/request"
	"github.com/jmoiron/sqlx"
)

type Handlers struct {
	DB                                  *sqlx.DB
	RedisPool                           *redis.Pool
	Requester                           *request.Requester
	ShopCreator                         ShopCreator
	ShopUpdater                         ShopUpdater
	ShopSearcher                        ShopSearcher
	NonceCreator                        NonceCreator
	NonceUpdater                        NonceUpdater
	NonceSearcher                       NonceSearcher
	SessionCreator                      SessionCreator
	SessionUpdater                      SessionUpdater
	SessionSearcher                     SessionSearcher
	OrderCreator                        OrderCreator
	OrderUpdater                        OrderUpdater
	OrderSearcher                       OrderSearcher
	GatewayCheckoutCreator              GatewayCheckoutCreator
	GatewayCheckoutUpdater              GatewayCheckoutUpdater
	GatewayCheckoutSearcher             GatewayCheckoutSearcher
	GatewayAccountCreator               GatewayAccountCreator
	GatewayAccountUpdater               GatewayAccountUpdater
	GatewayAccountSearcher              GatewayAccountSearcher
	GatewaySessionCreator               GatewaySessionCreator
	GatewaySessionUpdater               GatewaySessionUpdater
	GatewaySessionSearcher              GatewaySessionSearcher
	GatewayPasswordResetRequestCreator  GatewayPasswordResetRequestCreator
	GatewayPasswordResetRequestUpdater  GatewayPasswordResetRequestUpdater
	GatewayPasswordResetRequestSearcher GatewayPasswordResetRequestSearcher
	PlusGatewayCheckoutCreator          PlusGatewayCheckoutCreator
	PlusGatewayCheckoutUpdater          PlusGatewayCheckoutUpdater
	PlusGatewayCheckoutSearcher         PlusGatewayCheckoutSearcher
	DraftOrderCartSearcher              DraftOrderCartSearcher
	DraftOrderCartUpdater               DraftOrderCartUpdater
	DraftOrderCartCreator               DraftOrderCartCreator
	DraftOrderCartCheckoutSearcher      DraftOrderCartCheckoutSearcher
	DraftOrderCartCheckoutUpdater       DraftOrderCartCheckoutUpdater
	DraftOrderCartCheckoutCreator       DraftOrderCartCheckoutCreator
	AnalyticsOrderCreator               AnalyticsOrderCreator
	AnalyticsOrderUpdater               AnalyticsOrderUpdater
	AnalyticsOrderSearcher              AnalyticsOrderSearcher
	GiftCardOrderCreator                GiftCardOrderCreator
}

func NewHandlers(db *sqlx.DB, redisPool *redis.Pool, requester *request.Requester) *Handlers {
	return &Handlers{
		DB:                                  db,
		RedisPool:                           redisPool,
		Requester:                           requester,
		ShopCreator:                         newSqlShopCreator(db),
		ShopUpdater:                         newSqlShopUpdater(db),
		ShopSearcher:                        newSqlShopSearcher(db),
		NonceCreator:                        newSqlNonceCreator(db),
		NonceUpdater:                        newSqlNonceUpdater(db),
		NonceSearcher:                       newSqlNonceSearcher(db),
		SessionCreator:                      newSqlSessionCreator(db),
		SessionUpdater:                      newSqlSessionUpdater(db),
		SessionSearcher:                     newSqlSessionSearcher(db),
		OrderCreator:                        newSqlOrderCreator(db),
		OrderUpdater:                        newSqlOrderUpdater(db),
		OrderSearcher:                       newSqlOrderSearcher(db),
		GatewayCheckoutCreator:              newSqlGatewayCheckoutCreator(db),
		GatewayCheckoutUpdater:              newSqlGatewayCheckoutUpdater(db),
		GatewayCheckoutSearcher:             newSqlGatewayCheckoutSearcher(db),
		GatewayAccountCreator:               newSqlGatewayAccountCreator(db),
		GatewayAccountUpdater:               newSqlGatewayAccountUpdater(db),
		GatewayAccountSearcher:              newSqlGatewayAccountSearcher(db),
		GatewaySessionCreator:               newSqlGatewaySessionCreator(db),
		GatewaySessionUpdater:               newSqlGatewaySessionUpdater(db),
		GatewaySessionSearcher:              newSqlGatewaySessionSearcher(db),
		GatewayPasswordResetRequestCreator:  newSqlGatewayPasswordResetRequestCreator(db),
		GatewayPasswordResetRequestUpdater:  newSqlGatewayPasswordResetRequestUpdater(db),
		GatewayPasswordResetRequestSearcher: newSqlGatewayPasswordResetRequestSearcher(db),
		PlusGatewayCheckoutCreator:          newSqlPlusGatewayCheckoutCreator(db),
		PlusGatewayCheckoutUpdater:          newSqlPlusGatewayCheckoutUpdater(db),
		PlusGatewayCheckoutSearcher:         newSqlPlusGatewayCheckoutSearcher(db),
		DraftOrderCartSearcher:              newSqlDraftOrderCartSearcher(db),
		DraftOrderCartUpdater:               newSqlDraftOrderCartUpdater(db),
		DraftOrderCartCreator:               newSqlDraftOrderCartCreator(db),
		DraftOrderCartCheckoutSearcher:      newSqlDraftOrderCartCheckoutSearcher(db),
		DraftOrderCartCheckoutUpdater:       newSqlDraftOrderCartCheckoutUpdater(db),
		DraftOrderCartCheckoutCreator:       newSqlDraftOrderCartCheckoutCreator(db),
		AnalyticsOrderCreator:               newSqlAnalyticsOrderCreator(db),
		AnalyticsOrderUpdater:               newSqlAnalyticsOrderUpdater(db),
		AnalyticsOrderSearcher:              newSqlAnalyticsOrderSearcher(db),
		GiftCardOrderCreator:                newSqlGiftCardOrderCreator(db),
	}
}
