package bread

import "github.com/getbread/shopify_plugin_backend/service/spb_config"

//public config sets private variable scoped to package.
// (lower case vars and functions are private)
var breadConfig spb_config.ShopifyPluginBackendConfig

func BreadConfigInit(globalConfig spb_config.ShopifyPluginBackendConfig) {
	breadConfig = globalConfig
}
