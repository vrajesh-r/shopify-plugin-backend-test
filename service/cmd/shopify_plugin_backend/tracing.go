package main

import (
	"io"

	"github.com/getbread/gokit/metrics"
	"github.com/getbread/gokit/metrics/datadog"
	"github.com/getbread/gokit/tracing"
	cfg "github.com/getbread/shopify_plugin_backend/service/spb_config"
	"github.com/opentracing/opentracing-go"
	"github.com/rs/zerolog"
)

func setupTracing(config cfg.ShopifyPluginBackendConfig) (opentracing.Tracer, io.Closer) {
	if !config.Service.TracingEnabled {
		return tracing.Noop()
	}

	return tracing.NewDataDog(config.Service.Name,
		tracing.TraceEndpoint(config.Tracing.Endpoint, config.Tracing.Port),
		tracing.MetricEndpoint(config.Metrics.Endpoint, config.Metrics.Port))
}

func setupMetrics(config cfg.ShopifyPluginBackendConfig, logger zerolog.Logger) metrics.Metrics {
	if !config.Service.TracingEnabled {
		return metrics.Noop{}
	}

	return datadog.New(
		datadog.EnableLiveMode(),
		datadog.Logger(logger),
		datadog.Endpoint(config.Metrics.Endpoint, config.Metrics.Port),
	)
}
