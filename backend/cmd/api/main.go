package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jonatadev/pixguard-ai/internal/agents"
	"github.com/jonatadev/pixguard-ai/internal/config"
	"github.com/jonatadev/pixguard-ai/internal/httpapi"
	"github.com/jonatadev/pixguard-ai/internal/rag"
	"github.com/jonatadev/pixguard-ai/internal/store"
	"github.com/jonatadev/pixguard-ai/internal/telemetry"
)

func main() {
	cfg := config.Load()
	logger := telemetry.NewLogger(cfg.AppEnv)
	metrics := telemetry.NewMetrics()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	shutdownTracer, err := telemetry.InitTracer(ctx, cfg)
	if err != nil {
		logger.Warn("otel_tracer_init_failed", slog.String("error", err.Error()))
	} else {
		defer func() {
			shutdownCtx, c := context.WithTimeout(context.Background(), 5*time.Second)
			defer c()
			_ = shutdownTracer(shutdownCtx)
		}()
	}

	ragIndex := loadRegulatoryIndex(ctx, logger)
	var gen agents.JSONGenerator
	if cfg.GeminiConfigured() {
		g, err := agents.NewGeminiGenerator(ctx, cfg.GeminiAPIKey)
		if err != nil {
			logger.Warn("gemini_client_init_failed", slog.String("error", err.Error()))
		} else {
			gen = g
		}
	}

	orch := agents.NewOrchestrator(ragIndex, gen, cfg.GeminiChatModel, cfg.GeminiFastModel, metrics)
	scoreClient := telemetry.NewScoreClient(cfg)

	repo := store.NewMemoryStore()
	apiServer := httpapi.NewServerWithAgents(cfg, repo, logger, metrics, orch, scoreClient)

	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           apiServer.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("pixguard_api_starting",
			slog.String("port", cfg.Port),
			slog.String("env", cfg.AppEnv),
			slog.Bool("langfuse_configured", cfg.LangfuseConfigured()),
			slog.Bool("gemini_configured", cfg.GeminiConfigured()),
		)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("pixguard_api_error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("pixguard_api_shutting_down")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("pixguard_api_shutdown_error", slog.String("error", err.Error()))
	}
}

func loadRegulatoryIndex(ctx context.Context, logger *slog.Logger) *rag.Index {
	for _, dir := range []string{"../docs/regulatory", "./docs/regulatory"} {
		chunks, err := rag.LoadCorpusFromDir(dir)
		if err != nil {
			continue
		}
		idx, err := rag.NewIndex(ctx, chunks, rag.DeterministicEmbedder{Dims: 256})
		if err == nil {
			logger.Info("regulatory_rag_loaded", slog.String("dir", dir), slog.Int("chunks", len(chunks)))
			return idx
		}
	}
	logger.Warn("regulatory_rag_not_loaded")
	return nil
}
