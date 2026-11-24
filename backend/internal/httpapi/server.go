package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/pixguard-ai/pixguard-ai/internal/config"
	"github.com/pixguard-ai/pixguard-ai/internal/domain"
	"github.com/pixguard-ai/pixguard-ai/internal/guardrail"
	"github.com/pixguard-ai/pixguard-ai/internal/store"
	"github.com/pixguard-ai/pixguard-ai/internal/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const medMaxWindowDays = 80

type Server struct {
	cfg     config.Config
	repo    store.CaseRepository
	logger  *slog.Logger
	metrics *telemetry.Metrics
	tracer  trace.Tracer
	mux     *http.ServeMux
}

func NewServer(cfg config.Config, repo store.CaseRepository, logger *slog.Logger, metrics *telemetry.Metrics) *Server {
	s := &Server{
		cfg:     cfg,
		repo:    repo,
		logger:  logger,
		metrics: metrics,
		tracer:  otel.Tracer("pixguard-ai/httpapi"),
		mux:     http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

func (s *Server) Handler() http.Handler {
	return s.corsMiddleware(s.observabilityMiddleware(s.mux))
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /healthz", s.handleHealthz)
	s.mux.HandleFunc("GET /readyz", s.handleReadyz)
	s.mux.Handle("GET /metrics", s.metrics.Handler())
	s.mux.HandleFunc("GET /api/v1/cases", s.handleListCases)
	s.mux.HandleFunc("GET /api/v1/cases/{id}", s.handleGetCase)
	s.mux.HandleFunc("POST /api/v1/guardrail/preview", s.handleGuardrailPreview)
}

func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, domain.HealthResponse{
		Status:  "ok",
		Service: "pixguard-ai-api",
		Env:     s.cfg.AppEnv,
	})
}

func (s *Server) handleReadyz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, domain.ReadyResponse{
		Status: "ready",
		Integrations: map[string]bool{
			"gemini":   s.cfg.GeminiConfigured(),
			"langfuse": s.cfg.LangfuseConfigured(),
			"database": s.cfg.DatabaseConfigured(),
		},
		Models: map[string]string{
			"chat":      s.cfg.GeminiChatModel,
			"fast":      s.cfg.GeminiFastModel,
			"embedding": s.cfg.GeminiEmbeddingModel,
		},
	})
}

func (s *Server) handleListCases(w http.ResponseWriter, r *http.Request) {
	risk := r.URL.Query().Get("risk")
	status := r.URL.Query().Get("status")
	cases := s.repo.List(risk, status)
	writeJSON(w, http.StatusOK, map[string]any{
		"cases": cases,
		"total": len(cases),
	})
}

func (s *Server) handleGetCase(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	c, ok := s.repo.GetByID(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "contestacao PIX nao encontrada",
		})
		return
	}

	sanitized, report := guardrail.Sanitize(c.RawDescription)
	s.recordGuardrailMetrics(report)

	writeJSON(w, http.StatusOK, domain.CaseDetailResponse{
		Case:                 c,
		SanitizedDescription: sanitized,
		GuardrailReport:      report,
		MedWindowValid:       c.DaysSinceTx <= medMaxWindowDays,
	})
}

func (s *Server) handleGuardrailPreview(w http.ResponseWriter, r *http.Request) {
	var req domain.GuardrailPreviewRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "JSON invalido no corpo da requisicao",
		})
		return
	}

	sanitized, report := guardrail.Sanitize(req.Text)
	s.recordGuardrailMetrics(report)

	writeJSON(w, http.StatusOK, domain.GuardrailPreviewResponse{
		OriginalLength: len(req.Text),
		SanitizedText:  sanitized,
		Report:         report,
	})
}

func (s *Server) recordGuardrailMetrics(report domain.GuardrailReport) {
	for category, count := range report.PIICategories {
		if count > 0 {
			s.metrics.GuardrailPIIRedactions.WithLabelValues(category).Add(float64(count))
		}
	}
	if report.PromptInjectionDetected {
		s.metrics.GuardrailInjectionBlock.Inc()
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}

func (s *Server) observabilityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx, span := s.tracer.Start(
			r.Context(),
			r.Method+" "+r.URL.Path,
			trace.WithAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.path", r.URL.Path),
			),
		)
		defer span.End()

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r.WithContext(ctx))

		duration := time.Since(start)
		routePattern := r.Pattern
		if routePattern == "" {
			routePattern = r.URL.Path
		}

		s.metrics.HTTPRequestsTotal.WithLabelValues(r.Method, routePattern, strconv.Itoa(rec.status)).Inc()
		s.metrics.HTTPRequestDuration.WithLabelValues(r.Method, routePattern).Observe(duration.Seconds())

		traceID := ""
		if span.SpanContext().HasTraceID() {
			traceID = span.SpanContext().TraceID().String()
		}

		s.logger.InfoContext(ctx, "http_request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rec.status),
			slog.Int64("duration_ms", duration.Milliseconds()),
			slog.String("trace_id", traceID),
		)
	})
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
