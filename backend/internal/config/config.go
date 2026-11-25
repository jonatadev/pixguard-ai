package config

import (
	"os"
	"strings"
)

// Config holds runtime settings for the PixGuard AI API.
type Config struct {
	Port                string
	AppEnv              string
	DatabaseURL         string
	GeminiAPIKey        string
	GeminiChatModel     string
	GeminiFastModel     string
	GeminiEmbeddingModel string
	OTelEndpoint        string
	LangfuseBaseURL     string
	LangfusePublicKey   string
	LangfuseSecretKey   string
}

// Load reads configuration from environment variables with safe development defaults.
func Load() Config {
	return Config{
		Port:                 getEnv("PORT", "8080"),
		AppEnv:               getEnv("APP_ENV", "development"),
		DatabaseURL:          strings.TrimSpace(os.Getenv("DATABASE_URL")),
		GeminiAPIKey:         strings.TrimSpace(os.Getenv("GEMINI_API_KEY")),
		GeminiChatModel:      getEnv("GEMINI_CHAT_MODEL", "gemini-3.8-flash"),
		GeminiFastModel:      getEnv("GEMINI_FAST_MODEL", "gemini-3.5-flash-lite"),
		GeminiEmbeddingModel: getEnv("GEMINI_EMBEDDING_MODEL", "gemini-embedding-001"),
		OTelEndpoint:         strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")),
		LangfuseBaseURL:      getEnv("LANGFUSE_BASE_URL", "http://localhost:3000"),
		LangfusePublicKey:    strings.TrimSpace(os.Getenv("LANGFUSE_PUBLIC_KEY")),
		LangfuseSecretKey:    strings.TrimSpace(os.Getenv("LANGFUSE_SECRET_KEY")),
	}
}

func (c Config) GeminiConfigured() bool {
	return c.GeminiAPIKey != ""
}

func (c Config) LangfuseConfigured() bool {
	return c.LangfusePublicKey != "" && c.LangfuseSecretKey != ""
}

func (c Config) DatabaseConfigured() bool {
	return c.DatabaseURL != ""
}

func getEnv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
