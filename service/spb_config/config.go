package spb_config

import (
	"os"
	"strconv"

	"github.com/getbread/gokit/config"
	"github.com/sirupsen/logrus"
)

type ShopifyPluginBackendConfig struct {
	Environment   string
	Dockerfile    string
	Host          string
	FeHost        string
	ServeFERoutes bool
	GatewayTenant string
	DbExposedPort string
	MiltonPort    string
	HeapAppId     string
	DataDogToken  config.MaskedString
	DatadogSite   string
	Tracer        string

	//real stuff
	AvalaraKey                     config.MaskedString
	MiltonGatewayAutoSettleTimeout int
	Postgres                       config.Postgres
	HostConfig                     HostConfig
	ShopifyConfig                  ShopifyConfig
	SendgridClient                 SendgridConfig
	FeatureFlag                    config.FeatureFlags
	AdminAuth                      AdminAuthConfig
	TransactionService             TransactionServiceConfig
	RedisConfig                    RedisConfig
	Service                        ServiceConfig
	Tracing                        TracingConfig
	Metrics                        MetricsConfig
}

type TracingConfig struct {
	Name     string
	Endpoint string
	Port     string
}

type MetricsConfig struct {
	Name     string
	Endpoint string
	Port     string
}

type ServiceConfig struct {
	Name           string
	Sandbox        bool
	TracingEnabled bool
}

type HostConfig struct {
	MiltonHost                      string
	BreadHost                       string
	BreadHostDevelopment            string
	CheckoutHost                    string
	CheckoutHostDevelopment         string
	CheckoutHostLocal               string
	PlatformCheckoutHost            string
	PlatformCheckoutHostDevelopment string
}
type ShopifyConfig struct {
	ShopifyApiVersion   string
	ShopifyApiKey       config.MaskedString
	ShopifySharedSecret config.MaskedString
}
type SendgridConfig struct {
	SendgridMiltonApiKey                config.MaskedString
	SendgridMiltonPasswordResetTemplate string
}

type LaunchdarklyConfig struct {
	LaunchDarklyKey          config.MaskedString
	LaunchdarklyClientSideId config.MaskedString
}
type AdminAuthConfig struct {
	OauthClientId      string
	OauthClientSecret  config.MaskedString
	AdminSessionSecret config.MaskedString
}
type TransactionServiceConfig struct {
	TransactionServiceHost            string
	TransactionServiceHostDevelopment string
}
type RedisConfig struct {
	// no redis ENV variables used in code. leaving empty
	URL      string
	Password config.MaskedString
}

func IsClassicConfig() bool {
	isClassic := config.LookupEnvOrDefault("IS_CLASSIC", "false")
	result, _ := strconv.ParseBool(isClassic)
	return result
}

func LoadEnvConfig() ShopifyPluginBackendConfig {
	var cfg ShopifyPluginBackendConfig

	logrus.Info(os.Getwd())

	cfg.Environment = config.LookupEnvOrDefault("ENV", "development")
	cfg.Dockerfile = config.LookupEnvOrDefault("Dockerfile", "scripts/chef_build/Dockerfile-fullstack")
	cfg.Host = config.LookupEnvOrDefault("HOST", "")
	cfg.FeHost = config.LookupEnvOrDefault("FE_HOST", "")
	// cfg.ServeFERoutes = config.LookupEnvOrDefault("CONFIG_FILE_NAME", "") // (?)
	cfg.GatewayTenant = config.LookupEnvOrDefault("GATEWAY_TENANT", "")
	cfg.DbExposedPort = config.LookupEnvOrDefault("DB_EXPOSED_PORT", "")
	cfg.MiltonPort = config.LookupEnvOrDefault("MILTON_PORT", "")
	cfg.HeapAppId = config.LookupEnvOrDefault("HEAP_APP_ID", "")
	cfg.DataDogToken = config.MaskedString(config.LookupEnvOrDefault("DATADOG_CLIENT_TOKEN", ""))
	cfg.DatadogSite = config.LookupEnvOrDefault("DATADOG_SITE", "")
	cfg.Tracer = config.LookupEnvOrDefault("TRACER", "datadog")

	cfg.AvalaraKey = config.MaskedString(config.LookupEnvOrDefault("AVALARA_KEY", ""))
	cfg.MiltonGatewayAutoSettleTimeout, _ = strconv.Atoi(config.LookupEnvOrDefault("MILTON_GATEWAY_AUTO_SETTLE_TIMEOUT", ""))

	cfg.Postgres.Host = config.LookupEnvOrDefault("DB_HOST", "")
	cfg.Postgres.Port = config.LookupEnvOrDefault("DB_PORT", "")
	cfg.Postgres.User = config.LookupEnvOrDefault("DB_USER", "")
	cfg.Postgres.Password = config.MaskedString(config.LookupEnvOrDefault("DB_PASSWORD", ""))
	cfg.Postgres.Database = config.LookupEnvOrDefault("DB_DATABASE", "")

	cfg.HostConfig.BreadHost = config.LookupEnvOrDefault("BREAD_HOST", "")
	cfg.HostConfig.BreadHostDevelopment = config.LookupEnvOrDefault("BREAD_HOST_DEVELOPMENT", "")
	cfg.HostConfig.CheckoutHost = config.LookupEnvOrDefault("CHECKOUT_HOST", "")
	cfg.HostConfig.CheckoutHostDevelopment = config.LookupEnvOrDefault("CHECKOUT_HOST_DEVELOPMENT", "")
	cfg.HostConfig.CheckoutHostLocal = config.LookupEnvOrDefault("CHECKOUT_HOST_LOCAL", "")
	cfg.HostConfig.PlatformCheckoutHost = config.LookupEnvOrDefault("PLATFORM_CHECKOUT_HOST", "")
	cfg.HostConfig.PlatformCheckoutHostDevelopment = config.LookupEnvOrDefault("PLATFORM_CHECKOUT_HOST_DEVELOPMENT", "")

	cfg.ShopifyConfig.ShopifyApiKey = config.MaskedString(config.LookupEnvOrDefault("SHOPIFY_API_KEY", ""))
	cfg.ShopifyConfig.ShopifyApiVersion = config.LookupEnvOrDefault("SHOPIFY_API_VERSION", "")
	cfg.ShopifyConfig.ShopifySharedSecret = config.MaskedString(config.LookupEnvOrDefault("SHOPIFY_SHARED_SECRET", ""))

	cfg.ServeFERoutes, _ = strconv.ParseBool(config.LookupEnvOrDefault("SERVE_FE_ROUTES", "false"))
	cfg.SendgridClient.SendgridMiltonApiKey = config.MaskedString(config.LookupEnvOrDefault("SENDGRID_MILTON_API_KEY", ""))
	cfg.SendgridClient.SendgridMiltonPasswordResetTemplate = config.LookupEnvOrDefault("SENDGRID_MILTON_PASSWORD_RESET_TEMPLATE", "")

	cfg.FeatureFlag.ClientInitTimeout = 10
	cfg.FeatureFlag.ClientSideKey = config.LookupEnvOrDefault("LAUNCHDARKLY_KEY", "")
	cfg.FeatureFlag.ClientSideID = config.LookupEnvOrDefault("LAUNCHDARKLY_CLIENT_SIDE_ID", "")

	cfg.AdminAuth.AdminSessionSecret = config.MaskedString(config.LookupEnvOrDefault("MILTON_ADMIN_SESSION_SECRET", ""))
	cfg.AdminAuth.OauthClientSecret = config.MaskedString(config.LookupEnvOrDefault("MILTON_OAUTH_CLIENT_SECRET", ""))
	cfg.AdminAuth.OauthClientId = config.LookupEnvOrDefault("MILTON_OAUTH_CLIENT_ID", "")

	cfg.TransactionService.TransactionServiceHost = config.LookupEnvOrDefault("TRANSACTION_SERVICE_HOST", "")
	cfg.TransactionService.TransactionServiceHostDevelopment = config.LookupEnvOrDefault("TRANSACTION_SERVICE_HOST_DEVELOPMENT", "")

	cfg.RedisConfig.URL = config.LookupEnvOrDefault("REDIS_URL", "")
	cfg.RedisConfig.Password = config.MaskedString(config.LookupEnvOrDefault("REDIS_PASSWORD", ""))

	return cfg
}

func LoadCustomConfig() ShopifyPluginBackendConfig {
	var cfg ShopifyPluginBackendConfig
	//useful for debugging
	logrus.Info(os.Getwd())

	configFileName := config.LookupEnvOrDefault("CONFIG_FILE_NAME", "config")
	configFilePath := config.LookupEnvOrDefault("CONFIG_FILE_PATH", "../deploy/chart/")
	envName := config.LookupEnvOrDefault("ENVIRONMENT", "local")

	err := config.Get(configFilePath, configFileName, envName, "yaml", &cfg)

	if err != nil {

		logrus.Fatal("config not found! exiting")
		os.Exit(1)
	}
	return cfg
}

func SetEnvsForYeastFromConfig(cfg ShopifyPluginBackendConfig) {

	os.Setenv("DB_USER", cfg.Postgres.User)
	os.Setenv("DB_PASSWORD", cfg.Postgres.Password.Unmask())
	os.Setenv("DB_HOST", cfg.Postgres.Host)
	os.Setenv("DB_PORT", cfg.Postgres.Port)
	os.Setenv("DB_DATABASE", cfg.Postgres.Database)

	os.Setenv("DB_MAX_IDLE_CONNS", "10")
	os.Setenv("DB_MAX_OPEN_CONNS", "10")
	//this one won't break anything if missed, but easy to set here:
	os.Setenv("DEFAULT_LOG_LEVEL", "DEBUG")

	os.Setenv("LAUNCHDARKLY_KEY", cfg.FeatureFlag.ClientSideKey)
	os.Setenv("LAUNCHDARKLY_CLIENT_SIDE_ID", cfg.FeatureFlag.ClientSideID)

	os.Setenv("REDIS_URL", cfg.RedisConfig.URL)
	os.Setenv("REDIS_PASSWORD", cfg.RedisConfig.Password.Unmask())

	os.Setenv("TRACER", cfg.Tracer)
	var fullAddr = cfg.Tracing.Endpoint + ":" + cfg.Tracing.Port
	os.Setenv("TRACER_ENDPOINT", fullAddr)
}
