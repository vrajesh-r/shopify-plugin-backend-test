package gateway

import "github.com/getbread/shopify_plugin_backend/service/spb_config"

//public config sets private variable scoped to package.
// (lower case vars and functions are private)
var gatewayConfig spb_config.ShopifyPluginBackendConfig

func GatewayConfigInit(globalConfig spb_config.ShopifyPluginBackendConfig) {
	gatewayConfig = globalConfig
}
