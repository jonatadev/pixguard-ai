---
name: pixguard-prompt-eval-ops
description: >-
  Use this skill when updating versioned agent prompts in prompts/*.yaml, changing
  regulatory markdown docs in docs/regulatory/, running the offline AI Eval Gate,
  or converting production feedback into regression eval cases.
---

# PixGuard AI — Prompt Versioning & AI Eval Ops

Runbook para manter governança sobre prompts YAML, corpus regulatório e portão de avaliação automatizado (*AI Eval Gate*).

## Passos

1. **Verificar Contratos Antes de Editar Prompts:**
   - Leia [specs/002-multi-agent-orchestration.md](../../../specs/002-multi-agent-orchestration.md) e [specs/004-llmops-eval-and-observability.md](../../../specs/004-llmops-eval-and-observability.md).
   - Inspecione os arquivos atuais em `prompts/`:
     - `fraud-investigator.v1.yaml`
     - `regulatory-rag.v1.yaml`
     - `adjudicator.v1.yaml`
     - `compliance-critic.v1.yaml`

2. **Versionar Mudanças de Prompt:**
   - Para ajustes compatíveis sem quebra de schema, atualize `version` semântica dentro do YAML (ex.: `1.0.1`).
   - Para mudanças no `output_contract` ou comportamento estrutural de decisão, crie um novo arquivo `prompts/<agent>.v2.yaml` preservando o `v1` durante a transição e atualize a spec 002.
   - Mantenha instruções explícitas contra *Prompt Injection* e vazamento de PII em `system_prompt`.

3. **Atualizar Corpus Regulatório (`docs/regulatory/`):**
   - Preserve cabeçalhos claros (`##`, `###`) e tabelas inteiras (para chunking estrutural da Fase 2).
   - Garanta que artigos, prazos (`80 dias`, `72 horas`, `7 dias`, `96 horas`) e regras de exclusão (desacordo comercial) permaneçam consistentes entre os arquivos `.md` e os prompts.

4. **Executar o AI Eval Gate & Converter Divergências:**
   - Sempre que houver feedback humano divergente (👍/👎) ou novo vetor de falha identificado no Langfuse, adicione um novo caso reproduzível em [backend/cmd/eval/main.go](../../../backend/cmd/eval/main.go).
   - Rode na pasta `backend/`:
     ```powershell
     go test -v -cover ./...
     go run ./cmd/eval
     ```
   - Exija `N/N passed` com saída `0` antes de qualquer commit.
