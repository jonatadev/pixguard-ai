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

	"github.com/pixguard-ai/pixguard-ai/internal/config"
	"github.com/pixguard-ai/pixguard-ai/internal/httpapi"
	"github.com/pixguard-ai/pixguard-ai/internal/store"
	"github.com/pixguard-ai/pixguard-ai/internal/telemetry"
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

	repo := store.NewMemoryStore()
	apiServer := httpapi.NewServer(cfg, repo, logger, metrics)

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
