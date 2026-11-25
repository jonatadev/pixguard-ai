package httpapi

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jonatadev/pixguard-ai/internal/agents"
	"github.com/jonatadev/pixguard-ai/internal/config"
	"github.com/jonatadev/pixguard-ai/internal/domain"
	"github.com/jonatadev/pixguard-ai/internal/guardrail"
	"github.com/jonatadev/pixguard-ai/internal/store"
	"github.com/jonatadev/pixguard-ai/internal/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const medMaxWindowDays = 80

type Server struct {
	cfg         config.Config
	repo        store.CaseRepository
	logger      *slog.Logger
	metrics     *telemetry.Metrics
	orch        *agents.Orchestrator
	scoreClient telemetry.ScoreClient
	tracer      trace.Tracer
	mux         *http.ServeMux
}

func NewServer(cfg config.Config, repo store.CaseRepository, logger *slog.Logger, metrics *telemetry.Metrics) *Server {
	orch := agents.NewOrchestrator(nil, nil, cfg.GeminiChatModel, cfg.GeminiFastModel, metrics)
	return NewServerWithAgents(cfg, repo, logger, metrics, orch, telemetry.NewScoreClient(cfg))
}

func NewServerWithAgents(
	cfg config.Config,
	repo store.CaseRepository,
	logger *slog.Logger,
	metrics *telemetry.Metrics,
	orch *agents.Orchestrator,
	scoreClient telemetry.ScoreClient,
) *Server {
	if orch == nil {
		orch = agents.NewOrchestrator(nil, nil, cfg.GeminiChatModel, cfg.GeminiFastModel, metrics)
	}
	s := &Server{
		cfg:         cfg,
		repo:        repo,
		logger:      logger,
		metrics:     metrics,
		orch:        orch,
		scoreClient: scoreClient,
		tracer:      otel.Tracer("pixguard-ai/httpapi"),
		mux:         http.NewServeMux(),
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
	s.mux.HandleFunc("POST /api/v1/cases/{id}/analyze", s.handleAnalyzeCase)
	s.mux.HandleFunc("POST /api/v1/cases/{id}/feedback", s.handleSubmitFeedback)
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

func (s *Server) handleAnalyzeCase(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	c, ok := s.repo.GetByID(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "contestacao PIX nao encontrada",
		})
		return
	}

	if c.Status == domain.StatusPendente {
		s.repo.UpdateStatus(c.ID, domain.StatusEmAnalise)
		c.Status = domain.StatusEmAnalise
	}

	if strings.Contains(r.Header.Get("Accept"), "text/event-stream") {
		s.streamAnalysisSSE(w, r, c)
		return
	}

	res := s.orch.Analyze(r.Context(), c, nil)
	writeJSON(w, http.StatusOK, res)
}

func (s *Server) streamAnalysisSSE(w http.ResponseWriter, r *http.Request, c domain.PixCase) {
	flusher, hasFlusher := w.(http.Flusher)
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	writeSSE := func(eventName string, payload any) {
		raw, err := json.Marshal(payload)
		if err != nil {
			return
		}
		_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventName, raw)
		if hasFlusher {
			flusher.Flush()
		}
	}

	res := s.orch.Analyze(r.Context(), c, func(ev agents.AnalysisEvent) {
		writeSSE("stage", ev)
	})
	writeSSE("complete", res)
}

func (s *Server) handleSubmitFeedback(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	c, ok := s.repo.GetByID(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "contestacao PIX nao encontrada",
		})
		return
	}

	var req domain.AnalystFeedbackRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "JSON invalido no corpo da requisicao",
		})
		return
	}

	var nextStatus domain.CaseStatus
	switch req.Decision {
	case domain.DecisionAprovado, domain.DecisionEditado:
		nextStatus = domain.StatusAprovado
	case domain.DecisionRejeitado:
		nextStatus = domain.StatusRejeitado
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "decision deve ser APROVADO, EDITADO ou REJEITADO",
		})
		return
	}

	s.repo.UpdateStatus(c.ID, nextStatus)
	if s.metrics != nil && s.metrics.AnalystFeedbackTotal != nil {
		s.metrics.AnalystFeedbackTotal.WithLabelValues(string(req.Decision)).Inc()
	}

	synced, err := s.scoreClient.PostAnalystDecision(r.Context(), req.TraceID, req.Decision, req.Notes)
	if err != nil {
		s.logger.WarnContext(r.Context(), "langfuse_score_sync_failed",
			slog.String("case_id", c.ID),
			slog.String("error", err.Error()),
		)
	}

	rec := domain.AnalystFeedbackRecord{
		ID:             fmt.Sprintf("FB-%d", time.Now().UTC().UnixNano()),
		CaseID:         c.ID,
		TraceID:        req.TraceID,
		Decision:       req.Decision,
		Notes:          strings.TrimSpace(req.Notes),
		UpdatedStatus:  nextStatus,
		LangfuseSynced: synced,
		CreatedAt:      time.Now().UTC(),
	}
	s.repo.SaveFeedback(rec)
	writeJSON(w, http.StatusOK, rec)
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

func (sr *statusRecorder) Flush() {
	if flusher, ok := sr.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
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
