// Package otel initializes OpenTelemetry trace export for the application and
// exposes small noop-safe helpers used by every flow.
//
// Two tracer providers are created: the process-global one (used by the web
// API and Telegram interactions) samples at the configured rate, while the
// collector provider always samples. When no TRACING_URL is configured the
// SDK is not started at all and every helper is a no-op, so instrumentation
// can be sprinkled everywhere without cost.
package otel

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/fmotalleb/north_outage/config"
)

const serviceName = "north_outage"

// collectorTP is the always-sampled provider used by the collector flow.
// It is nil until Init registers one; when tracing is disabled it stays
// nil and the accessors below fall back to a noop provider.
var collectorTP *sdktrace.TracerProvider

// Init creates the exporters and tracer providers and registers the
// rate-sampled provider as the process-global one. It returns a shutdown
// function that flushes and closes every provider; call it exactly once
// before the process exits. When cfg.URL is empty it returns a noop shutdown
// without creating anything.
func Init(ctx context.Context, cfg config.Tracing) (shutdown func(context.Context) error, err error) {
	if strings.TrimSpace(cfg.URL) == "" {
		return func(context.Context) error { return nil }, nil
	}

	u, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid tracing url %q: %w", cfg.URL, err)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("invalid tracing url %q: missing host", cfg.URL)
	}
	res := resource.Default()

	res, err = resource.Merge(resource.Default(), resource.NewWithAttributes(
		res.SchemaURL(),
		semconv.ServiceName(serviceName),
	))
	if err != nil {
		return nil, err
	}
	// The exporters only need a short-lived context to connect; never derive
	// a cancellable child of the app's shared context here.
	initCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()

	exp, err := newExporter(initCtx, u, map[string]string(cfg.Headers))
	if err != nil {
		return nil, err
	}

	appTP := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(clampRate(cfg.Rate)))),
		sdktrace.WithResource(res),
	)
	collectorTP = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(appTP)

	return func(ctx context.Context) error {
		var errs []error
		if err := appTP.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
		if err := collectorTP.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
		return errors.Join(errs...)
	}, nil
}

// Tracer returns the process-global tracer used by the web API and Telegram
// interactions (rate-sampled). It is a no-op tracer when tracing is disabled.
func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

// CollectorTracer returns a tracer backed by the always-sampled collector
// provider, or a no-op tracer when tracing is disabled.
func CollectorTracer(name string) trace.Tracer {
	if collectorTP == nil {
		return noop.NewTracerProvider().Tracer(name)
	}
	return collectorTP.Tracer(name)
}

// CollectorProvider returns the always-sampled tracer provider, or a noop
// provider when tracing is disabled. It is used to instrument the collector's
// outbound HTTP client so its spans stay outside the sampling rate.
func CollectorProvider() trace.TracerProvider {
	if collectorTP == nil {
		return noop.NewTracerProvider()
	}
	return collectorTP
}

func newExporter(ctx context.Context, u *url.URL, headers map[string]string) (sdktrace.SpanExporter, error) {
	switch strings.ToLower(u.Scheme) {
	case "http":
		return newHTTPExporter(ctx, u, headers, true)
	case "https":
		return newHTTPExporter(ctx, u, headers, false)
	case "grpc":
		return newGRPCExporter(ctx, u, headers, true)
	case "grpcs":
		return newGRPCExporter(ctx, u, headers, false)
	default:
		return nil, fmt.Errorf("unsupported tracing url scheme %q (supported: http, https, grpc, grpcs)", u.Scheme)
	}
}

func newHTTPExporter(ctx context.Context, u *url.URL, headers map[string]string, insecure bool) (*otlptrace.Exporter, error) {
	opts := []otlptracehttp.Option{otlptracehttp.WithEndpoint(u.Host)}
	if insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}
	if p := strings.TrimPrefix(u.Path, "/"); p != "" {
		opts = append(opts, otlptracehttp.WithURLPath("/"+p))
	}
	if len(headers) != 0 {
		opts = append(opts, otlptracehttp.WithHeaders(headers))
	}
	return otlptracehttp.New(ctx, opts...)
}

func newGRPCExporter(ctx context.Context, u *url.URL, headers map[string]string, insecure bool) (*otlptrace.Exporter, error) {
	opts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(u.Host)}
	if insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}
	if len(headers) != 0 {
		opts = append(opts, otlptracegrpc.WithHeaders(headers))
	}
	return otlptracegrpc.New(ctx, opts...)
}

func clampRate(rate float64) float64 {
	if rate < 0 {
		return 0
	}
	if rate > 1 {
		return 1
	}
	return rate
}
