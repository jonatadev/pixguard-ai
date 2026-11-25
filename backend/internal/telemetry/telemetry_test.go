package telemetry

import (
	"context"
	"encoding/base64"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/jonatadev/pixguard-ai/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestNewLoggerLevels(t *testing.T) {
	devLogger := NewLogger("development")
	if !devLogger.Enabled(context.Background(), slog.LevelDebug) {
		t.Fatalf("expected debug logging enabled in development")
	}

	prodLogger := NewLogger("production")
	if prodLogger.Enabled(context.Background(), slog.LevelDebug) {
		t.Fatalf("expected debug logging disabled in production")
	}
	if !prodLogger.Enabled(context.Background(), slog.LevelInfo) {
		t.Fatalf("expected info logging enabled in production")
	}
}

func TestNewMetricsIsolatedRegistryAndHandler(t *testing.T) {
	m1 := NewMetrics()
	m2 := NewMetrics()
	if m1.Registry == m2.Registry {
		t.Fatalf("expected isolated Prometheus registries across instances")
	}

	m1.HTTPRequestsTotal.WithLabelValues("GET", "/healthz", "200").Inc()
	m1.HTTPRequestDuration.WithLabelValues("GET", "/healthz").Observe(0.015)
	m1.GuardrailPIIRedactions.WithLabelValues("CPF").Add(2)
	m1.GuardrailInjectionBlock.Inc()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	m1.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 from metrics handler, got %d", rec.Code)
	}
	body := rec.Body.String()
	expectedMetrics := []string{
		"pixguard_http_requests_total",
		"pixguard_http_request_duration_seconds",
		"pixguard_guardrail_pii_redactions_total",
		"pixguard_guardrail_injection_blocks_total",
	}
	for _, metricName := range expectedMetrics {
		if !strings.Contains(body, metricName) {
			t.Fatalf("expected metric %s in /metrics output", metricName)
		}
	}
}

func TestInitTracerNoopWhenUnconfigured(t *testing.T) {
	defer otel.SetTracerProvider(noop.NewTracerProvider())

	cases := []struct {
		name string
		cfg  config.Config
	}{
		{
			name: "missing endpoint",
			cfg: config.Config{
				LangfusePublicKey: "pk-test",
				LangfuseSecretKey: "sk-test",
			},
		},
		{
			name: "missing langfuse keys",
			cfg: config.Config{
				OTelEndpoint: "http://localhost:3002/api/public/otel",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			shutdown, err := InitTracer(context.Background(), tc.cfg)
			if err != nil {
				t.Fatalf("expected nil error for noop fallback, got %v", err)
			}
			if err := shutdown(context.Background()); err != nil {
				t.Fatalf("expected nil error on noop shutdown, got %v", err)
			}
		})
	}
}

func TestInitTracerOTLPExportsWithBasicAuthAndHeaders(t *testing.T) {
	defer otel.SetTracerProvider(noop.NewTracerProvider())

	var mu sync.Mutex
	var gotPath string
	var gotAuth string
	var gotIngestionVersion string
	var gotBodyLen int

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, _ := io.ReadAll(r.Body)
		mu.Lock()
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotIngestionVersion = r.Header.Get("x-langfuse-ingestion-version")
		gotBodyLen = len(payload)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := config.Config{
		AppEnv:            "test",
		OTelEndpoint:      srv.URL + "/api/public/otel/",
		LangfusePublicKey: "pk-lf-unit",
		LangfuseSecretKey: "sk-lf-unit",
	}

	shutdown, err := InitTracer(context.Background(), cfg)
	if err != nil {
		t.Fatalf("InitTracer failed: %v", err)
	}

	_, span := otel.Tracer("pixguard-ai/test").Start(context.Background(), "unit_span")
	span.End()

	if err := shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if gotPath != "/api/public/otel/v1/traces" {
		t.Fatalf("expected /api/public/otel/v1/traces, got %q", gotPath)
	}
	expectedAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("pk-lf-unit:sk-lf-unit"))
	if gotAuth != expectedAuth {
		t.Fatalf("expected Authorization %q, got %q", expectedAuth, gotAuth)
	}
	if gotIngestionVersion != "4" {
		t.Fatalf("expected x-langfuse-ingestion-version=4, got %q", gotIngestionVersion)
	}
	if gotBodyLen == 0 {
		t.Fatalf("expected non-empty OTLP protobuf trace payload")
	}
}

func TestInitTracerPreservesExistingV1TracesSuffix(t *testing.T) {
	defer otel.SetTracerProvider(noop.NewTracerProvider())

	var mu sync.Mutex
	var gotPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		gotPath = r.URL.Path
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := config.Config{
		AppEnv:            "test",
		OTelEndpoint:      srv.URL + "/api/public/otel/v1/traces",
		LangfusePublicKey: "pk-lf-unit",
		LangfuseSecretKey: "sk-lf-unit",
	}

	shutdown, err := InitTracer(context.Background(), cfg)
	if err != nil {
		t.Fatalf("InitTracer failed: %v", err)
	}

	_, span := otel.Tracer("pixguard-ai/test").Start(context.Background(), "suffix_span")
	span.End()

	if err := shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if gotPath != "/api/public/otel/v1/traces" {
		t.Fatalf("expected suffix not to be duplicated, got %q", gotPath)
	}
}
