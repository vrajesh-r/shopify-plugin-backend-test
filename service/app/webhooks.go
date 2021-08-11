package app

import (
	"fmt"

	"github.com/getbread/shopify_plugin_backend/service/shopify"
	"github.com/getbread/shopify_plugin_backend/service/types"
	"github.com/sirupsen/logrus"
)

// Shopify webhooks used by Milton App
func getWebhookTopics(enableWebhooks bool) map[string]string {
	if !enableWebhooks {
		return map[string]string{"app/uninstalled": "/webhooks/app/uninstall"}
	}

	return map[string]string{
		"app/uninstalled":           "/webhooks/app/uninstall",
		"orders/updated":            "/webhooks/orders",
		"orders/cancelled":          "/webhooks/orders/cancel",
		"orders/create":             "/webhooks/orders/create",
		"order_transactions/create": "/webhooks/orders/transactions",
		"orders/fulfilled":          "/webhooks/orders/fulfilled",
		"checkouts/create":          "/webhooks/checkouts/create",
		"checkouts/update":          "/webhooks/checkouts/update",
	}
}

func createWebhookRequests(topics map[string]string) (reqs []*shopify.RegisterWebhookRequest) {
	for topic, path := range topics {
		reqs = append(reqs, &shopify.RegisterWebhookRequest{
			Webhook: shopify.MiniWebhook{
				Topic:   topic,
				Address: appConfig.HostConfig.MiltonHost + path,
				Format:  "json",
			},
		})
	}
	return
}

func updateWebhooks(shop types.Shop, deleteOldWebhooksFirst bool) (bool, error) {

	sc := shopify.NewClient(shop.Shop, shop.AccessToken)
	var response shopify.QueryWebhooksResponse
	if err := sc.QueryWebhook(sc.ShopName, &response); err != nil {
		return false, fmt.Errorf("Failed to query for webhooks of shop %s due to error: %s", sc.ShopName, err.Error())
	}

	topics := getWebhookTopics(shop.EnableOrderWebhooks)
	if deleteOldWebhooksFirst {
		for _, webhook := range response.Webhooks {
			if err := sc.DeleteWebhook(webhook.Id); err != nil {
				return false, fmt.Errorf("Failed to delete webhook %d of shop %s due to error: %s", webhook.Id, sc.ShopName, err.Error())
			}
		}
	} else {
		// Crude check, consider deep comparison
		if len(response.Webhooks) == len(topics) {
			return false, nil
		}

		// Remove redundant webhooks from the register webhook request
		for _, webhook := range response.Webhooks {
			delete(topics, webhook.Topic)
		}
	}

	requests := createWebhookRequests(topics)

	if errors := registerWebhooks(sc, requests); len(errors) > 0 {
		logrus.WithFields(logrus.Fields{
			"errors": fmt.Sprintf("%+v", errors),
			"shop":   fmt.Sprintf("%+v", shop),
		}).Info("(updateWebhooks) registering webhooks produced error")

		if len(errors) == 1 {
			return false, errors[0]
		} else {
			return false, fmt.Errorf("Failed to create new webhooks due to %d errors.", len(errors))
		}
	}

	return true, nil
}

func registerWebhooks(sc *shopify.Client, webhookReqs []*shopify.RegisterWebhookRequest) []error {

	var err error
	errChan := make(chan error, len(webhookReqs))
	for _, wr := range webhookReqs {
		go func(wr *shopify.RegisterWebhookRequest) {
			errChan <- sc.RegisterWebhook(wr, &shopify.RegisterWebhookResponse{})
		}(wr)
	}
	var errors []error
	for i := 0; i < cap(errChan); i++ {
		err = <-errChan
		if err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

func UpdateWebhooksExt(shop types.Shop, deleteOldWebhooksFirst bool) (bool, error) {

	sc := shopify.NewClient(shop.Shop, shop.AccessToken)
	var response shopify.QueryWebhooksResponse
	if err := sc.QueryWebhook(sc.ShopName, &response); err != nil {
		return false, fmt.Errorf("Failed to query for webhooks of shop %s due to error: %s", sc.ShopName, err.Error())
	}

	topics := getWebhookTopics(true)
	if deleteOldWebhooksFirst {
		for _, webhook := range response.Webhooks {
			if err := sc.DeleteWebhook(webhook.Id); err != nil {
				return false, fmt.Errorf("Failed to delete webhook %d of shop %s due to error: %s", webhook.Id, sc.ShopName, err.Error())
			}
		}
	} else {
		// Crude check, consider deep comparison
		if len(response.Webhooks) == len(topics) {
			return false, nil
		}

		// Remove redundant webhooks from the register webhook request
		for _, webhook := range response.Webhooks {
			delete(topics, webhook.Topic)
		}
	}

	requests := createWebhookRequests(topics)

	if errors := registerWebhooks(sc, requests); len(errors) > 0 {
		logrus.WithFields(logrus.Fields{
			"errors": fmt.Sprintf("%+v", errors),
			"shop":   fmt.Sprintf("%+v", shop),
		}).Info("(updateWebhooks) registering webhooks produced error")

		if len(errors) == 1 {
			return false, errors[0]
		} else {
			return false, fmt.Errorf("Failed to create new webhooks due to %d errors.", len(errors))
		}
	}

	return true, nil
}
