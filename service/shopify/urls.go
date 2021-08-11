package shopify

import (
	"fmt"
	"net/url"
)

func getApiVersion() string {
	return shopifyConfig.ShopifyConfig.ShopifyApiVersion
}

func CreateDraftOrderUrl(shopName string) string {
	format := "https://%s.myshopify.com/admin/draft_orders.json"
	return fmt.Sprintf(format, shopName)
}

func GetDraftOrdersCountUrl(shopName string, query url.Values) string {
	format := "https://%s.myshopify.com/admin/draft_orders/count.json?%s"
	return fmt.Sprintf(format, shopName, query.Encode())
}

func GetDraftOrdersUrl(shopName string, query url.Values) string {
	format := "https://%s.myshopify.com/admin/draft_orders.json?%s"
	return fmt.Sprintf(format, shopName, query.Encode())
}

func GetDraftOrderUrl(shopName, id string) string {
	format := "https://%s.myshopify.com/admin/draft_orders/%s.json"
	return fmt.Sprintf(format, shopName, id)
}

func CompleteDraftOrderUrl(shopName, id string) string {
	format := "https://%s.myshopify.com/admin/draft_orders/%s/complete.json"
	return fmt.Sprintf(format, shopName, id)
}

func DeleteDraftOrderUrl(shopName, id string) string {
	format := "https://%s.myshopify.com/admin/draft_orders/%s.json"
	return fmt.Sprintf(format, shopName, id)
}

func EmbedScriptUrl(shopName string) string {
	format := "https://%s.myshopify.com/admin/script_tags.json"
	return fmt.Sprintf(format, shopName)
}

func OAuthExchangeUrl(shopName string) string {
	format := "https://%s.myshopify.com/admin/oauth/access_token"
	return fmt.Sprintf(format, shopName)
}

func AppAdminUrl(shopName string) string {
	format := "https://%s.myshopify.com/admin/apps/%s"
	return fmt.Sprintf(format, shopName, shopifyConfig.ShopifyConfig.ShopifyApiKey.Unmask())
}

func WebhookUrl(shopName string) string {
	return fmt.Sprintf("https://%s.myshopify.com/admin/webhooks.json", shopName)
}

func SingleWebhookUrl(shopName string, id int) string {
	return fmt.Sprintf("https://%s.myshopify.com/admin/webhooks/%d.json", shopName, id)
}

func LocationUrl(shopName string) string {
	format := "https://%s.myshopify.com/admin/locations.json"
	return fmt.Sprintf(format, shopName)
}

func InstallUrl(shopName, nonce, redirectPath string) string {
	format := "https://%s.myshopify.com/admin/oauth/authorize?client_id=%s&scope=%s&redirect_uri=%s&state=%s"
	redirectUri := shopifyConfig.HostConfig.MiltonHost + redirectPath
	scopes := "read_products,write_products,read_customers,write_customers,read_orders,write_orders,read_script_tags,write_script_tags,read_draft_orders,write_draft_orders" // pass these in
	return fmt.Sprintf(format, shopName, shopifyConfig.ShopifyConfig.ShopifyApiKey.Unmask(), scopes, redirectUri, nonce)
}

func CreateOrderUrl(shopName string) string {
	apiVersion := getApiVersion()
	if apiVersion == "" {
		return fmt.Sprintf("https://%s.myshopify.com/admin/orders.json", shopName)
	}

	return fmt.Sprintf("https://%s.myshopify.com/admin/api/%s/orders.json", shopName, apiVersion)
}

func CancelOrderUrl(shopName string, orderId string) string {
	apiVersion := getApiVersion()
	if apiVersion == "" {
		return fmt.Sprintf("https://%s.myshopify.com/admin/orders/%s/cancel.json", shopName, orderId)
	}

	return fmt.Sprintf("https://%s.myshopify.com/admin/api/%s/orders/%s/cancel.json", shopName, apiVersion, orderId)
}

func UpdateOrderUrl(shopName string, orderId string) string {
	apiVersion := getApiVersion()
	if apiVersion == "" {
		return fmt.Sprintf("https://%s.myshopify.com/admin/orders/%s.json", shopName, orderId)
	}

	return fmt.Sprintf("https://%s.myshopify.com/admin/api/%s/orders/%s.json", shopName, apiVersion, orderId)
}

func CreateCustomerUrl(shopName string) string {
	format := "https://%s.myshopify.com/admin/customers.json"
	return fmt.Sprintf(format, shopName)
}

func CreateTransactionUrl(shopName, orderId string) string {
	format := "https://%s.myshopify.com/admin/orders/%s/transactions.json"
	return fmt.Sprintf(format, shopName, orderId)
}

func SearchCustomerUrl(shopName string) string {
	format := "https://%s.myshopify.com/admin/customers/search.json"
	return fmt.Sprintf(format, shopName)
}

func SearchTransactionsUrl(shopName, orderId string) string {
	return fmt.Sprintf("https://%s.myshopify.com/admin/orders/%s/transactions.json", shopName, orderId)
}

func SearchTransactionUrl(shopName, orderId, transactionId string) string {
	format := "https://%s.myshopify.com/admin/orders/%s/transactions/%s.json"
	return fmt.Sprintf(format, shopName, orderId, transactionId)
}

func SearchOrderUrl(shopName, orderId string) string {
	apiVersion := getApiVersion()
	if apiVersion == "" {
		return fmt.Sprintf("https://%s.myshopify.com/admin/orders/%s.json", shopName, orderId)
	}

	return fmt.Sprintf("https://%s.myshopify.com/admin/api/%s/orders/%s.json", shopName, apiVersion, orderId)
}

func SearchProductByIdUrl(shopName, productId string) string {
	format := "https://%s.myshopify.com/admin/products/%s.json"
	return fmt.Sprintf(format, shopName, productId)
}

func SearchProductVariantUrl(shopName, variantId string) string {
	apiVersion := getApiVersion()
	if apiVersion == "" {
		return fmt.Sprintf("https://%s.myshopify.com/admin/variants/%s.json", shopName, variantId)
	}

	return fmt.Sprintf("https://%s.myshopify.com/admin/api/%s/variants/%s.json", shopName, apiVersion, variantId)
}

func CartUrl(shopName string) string {
	format := "https://%s.myshopify.com/cart.js"
	return fmt.Sprintf(format, shopName)
}

func AddToCartUrl(shopName string) string {
	format := "https://%s.myshopify.com/cart/add.js"
	return fmt.Sprintf(format, shopName)
}

func ClearCartUrl(shopName string) string {
	format := "https://%s.myshopify.com/cart/clear.js"
	return fmt.Sprintf(format, shopName)
}

func ShippingRatesUrl(shopName, zip, country, province string) string {
	format := "https://%s.myshopify.com/cart/shipping_rates.json?"
	u := fmt.Sprintf(format, shopName)
	v := url.Values{}
	v.Add("shipping_address[zip]", zip)
	v.Add("shipping_address[country]", country)
	v.Add("shipping_address[province]", province)
	u += v.Encode()
	return u
}

func CartCheckoutUrl(shopName string) string {
	format := "https://%s.myshopify.com/cart"
	return fmt.Sprintf(format, shopName) + "?note=&checkout=Check+Out"
}

func ShopUrl(shopName string) string {
	format := "https://%s.myshopify.com/admin/shop.json"
	return fmt.Sprintf(format, shopName)
}

func SearchEmbeddedScriptUrl(shopName string) string {
	format := "https://%s.myshopify.com/admin/script_tags.json"
	return fmt.Sprintf(format, shopName)
}

func DeleteEmbeddedScriptUrl(shopName, scriptID string) string {
	format := "https://%s.myshopify.com/admin/script_tags/%s.json"
	return fmt.Sprintf(format, shopName, scriptID)
}

func GetCheckoutURL(shopName, checkoutToken string) string {
	apiVersion := getApiVersion()
	return fmt.Sprintf("https://%s.myshopify.com/admin/api/%s/checkouts/%s.json", shopName, apiVersion, checkoutToken)
}
