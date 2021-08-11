package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/getbread/gokit/metrics"
	"github.com/getbread/shopify_plugin_backend/service/dbhandlers"
	"github.com/getbread/shopify_plugin_backend/service/shopify"
	"github.com/opentracing/opentracing-go"
	"github.com/rs/zerolog"

	"github.com/garyburd/redigo/redis"
	"github.com/getbread/breadkit/desmond"
	"github.com/getbread/breadkit/desmond/request"
	"github.com/getbread/breadkit/yeast"
	"github.com/getbread/shopify_plugin_backend/service/admin"
	"github.com/getbread/shopify_plugin_backend/service/app"
	"github.com/getbread/shopify_plugin_backend/service/bread"
	"github.com/getbread/shopify_plugin_backend/service/gateway"
	"github.com/getbread/shopify_plugin_backend/service/mailer"
	"github.com/getbread/shopify_plugin_backend/service/routes"
	cfg "github.com/getbread/shopify_plugin_backend/service/spb_config"
	"github.com/getbread/shopify_plugin_backend/service/types"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"go.opencensus.io/plugin/ochttp"
)

var db *sqlx.DB
var appPath string
var redisPool *redis.Pool
var queue desmond.Queue
var appConfig cfg.ShopifyPluginBackendConfig

var ddTracer opentracing.Tracer
var ddCloser io.Closer
var ddMetrics metrics.Metrics

const (
	// LogStrKeyModule is for use with the logger as a key to specify the module name.
	LogStrKeyModule = "module"
	// LogStrKeyRecoveredValue is for use with the logger as a key to specify the value recovered from a Panic().
	LogStrKeyRecoveredValue = "recoveredValue"
	// LogStrKeyService is for use with the logger as a key to specify the service name.
	LogStrKeyService = "service"
)

func main() {
	defer ddCloser.Close()
	gin.SetMode("debug")

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	routes.RoutesConfigInit(appConfig)
	// instantiate handlers
	requester := request.NewRequester(queue)
	miltonHandlers := dbhandlers.NewHandlers(db, redisPool, requester)
	appHandlers := app.NewHandlers(miltonHandlers)
	gatewayHandlers := gateway.NewHandlers(miltonHandlers)
	adminHandlers := admin.NewHandlers(miltonHandlers)

	// load html templates into gin
	r.LoadHTMLGlob("service/cmd/shopify_plugin_backend/build/**/*.html")
	routes.AttachRoutes(r, appHandlers, gatewayHandlers, adminHandlers)

	var err error
	addr := fmt.Sprintf("0.0.0.0:%s", "8000")

	if appConfig.Environment == "local" {
		//only used for local development outside of slice
		err = http.ListenAndServeTLS(addr, "cmd/certs/localhost.pem", "cmd/certs/localhost-key.pem", r)
	} else if appConfig.Tracer != "" {
		err = http.ListenAndServe(addr, &ochttp.Handler{Handler: r})
	} else {
		err = http.ListenAndServe(addr, r)
	}
	if err != nil {
		logrus.WithError(err).Fatal("failed to listen on http server")
	}

	end := make(chan os.Signal)
	signal.Notify(end, syscall.SIGINT, syscall.SIGTERM)
	<-end

	log.Println("shutting down ...")
	return
}

func init() {
	flag.Parse()
	dir, err := os.Getwd()
	if err != nil {
		fmt.Print(err)
	}

	// If the .env variable IS_CLASSIC is true, use the .env, otherwise, use the config.yml

	//yeast expects a .env file and will exit with error if it doesn't find one.
	//temporary work around: have an empty "dummy file", while the envs are actually
	//set by this function.
	//TODO: refactor so we no longer use bread classic libraries like yeast
	if _, err := os.Stat(".env"); err == nil {
		logrus.Info("Using .env file")
		yeast.LoadEnvironment(".env")
		appConfig = cfg.LoadEnvConfig()
		queue, db = yeast.Rise("shopify-plugin-backend", yeast.WithEnvPath(dir+"/.env"))
	} else if os.IsNotExist(err) {
		logrus.Info("Using config.yml")
		appConfig = cfg.LoadCustomConfig()
		cfg.SetEnvsForYeastFromConfig(appConfig)
		//set the root directory, assume we are in ./service
		repoRoot := strings.TrimSuffix(dir, "/service")
		queue, db = yeast.Rise("shopify-plugin-backend", yeast.WithEnvPath(repoRoot+"/.dummy_env"))
	} else {
		logrus.Info(".env or config does or does not exist?")
	}

	z := zerolog.New(os.Stderr).With().Str(LogStrKeyService, "application").Timestamp().Logger()
	zlogger := z.With().Str(LogStrKeyModule, "main").Logger()
	logrus.Infof("%+v", appConfig)

	ddTracer, ddCloser = setupTracing(appConfig)

	ddMetrics = setupMetrics(appConfig, zlogger)
	ddMetrics.Incr("shopify-plugin-backend.init", []string{"environment:dev"})

	redisPool = yeast.NewRedisPool()

	flag.StringVar(&appPath, "app_path", "build/static", "path to the app assets")

	span := ddTracer.StartSpan("shopify-plugin-backend.init")
	defer span.Finish()

	// Initialize env variables and session store for Admin package
	x := appConfig.AdminAuth
	admin.InitConfig(x.OauthClientId, x.OauthClientSecret.Unmask(), x.AdminSessionSecret.Unmask(), appConfig.HostConfig.MiltonHost)
	routes.RoutesConfigInit(appConfig)
	bread.BreadConfigInit(appConfig)
	gateway.GatewayConfigInit(appConfig)
	types.TypesConfigInit(appConfig)
	app.AppConfigInit(appConfig)
	shopify.ShopifyConfigInit(appConfig)
	mailer.InitMailer(appConfig.SendgridClient)
}
