# Changelog — PixGuard AI

Todas as alterações relevantes deste projeto são documentadas neste arquivo seguindo o padrão [Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/) e versionamento [SemVer](https://semver.org/lang/pt-BR/).

## [Unreleased]

### Fixed
- Corrigido o profile `langfuse` em `deploy/docker-compose.yml` para usar `cgr.dev/chainguard/minio` (substituindo `minio/minio:latest` removida do Docker Hub), healthcheck IPv4 (`127.0.0.1:8123/ping`) no ClickHouse e porta configurável via `LANGFUSE_PORT` / `LANGFUSE_BASE_URL`.

### Added
- **TDD & Cobertura de Testes em Go (`backend/internal/`):**
  - Adicionados testes unitários para `internal/config` (`config_test.go`), `internal/store` (`memory_test.go`, incluindo leituras concorrentes com `-race`) e `internal/telemetry` (`telemetry_test.go` com validação OTLP/Langfuse via `httptest`).
  - Adicionado gate mínimo de cobertura de `>= 85%` em `.github/workflows/ci.yml` para `backend/internal/...`.
- **Eval & Red-Team Harness (`evals/` + `backend/internal/evalharness`):**
  - Dataset versionado `evals/datasets/pixguard-golden-v1.json` com 14 cenários determinísticos (LGPD, Red-Team OWASP LLM01 e janela legal de 80 dias do MED).
  - Pacote Go `internal/evalharness` e flags em `cmd/eval` para gerar relatórios JSON/Markdown e sincronizar Datasets, traces sanitizados, Experiment Runs e Scores no **Langfuse** (`-langfuse-sync`).
  - Configuração **Promptfoo** em `evals/promptfoo/promptfooconfig.yaml` para validação HTTP em `/api/v1/guardrail/preview`.
  - Pipeline **GitHub Actions CI & AI Eval Gate** (`.github/workflows/ci.yml`) com `go vet`, `go test -race`, `govulncheck`, comentário automático do Eval no PR, `ng build`, validação Docker Compose e `Gitleaks`.
- Portal de documentação **Astro Starlight** em `docs-site/` com busca local Pagefind, diagramas Mermaid (`astro-mermaid`), script single-source-of-truth (`sync-docs.mjs`), geração automática de `llms.txt` e workflow GitHub Actions (`.github/workflows/docs.yml`) para PR check e deploy no GitHub Pages.

### Planned
- Motor Multi-Agente em Go (`google.golang.org/genai`, `errgroup`, Tool Calling e streaming SSE).
- RAG híbrido (`pgvector` + `tsvector` + RRF) sobre `docs/regulatory/`.
- Pipelines GitHub Actions (CI, DevSecOps `Gitleaks`/`Trivy`, AI Eval Gate e CD para GHCR).

---

## [0.2.0] - 2026-11-24

### Added
- **Especificações de Engenharia AI-Native (`specs/`):**
  - `specs/README.md` com índice mestre e ciclo de vida Spec-Driven Development.
  - `specs/001-system-architecture.md` com arquitetura, limites de pacotes Go, schema `pgvector`, contratos HTTP e SLI/SLOs.
  - `specs/002-multi-agent-orchestration.md` com DAG controlado, Tool Calling determinístico, RAG híbrido (RRF) e tiering Gemini.
  - `specs/003-lgpd-security-guardrails.md` com mapeamento OWASP LLM Top 10, precedência de redação LGPD e regras Bacen/MED/PLD-FT.
  - `specs/004-llmops-eval-and-observability.md` com critérios do AI Eval Gate, OTLP Langfuse v3, Prometheus/Grafana/Loki e CI/CD.
  - `specs/adr/0001-tech-stack-and-ai-sdlc.md` documentando a decisão da stack Go + Angular Material + PostgreSQL/`pgvector` + observabilidade open-source.
- **Regras Hierárquicas `AGENTS.md`:** instruções de engenharia na raiz (`AGENTS.md`), `backend/AGENTS.md` e `frontend/AGENTS.md`.
- **Antigravity Workspace Skills (`.agents/skills/`):**
  - `pixguard-spec-driven-dev` (inclui `resources/adr-template.md`)
  - `pixguard-lgpd-guardrail`
  - `pixguard-prompt-eval-ops`
  - `pixguard-observability-ops`
  - `pixguard-release-conventional`
- **Governança & Normalização Git:** `CONTRIBUTING.md`, `CHANGELOG.md` e `.gitattributes`.

### Changed
- Alinhado módulo Go (`backend/go.mod` e imports internos) para `github.com/jonatadev/pixguard-ai`.
- Atualizado `.github/CODEOWNERS` para `@jonatadev`.
- Atualizado `README.md` com seção de Engenharia AI-Native, specs, skills e versionamento.

---

## [0.1.0] - 2026-11-24

### Added
- Fundação completa da Fase 1 do **PixGuard AI**:
  - Backend Go 1.24 (`cmd/api`, `cmd/eval`, `internal/config`, `internal/domain`, `internal/guardrail`, `internal/httpapi`, `internal/store`, `internal/telemetry`) e migration PostgreSQL + `pgvector` (`001_init.sql`).
  - Guardrail determinístico LGPD (CPF, CNPJ, e-mail, telefone, chave EVP) e detecção de Prompt Injection com testes unitários (`97,0%` cobertura) e Eval CLI (`5/5`).
  - Frontend Angular 19 + Angular Material 3 (`frontend/`) com mesa de análise PIX/MED e preview de guardrail.
  - Corpus regulatório simulado (`docs/regulatory/`) e prompts YAML v1 (`prompts/`).
  - Stack Docker Compose (`deploy/docker-compose.yml`) com PostgreSQL/`pgvector`, Prometheus, Grafana, Loki e Langfuse v3 self-hosted.

[Unreleased]: https://github.com/jonatadev/pixguard-ai/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/jonatadev/pixguard-ai/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/jonatadev/pixguard-ai/releases/tag/v0.1.0
