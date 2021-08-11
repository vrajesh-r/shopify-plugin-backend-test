package shopify

import "github.com/getbread/shopify_plugin_backend/service/spb_config"

//public config sets private variable scoped to package.
// (lower case vars and functions are private)
var shopifyConfig spb_config.ShopifyPluginBackendConfig

func ShopifyConfigInit(globalConfig spb_config.ShopifyPluginBackendConfig) {
	shopifyConfig = globalConfig
}
