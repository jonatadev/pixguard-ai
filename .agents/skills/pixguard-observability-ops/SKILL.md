---
name: pixguard-observability-ops
description: >-
  Use this skill when configuring, troubleshooting, or validating PixGuard AI Docker
  Compose profiles, OpenTelemetry OTLP export to Langfuse v3, Prometheus metrics,
  Grafana dashboards, or structured JSON logs for Loki.
---

# PixGuard AI — Observabilidade & Stack LLMOps

Runbook para operar e validar a stack open-source de observabilidade (Langfuse v3, Prometheus, Grafana, Loki e OpenTelemetry).

## Passos

1. **Validar Sintaxe e Perfis do Docker Compose:**
   - Arquivo alvo: [deploy/docker-compose.yml](../../../deploy/docker-compose.yml)
   - Valide todos os perfis juntos antes de subir ou commitar:
     ```powershell
     docker compose -f deploy/docker-compose.yml --profile monitoring --profile langfuse config --quiet
     ```
   - Perfis disponíveis:
     - Padrão: `postgres` (`:5432`), `api` (`:8080`), `web` (`:4200`)
     - `--profile monitoring`: `prometheus` (`:9090`), `loki` (`:3100`), `grafana` (`:3001`)
     - `--profile langfuse`: `langfuse-web` (`:3000`) + `langfuse-worker`, Postgres, ClickHouse, Redis, MinIO

2. **Verificar Telemetria na API Go:**
   - Confira [backend/internal/telemetry/telemetry.go](../../../backend/internal/telemetry/telemetry.go):
     - O endpoint OTLP garante sufixo `/v1/traces` e envia header `x-langfuse-ingestion-version: 4` com auth `Basic`.
     - Quando `LANGFUSE_PUBLIC_KEY` / `LANGFUSE_SECRET_KEY` estiverem vazios, o sistema deve cair graciosamente para `noop.NewTracerProvider()`.
     - Métricas devem estar registradas em `telemetry.NewMetrics()` usando `prometheus.NewRegistry()` isolado.

3. **Checar Endpoints de Runtime (quando a API estiver rodando):**
   - `GET http://localhost:8080/healthz` → `status: "ok"`
   - `GET http://localhost:8080/readyz` → flags reais de `gemini`, `langfuse`, `database`
   - `GET http://localhost:8080/metrics` → contém `pixguard_http_requests_total`, `pixguard_http_request_duration_seconds`, `pixguard_guardrail_pii_redactions_total`, `pixguard_guardrail_injection_blocks_total`

4. **Provisionamento Grafana & Dashboards:**
   - Ao editar [deploy/grafana/dashboards/pixguard-overview.json](../../../deploy/grafana/dashboards/pixguard-overview.json) ou [deploy/prometheus/prometheus.yml](../../../deploy/prometheus/prometheus.yml), garanta que os nomes das métricas coincidam exatamente com `telemetry.go`.
