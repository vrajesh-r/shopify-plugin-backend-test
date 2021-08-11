package routes

import (
	"time"

	_ "net/http/pprof"

	"github.com/getbread/breadkit/middleware/http"
	"github.com/getbread/shopify_plugin_backend/service/admin"
	"github.com/getbread/shopify_plugin_backend/service/app"
	"github.com/getbread/shopify_plugin_backend/service/gateway"
	sphealth "github.com/getbread/shopify_plugin_backend/service/health"
	"github.com/getbread/shopify_plugin_backend/service/spb_config"
	"github.com/gin-gonic/contrib/cors"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var routesConfig spb_config.ShopifyPluginBackendConfig

func RoutesConfigInit(globalConfig spb_config.ShopifyPluginBackendConfig) {
	routesConfig = globalConfig
}
func AttachRoutes(r *gin.Engine, appHandlers *app.Handlers, gatewayHandlers *gateway.Handlers, adminHandlers *admin.Handlers) {
	// CORS middleware
	corsConfig := cors.DefaultConfig()
	frontEndHost := routesConfig.FeHost
	if frontEndHost != "" {
		corsConfig.AllowCredentials = true
		corsConfig.AllowedOrigins = []string{frontEndHost}
		corsConfig.AllowAllOrigins = false
	}

	r.Use(cors.New(corsConfig))

	// bind routes for namespaces
	bindRootRoutes(r, appHandlers, gatewayHandlers)
	bindWebhookRoutes(r, appHandlers)
	bindProxyRoutes(r, appHandlers)
	bindPortalRoutes(r, appHandlers)
	bindGatewayRoutes(r, gatewayHandlers)
	bindCartsRoutes(r, appHandlers)
	bindAdminRoutes(r, adminHandlers)
	bindPOSRoutes(r, appHandlers)
	bindHealthRoutes(r, appHandlers)
}

func bindHealthRoutes(r *gin.Engine, h *app.Handlers) {
	queue := h.Requester.Queue()
	health := r.Group("/health")
	health.GET("/", http.TraceEndpoint(sphealth.Health, queue))
	health.GET("", http.TraceEndpoint(sphealth.Health, queue))
	health.GET("/live", http.TraceEndpoint(sphealth.Live, queue))
	health.GET("/ready", http.TraceEndpoint(sphealth.Ready, queue))

}
func bindRootRoutes(r *gin.Engine, h *app.Handlers, gh *gateway.Handlers) {
	queue := h.Requester.Queue()

	r.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/gateway")
	})
	r.GET("/install", http.TraceEndpoint(h.InstallApp, queue))
	r.GET("/static/:shopId/cart.js", http.TraceEndpoint(h.CartJS, queue))
	r.GET("/static/:shopId/cart_platform.js", http.TraceEndpoint(h.CartJS, queue))
	r.GET("/static/:gatewayKey/checkout.js", http.TraceEndpoint(gh.ShopifyPlusCheckoutJS, queue))
	if routesConfig.ServeFERoutes == true {
		r.StaticFile("/favicon.ico", "./build/gateway/assets/favicon.ico")
		r.StaticFile("/assets/no-image.gif", "./build/gateway/assets/no-image.gif")
	}
}

func bindWebhookRoutes(r *gin.Engine, h *app.Handlers) {
	logger := log.New()
	logger.SetFormatter(&log.JSONFormatter{})
	queue := h.Requester.Queue()
	webhooks := r.Group("/webhooks", ginrus.Ginrus(logger, time.RFC3339, true), http.TraceEndpoint(h.HttpAuthentication, queue), http.TraceEndpoint(h.VerifyHost, queue))

	webhooks.GET("/app", http.TraceEndpoint(h.AuthorizeApp, queue))
	webhooks.POST("/app/uninstall", http.TraceEndpoint(h.UninstallApp, queue))
	webhooks.POST("/orders", http.TraceEndpoint(h.OrderUpdated, queue))
	webhooks.POST("/orders/cancel", http.TraceEndpoint(h.CancelOrder, queue))
	webhooks.POST("/orders/create", http.TraceEndpoint(h.CreateOrder, queue))
	webhooks.POST("/orders/fulfilled", http.TraceEndpoint(h.OrdersFulfilled, queue))
	webhooks.POST("/orders/transactions", http.TraceEndpoint(h.NewOrderTransaction, queue))
	webhooks.POST("/redact/customer", http.TraceEndpoint(h.RedactCustomer, queue))
	webhooks.POST("/redact/shop", http.TraceEndpoint(h.RedactShop, queue))
	webhooks.POST("/customer/data-request", http.TraceEndpoint(h.CustomerDataRequest, queue))

	// Capture checkout data in Redis
	webhooks.POST("/checkouts/create", http.TraceEndpoint(h.PersistCheckout, queue))
	webhooks.POST("/checkouts/update", http.TraceEndpoint(h.PersistCheckout, queue))
}

// proxied routes need to be defined twice, once with trailing slash `/` & once without
// gin setting to auto try with a trailing slash won't suffice for this, because gin sends back a 302
// with a full fledged url and these requests are proxied
func bindProxyRoutes(r *gin.Engine, h *app.Handlers) {
	queue := h.Requester.Queue()
	proxy := r.Group("/proxy", http.TraceEndpoint(h.ProxyAuthentication, queue))

	proxy.POST("/orders", http.TraceEndpoint(h.CopyOrder, queue))
	proxy.POST("/orders/", http.TraceEndpoint(h.CopyOrder, queue))
	proxy.GET("/orders/confirmation/:order_id", http.TraceEndpoint(h.OrderConfirmation, queue))
	proxy.GET("/orders/confirmation/:order_id/", http.TraceEndpoint(h.OrderConfirmation, queue))
	proxy.GET("/carts/orders/confirmation", http.TraceEndpoint(h.CartOrderConfirmation, queue))
	proxy.GET("/carts/orders/confirmation/", http.TraceEndpoint(h.CartOrderConfirmation, queue))
	proxy.POST("/cart/shipping", http.TraceEndpoint(h.CartShippingOptions, queue))
	proxy.POST("/cart/shipping/", http.TraceEndpoint(h.CartShippingOptions, queue))
	proxy.POST("/cart/tax", http.TraceEndpoint(h.CartTaxTotal, queue))
	proxy.POST("/cart/tax/", http.TraceEndpoint(h.CartTaxTotal, queue))
	proxy.POST("/cart/tax/draftorder", http.TraceEndpoint(h.CartTaxTotalDraftOrder, queue))
	proxy.POST("/cart/tax/draftorder/", http.TraceEndpoint(h.CartTaxTotalDraftOrder, queue))

	proxy.POST("/errors", http.TraceEndpoint(h.LogFrontEndError, queue))
	proxy.POST("/errors/", http.TraceEndpoint(h.LogFrontEndError, queue))
}

func bindGatewayRoutes(r *gin.Engine, h *gateway.Handlers) {
	queue := h.Requester.Queue()
	gateway := r.Group("/gateway")

	// session endpoints
	gateway.GET("/session", http.TraceEndpoint(h.GatewayIsLoggedIn, queue))
	gateway.POST("/signup", http.TraceEndpoint(h.GatewaySignUp, queue))
	gateway.POST("/signin", http.TraceEndpoint(h.GatewaySignIn, queue))
	gateway.GET("/logout", http.TraceEndpoint(h.GatewayLogOut, queue))

	// account endpoints
	gateway.GET("/account", http.TraceEndpoint(h.PullGatewayAccount, queue))
	gateway.POST("/account/version", http.TraceEndpoint(h.UpdateGatewayAccountVersion, queue))
	gateway.POST("/account/password", http.TraceEndpoint(h.UpdateGatewayAccountPassword, queue))
	gateway.POST("/account/credentials", http.TraceEndpoint(h.RefreshGatewayCredentials, queue))
	gateway.POST("/account/:version", http.TraceEndpoint(h.UpdateGatewayAccount, queue))
	gateway.POST("/account/password/forgot", http.TraceEndpoint(h.ForgotGatewayAccountPassword, queue))
	gateway.POST("/account/password/reset", http.TraceEndpoint(h.ResetGatewayAccountPassword, queue))
	gateway.POST("/account/password/reset/validate", http.TraceEndpoint(h.ValidateResetToken, queue))

	// Hosted Payments SDK interface
	gateway.POST("/checkout", http.TraceEndpoint(h.GatewayCheckout, queue))
	gateway.POST("/checkout/test", http.TraceEndpoint(h.GatewayCheckout, queue))
	gateway.POST("/checkout/complete", http.TraceEndpoint(h.GatewayCheckoutComplete, queue))
	gateway.GET("/checkout/cancel", http.TraceEndpoint(h.GatewayCheckoutCancel, queue))
	gateway.GET("/checkout/confirmation", http.TraceEndpoint(h.GatewayCheckoutConfirmation, queue))
	gateway.GET("/checkout/confirmation-platform", http.TraceEndpoint(h.PlatformGatewayCheckoutConfirmation, queue))
	gateway.POST("/checkout/plus/record", http.TraceEndpoint(h.PlusGatewayTransactionRecord, queue))
	gateway.GET("/checkout/:CheckoutId", http.TraceEndpoint(h.GetCheckout, queue))
	gateway.GET("/product/images/:ShopName/:ShopId/:ProductId", http.TraceEndpoint(h.GetProductImage, queue))

	// Hosted Payments SDK order management
	gateway.POST("/orders/capture", http.TraceEndpoint(h.GatewayOrderManagement, queue))
	gateway.POST("/orders/cancel", http.TraceEndpoint(h.GatewayOrderManagement, queue))
	gateway.POST("/orders/refund", http.TraceEndpoint(h.GatewayOrderManagement, queue))

	// Milton hosted checkout routes

	gateway.GET("/tenant", h.GetTenant)
	gateway.GET("/checkout/fetch", http.TraceEndpoint(h.GetCheckout, queue))
	gateway.GET("/product/images/:ShopName/:ProductId", http.TraceEndpoint(h.GetProductImage, queue))
	if routesConfig.ServeFERoutes == true {
		// gateway static file server
		gateway.Static("/", "./build/gateway")
		//gateway.Static("/", "./shopify_plugin_backend/build/gateway")
	}

}

func bindPortalRoutes(r *gin.Engine, h *app.Handlers) {
	queue := h.Requester.Queue()
	portal := r.Group("/portal", http.TraceEndpoint(h.VerifyAppOAuthPermissionsUpToDate, queue))

	portal.GET("", http.TraceEndpoint(h.HttpAuthentication, queue), http.TraceEndpoint(h.AuthorizePortal, queue))
	portal.GET("/", http.TraceEndpoint(h.HttpAuthentication, queue), http.TraceEndpoint(h.AuthorizePortal, queue))
	portal.GET("/settings", http.TraceEndpoint(h.PortalSettings, queue))
	portal.GET("/settings/", http.TraceEndpoint(h.PortalSettings, queue))
	portal.POST("/settings/version", http.TraceEndpoint(h.UpdateVersion, queue))
	portal.POST("/settings/version/", http.TraceEndpoint(h.UpdateVersion, queue))
	portal.POST("/settings/:version", http.TraceEndpoint(h.UpdateSettings, queue))
	portal.POST("/settings/:version/", http.TraceEndpoint(h.UpdateSettings, queue))

	portal.GET("/draftorders", http.TraceEndpoint(h.GetDraftOrders, queue))
	portal.GET("/draftorders/", http.TraceEndpoint(h.GetDraftOrders, queue))
	portal.GET("/draftorder/:id", http.TraceEndpoint(h.ViewDraftOrder, queue))
	portal.GET("/draftorder/:id/", http.TraceEndpoint(h.ViewDraftOrder, queue))
	portal.POST("/draftorder/cart", http.TraceEndpoint(h.CreateDraftOrderCart, queue))
	portal.POST("/draftorder/cart/", http.TraceEndpoint(h.CreateDraftOrderCart, queue))
	portal.PUT("/draftorder/cart/:id", http.TraceEndpoint(h.UpdateDraftOrderCart, queue))
	portal.PUT("/draftorder/cart/:id/", http.TraceEndpoint(h.UpdateDraftOrderCart, queue))
	portal.POST("/draftorder/cart/email", http.TraceEndpoint(h.SendDraftOrderCartEmail, queue))
	portal.POST("/draftorder/cart/email/", http.TraceEndpoint(h.SendDraftOrderCartEmail, queue))
	portal.POST("/draftorder/cart/text", http.TraceEndpoint(h.SendDraftOrderCartText, queue))
	portal.POST("/draftorder/cart/text/", http.TraceEndpoint(h.SendDraftOrderCartText, queue))
	portal.POST("/draftorder/cart/callback", http.TraceEndpoint(h.DraftOrderCartCallback, queue))
	portal.POST("/draftorder/cart/callback/", http.TraceEndpoint(h.DraftOrderCartCallback, queue))
	portal.GET("/draftorder/cart/complete", http.TraceEndpoint(h.DraftOrderCartComplete, queue))
	portal.GET("/draftorder/cart/complete/", http.TraceEndpoint(h.DraftOrderCartComplete, queue))
	portal.GET("/draftorder/cart/error", http.TraceEndpoint(h.DraftOrderCartError, queue))
	portal.GET("/draftorder/cart/error/", http.TraceEndpoint(h.DraftOrderCartError, queue))

	portal.GET("/adminlinks/draftorders/details", http.TraceEndpoint(h.DraftOrderDetails, queue), http.TraceEndpoint(h.HttpAuthentication, queue), http.TraceEndpoint(h.AuthorizeDraftOrderView, queue))
	portal.GET("/adminlinks/draftorders/details/", http.TraceEndpoint(h.DraftOrderDetails, queue), http.TraceEndpoint(h.HttpAuthentication, queue), http.TraceEndpoint(h.AuthorizeDraftOrderView, queue))
}

// This set of routes are for links generated by merchants
func bindCartsRoutes(r *gin.Engine, h *app.Handlers) {
	queue := h.Requester.Queue()
	links := r.Group("/carts")

	links.POST("/orders", http.TraceEndpoint(h.SendOrderShopify, queue))
}

func bindAdminRoutes(r *gin.Engine, h *admin.Handlers) {
	queue := h.Requester.Queue()
	admin := r.Group("/admin")

	admin.GET("/", http.TraceEndpoint(h.AdminPortal, queue))
	admin.GET("/authenticate", http.TraceEndpoint(h.HandleOauthRedirect, queue))
	admin.GET("/logout", http.TraceEndpoint(h.PortalLogout, queue))
	admin.GET("/data", noCORS, http.TraceEndpoint(h.AuthMiddleware, queue), http.TraceEndpoint(h.GetShopifyShops, queue))
	admin.POST("/data/settings", noCORS, http.TraceEndpoint(h.AuthMiddleware, queue), http.TraceEndpoint(h.UpdateShopSettings, queue))
	admin.POST("/webhooks", http.TraceEndpoint(h.RegisterWebhooks, queue))

	// Static Assets
	public := admin.Group("/static/public")
	public.Static("/", "./build/admin/assets/public")

	private := admin.Group("/static/private", http.TraceEndpoint(h.AuthMiddleware, queue))
	private.Static("/", "./build/admin/assets/private")
}

func bindPOSRoutes(r *gin.Engine, h *app.Handlers) {
	queue := h.Requester.Queue()
	pos := r.Group("/pos", http.TraceEndpoint(h.HttpAuthentication, queue))

	pos.GET("/authorize", http.TraceEndpoint(h.RenderPOS, queue))
	pos.GET("/authorize/", http.TraceEndpoint(h.RenderPOS, queue))
}

func noCORS(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", routesConfig.HostConfig.MiltonHost)
	c.Writer.Header().Set("Vary", "Origin")
	c.Next()
}
