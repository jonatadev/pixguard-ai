---
name: pixguard-spec-driven-dev
description: >-
  Use this skill when proposing, designing, or modifying architecture, API endpoints,
  database schemas, agent workflows, or ADRs in PixGuard AI so specs stay synchronized
  with code, prompts, and frontend/backend contracts.
---

# PixGuard AI — Spec-Driven Development

Este runbook garante que nenhuma feature ou mudança estrutural no PixGuard AI fuja das especificações oficiais em `specs/`.

## Passos Obrigatórios

1. **Ler as Specs Afetadas Antes de Codar:**
   - Índice mestre: [specs/README.md](../../../specs/README.md)
   - Arquitetura & HTTP/DB: [specs/001-system-architecture.md](../../../specs/001-system-architecture.md)
   - Multi-Agente & RAG: [specs/002-multi-agent-orchestration.md](../../../specs/002-multi-agent-orchestration.md)
   - Guardrails LGPD & Negócio: [specs/003-lgpd-security-guardrails.md](../../../specs/003-lgpd-security-guardrails.md)
   - LLMOps, Eval & Observabilidade: [specs/004-llmops-eval-and-observability.md](../../../specs/004-llmops-eval-and-observability.md)

2. **Avaliar Necessidade de Nova ADR:**
   - Se a mudança alterar stack tecnológica, banco de dados, modelo de observabilidade, protocolo de agentes ou limite de segurança, crie `specs/adr/NNNN-titulo-curto.md` usando o template [./resources/adr-template.md](./resources/adr-template.md) e registre na tabela de `specs/README.md`.

3. **Atualizar Especificação e Contratos em Conjunto:**
   - Atualize a spec correspondente com status, diagramas Mermaid e contratos.
   - Verifique paridade entre:
     - `backend/internal/domain/models.go`
     - `frontend/src/app/core/models/pix.models.ts`
     - `backend/migrations/*.sql`
     - `prompts/*.yaml`

4. **Validar Consistência Final:**
   - Confirme que todos os links Markdown relativos dentro de `specs/` apontam para caminhos reais.
   - Rode os portões de verificação definidos no `AGENTS.md` da raiz (`go vet ./...`, `go test -cover ./...`, `go run ./cmd/eval`, `npm.cmd run build`).
