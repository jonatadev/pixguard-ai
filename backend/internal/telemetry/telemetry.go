package telemetry

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/jonatadev/pixguard-ai/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace/noop"
)

// NewLogger returns a structured JSON logger for correlation with Loki and OTel trace IDs.
func NewLogger(env string) *slog.Logger {
	level := slog.LevelInfo
	if strings.EqualFold(env, "development") {
		level = slog.LevelDebug
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}

// Metrics holds isolated Prometheus metrics to prevent duplicate global registration panics in tests.
type Metrics struct {
	Registry                *prometheus.Registry
	HTTPRequestsTotal       *prometheus.CounterVec
	HTTPRequestDuration     *prometheus.HistogramVec
	GuardrailPIIRedactions  *prometheus.CounterVec
	GuardrailInjectionBlock prometheus.Counter
}

func NewMetrics() *Metrics {
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	httpRequests := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pixguard_http_requests_total",
			Help: "Total HTTP requests processed by PixGuard AI API.",
		},
		[]string{"method", "path", "status"},
	)

	httpDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pixguard_http_request_duration_seconds",
			Help:    "HTTP request latency distribution in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	piiRedactions := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pixguard_guardrail_pii_redactions_total",
			Help: "Total LGPD PII redactions performed before LLM calls.",
		},
		[]string{"category"},
	)

	injectionBlocks := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "pixguard_guardrail_injection_blocks_total",
			Help: "Total prompt injection attempts flagged by the deterministic guardrail.",
		},
	)

	reg.MustRegister(httpRequests, httpDuration, piiRedactions, injectionBlocks)

	return &Metrics{
		Registry:                reg,
		HTTPRequestsTotal:       httpRequests,
		HTTPRequestDuration:     httpDuration,
		GuardrailPIIRedactions:  piiRedactions,
		GuardrailInjectionBlock: injectionBlocks,
	}
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.Registry, promhttp.HandlerOpts{})
}

// InitTracer configures OTLP HTTP tracing towards Langfuse v3 when credentials are present,
// or falls back to a safe local tracer provider when running offline.
func InitTracer(ctx context.Context, cfg config.Config) (func(context.Context) error, error) {
	if cfg.OTelEndpoint == "" || !cfg.LangfuseConfigured() {
		otel.SetTracerProvider(noop.NewTracerProvider())
		return func(context.Context) error { return nil }, nil
	}

	endpoint := strings.TrimRight(cfg.OTelEndpoint, "/")
	if !strings.HasSuffix(endpoint, "/v1/traces") {
		endpoint += "/v1/traces"
	}

	creds := base64.StdEncoding.EncodeToString([]byte(cfg.LangfusePublicKey + ":" + cfg.LangfuseSecretKey))
	exporter, err := otlptracehttp.New(
		ctx,
		otlptracehttp.WithEndpointURL(endpoint),
		otlptracehttp.WithHeaders(map[string]string{
			"Authorization":                fmt.Sprintf("Basic %s", creds),
			"x-langfuse-ingestion-version": "4",
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("create otlp trace exporter: %w", err)
	}

	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceName("pixguard-ai-api"),
			attribute.String("deployment.environment.name", cfg.AppEnv),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create otel resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	return tp.Shutdown, nil
}
