# Backend Go 1.24 — Diretrizes Locais (`backend/AGENTS.md`)

Este arquivo complementa o [`AGENTS.md`](../AGENTS.md) da raiz para todo desenvolvimento em Go dentro de `backend/`.

---

## 1. Convenções de Módulo e Arquitetura de Pacotes

- **Módulo Go:** `github.com/jonatadev/pixguard-ai` (definido em `go.mod`). Nunca reintroduza paths antigos em imports.
- **Separação de responsabilidades:**
  - `cmd/api`: apenas wiring de configuração, telemetria, store e ciclo de vida HTTP (`Shutdown` gracioso).
  - `cmd/eval`: runner offline determinístico de segurança e qualidade de IA; deve sair com código `!= 0` em qualquer regressão.
  - `internal/domain`: structs puras e constantes de domínio sem dependência de HTTP, SQL ou SDK de LLM.
  - `internal/guardrail`: lógica determinística pura de redação LGPD e detecção de Prompt Injection.
  - `internal/httpapi`: parsing/validação HTTP, handlers e middlewares de observabilidade; não implemente regras regex de PII aqui.
  - `internal/store`: interface `CaseRepository` e adaptadores (memória e futuro PostgreSQL/`pgvector`).
  - `internal/telemetry`: `slog` JSON, Prometheus Registry isolado e inicialização OTLP.
  - *Fase 2 (`internal/agents`, `internal/rag`):* manter prompts fora do código Go (carregar de `prompts/*.yaml`) e isolar chamadas ao SDK oficial `google.golang.org/genai`.

---

## 2. Regras Críticas de Guardrail, Concorrência e Telemetria

1. **Zero-PII em Logs e Traces:** Nunca passe `c.RawDescription` ou `req.Text` bruto para `s.logger`, atributos de span OpenTelemetry ou mensagens de erro HTTP.
2. **Métricas Isoladas em Testes:** Sempre instancie métricas via `telemetry.NewMetrics()` (que cria um `prometheus.NewRegistry()` dedicado) em vez de registrar no `prometheus.DefaultRegisterer`, evitando panics entre testes.
3. **Timeouts e Contexto:** Propague `context.Context` em toda operação de I/O, banco e LLM. Ao implementar agentes paralelos na Fase 2, use `golang.org/x/sync/errgroup` com `WithContext` e timeout explícito por etapa.
4. **Cobertura Mínima:** Mantenha cobertura `>= 90%` em `internal/guardrail` e `internal/httpapi`.
5. **Validação de Payload:** Preserve `http.MaxBytesReader` nos endpoints `POST` para evitar abuso de memória.

---

## 3. Verificação em `backend/`

```powershell
cd backend
go vet ./...
go test -v -cover ./...
go run ./cmd/eval
go build -o bin/pixguard-api.exe ./cmd/api
```
*(Lembrete: `bin/` é ignorado pelo `.gitignore`; remova binários temporários se não precisar mantê-los localmente).*
