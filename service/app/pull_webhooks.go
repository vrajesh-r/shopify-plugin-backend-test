package app

import (
	"github.com/getbread/shopify_plugin_backend/service/shopify"
	"github.com/getbread/shopify_plugin_backend/service/types"
)

func PullWebhooks(shop types.Shop) (*[]shopify.Webhook, error) {
	sc := shopify.NewClient(shop.Shop, shop.AccessToken)
	var res shopify.QueryWebhooksResponse
	if err := sc.QueryWebhook(shop.Shop, &res); err != nil {
		return nil, err
	}
	return &res.Webhooks, nil
}
