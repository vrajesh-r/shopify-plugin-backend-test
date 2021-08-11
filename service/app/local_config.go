package app

import "github.com/getbread/shopify_plugin_backend/service/spb_config"

//public config sets private variable scoped to package.
// (lower case vars and functions are private)
var appConfig spb_config.ShopifyPluginBackendConfig

func AppConfigInit(globalConfig spb_config.ShopifyPluginBackendConfig) {
	appConfig = globalConfig
}
