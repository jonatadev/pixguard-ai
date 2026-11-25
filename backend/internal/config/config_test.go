package config

import "testing"

func clearConfigEnv(t *testing.T) {
	t.Helper()
	keys := []string{
		"PORT",
		"APP_ENV",
		"DATABASE_URL",
		"GEMINI_API_KEY",
		"GEMINI_CHAT_MODEL",
		"GEMINI_FAST_MODEL",
		"GEMINI_EMBEDDING_MODEL",
		"OTEL_EXPORTER_OTLP_ENDPOINT",
		"LANGFUSE_BASE_URL",
		"LANGFUSE_PUBLIC_KEY",
		"LANGFUSE_SECRET_KEY",
	}
	for _, k := range keys {
		t.Setenv(k, "")
	}
}

func TestLoadDefaults(t *testing.T) {
	clearConfigEnv(t)

	cfg := Load()

	if cfg.Port != "8080" {
		t.Fatalf("expected default Port 8080, got %q", cfg.Port)
	}
	if cfg.AppEnv != "development" {
		t.Fatalf("expected default AppEnv development, got %q", cfg.AppEnv)
	}
	if cfg.GeminiChatModel != "gemini-3-flash-preview" {
		t.Fatalf("unexpected default GeminiChatModel: %q", cfg.GeminiChatModel)
	}
	if cfg.GeminiFastModel != "gemini-2.5-flash-lite" {
		t.Fatalf("unexpected default GeminiFastModel: %q", cfg.GeminiFastModel)
	}
	if cfg.GeminiEmbeddingModel != "gemini-embedding-001" {
		t.Fatalf("unexpected default GeminiEmbeddingModel: %q", cfg.GeminiEmbeddingModel)
	}
	if cfg.LangfuseBaseURL != "http://localhost:3000" {
		t.Fatalf("unexpected default LangfuseBaseURL: %q", cfg.LangfuseBaseURL)
	}
	if cfg.GeminiConfigured() || cfg.LangfuseConfigured() || cfg.DatabaseConfigured() {
		t.Fatalf("expected all integration flags false by default, got gemini=%v langfuse=%v db=%v",
			cfg.GeminiConfigured(), cfg.LangfuseConfigured(), cfg.DatabaseConfigured())
	}
}

func TestLoadCustomEnvAndTrimming(t *testing.T) {
	clearConfigEnv(t)

	t.Setenv("PORT", " 8081 ")
	t.Setenv("APP_ENV", " production ")
	t.Setenv("DATABASE_URL", " postgres://pixguard:pixguard@localhost:5432/pixguard ")
	t.Setenv("GEMINI_API_KEY", " test-gemini-key ")
	t.Setenv("GEMINI_CHAT_MODEL", " gemini-3-pro-preview ")
	t.Setenv("GEMINI_FAST_MODEL", " gemini-2.5-flash ")
	t.Setenv("GEMINI_EMBEDDING_MODEL", " gemini-embedding-001 ")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", " http://localhost:3002/api/public/otel ")
	t.Setenv("LANGFUSE_BASE_URL", " http://localhost:3002 ")
	t.Setenv("LANGFUSE_PUBLIC_KEY", " pk-lf-test ")
	t.Setenv("LANGFUSE_SECRET_KEY", " sk-lf-test ")

	cfg := Load()

	if cfg.Port != "8081" || cfg.AppEnv != "production" {
		t.Fatalf("expected trimmed Port/AppEnv, got %q / %q", cfg.Port, cfg.AppEnv)
	}
	if cfg.OTelEndpoint != "http://localhost:3002/api/public/otel" {
		t.Fatalf("unexpected trimmed OTelEndpoint: %q", cfg.OTelEndpoint)
	}
	if cfg.LangfuseBaseURL != "http://localhost:3002" {
		t.Fatalf("unexpected trimmed LangfuseBaseURL: %q", cfg.LangfuseBaseURL)
	}
	if !cfg.GeminiConfigured() || !cfg.LangfuseConfigured() || !cfg.DatabaseConfigured() {
		t.Fatalf("expected all integration flags true, got gemini=%v langfuse=%v db=%v",
			cfg.GeminiConfigured(), cfg.LangfuseConfigured(), cfg.DatabaseConfigured())
	}
}

func TestLangfuseConfiguredRequiresBothKeys(t *testing.T) {
	cases := []struct {
		name     string
		pub      string
		sec      string
		expected bool
	}{
		{name: "both empty", pub: "", sec: "", expected: false},
		{name: "only public", pub: "pk-lf-1", sec: "", expected: false},
		{name: "only secret", pub: "", sec: "sk-lf-1", expected: false},
		{name: "both set", pub: "pk-lf-1", sec: "sk-lf-1", expected: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := Config{LangfusePublicKey: tc.pub, LangfuseSecretKey: tc.sec}
			if got := cfg.LangfuseConfigured(); got != tc.expected {
				t.Fatalf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}
