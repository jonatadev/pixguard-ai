---
name: pixguard-lgpd-guardrail
description: >-
  Use this skill when adding, modifying, or auditing LGPD PII redaction rules,
  prompt-injection detection patterns, or business safety invariants in
  backend/internal/guardrail and its Eval fixtures.
---

# PixGuard AI — Guardrail LGPD & Segurança Determinística

Runbook para evoluir com segurança a camada determinística de proteção de dados (LGPD) e defesa contra *Prompt Injection* em `backend/internal/guardrail/`.

## Passos

1. **Consultar a Especificação de Segurança:**
   - Leia [specs/003-lgpd-security-guardrails.md](../../../specs/003-lgpd-security-guardrails.md) e [docs/regulatory/03-politica-interna-antifraude-lgpd.md](../../../docs/regulatory/03-politica-interna-antifraude-lgpd.md).

2. **Respeitar a Ordem de Precedência em `lgpd.go`:**
   - Edite [backend/internal/guardrail/lgpd.go](../../../backend/internal/guardrail/lgpd.go) preservando a ordem estrita de redação:
     1. `PIX_EVP` (UUID) → `[PIX_EVP_REDACTED]`
     2. `CPF` formatado → `[CPF_REDACTED]`
     3. `CPF` contextual (11 dígitos) → `${1}[CPF_REDACTED]`
     4. `CNPJ` formatado → `[CNPJ_REDACTED]`
     5. `EMAIL` → `[EMAIL_REDACTED]`
     6. `PHONE` brasileiro → `[PHONE_REDACTED]`
   - Nunca aplique regex de telefone antes de UUID, CPF ou CNPJ para evitar corrupção parcial de tokens.

3. **Adicionar Testes Unitários e Casos Red-Team no Eval:**
   - Adicione casos positivos e negativos (falsos positivos em narrativa limpa) em `backend/internal/guardrail/lgpd_test.go`.
   - Adicione cenários de regressão em [backend/cmd/eval/main.go](../../../backend/cmd/eval/main.go) com `forbiddenRawSubstrings` e `expectInjection`.

4. **Validar Invariante de Zero-PII nos Logs e Métricas:**
   - Verifique em [backend/internal/httpapi/server.go](../../../backend/internal/httpapi/server.go) que apenas contagens agregadas (`PIICategories`) são enviadas ao Prometheus (`pixguard_guardrail_pii_redactions_total` / `pixguard_guardrail_injection_blocks_total`) e nunca o texto bruto.

5. **Executar Verificação:**
   ```powershell
   cd backend
   go vet ./...
   go test -v -cover ./internal/guardrail/... ./internal/httpapi/...
   go run ./cmd/eval
   ```
   - Critério de aprovação: **0% vazamento de PII**, **100% detecção no Eval**, cobertura de `internal/guardrail` **>= 90%**.
