# PixGuard AI — Eval & Red-Team Harness (`evals/`)

Esta pasta centraliza os datasets versionados e harnesses de avaliação usados no **AI Eval Gate** e em experimentos no **Langfuse**.

## Estrutura

- [`datasets/pixguard-golden-v1.json`](./datasets/pixguard-golden-v1.json) — Dataset ouro determinístico (`14` casos cobrindo LGPD, Red-Team OWASP LLM01 e janela legal de 80 dias do MED).
- [`promptfoo/promptfooconfig.yaml`](./promptfoo/promptfooconfig.yaml) — Bateria HTTP open-source com **Promptfoo** contra `POST /api/v1/guardrail/preview`.

## 1. Rodar o Go Eval Harness Offline (portão do CI)

```powershell
cd backend
go run ./cmd/eval
```

Gerando relatórios para artefatos/PR:

```powershell
cd backend
go run ./cmd/eval -markdown-out ../eval-report.md -json-out ../eval-report.json
```

## 2. Sincronizar Dataset & Experiment Run no Langfuse

Com o Langfuse v3 rodando (ex.: `http://localhost:3002`), defina apenas na sessão do terminal:

```powershell
$env:LANGFUSE_BASE_URL = "http://localhost:3002"
$env:LANGFUSE_PUBLIC_KEY = "pk-lf-..."
$env:LANGFUSE_SECRET_KEY = "sk-lf-..."
cd backend
go run ./cmd/eval -langfuse-sync -run-name local-harness-v1
```

Isso registra o dataset `pixguard-golden-v1`, os `dataset-items`, os traces sanitizados (zero PII bruto), os `dataset-run-items` e os Scores (`eval_passed`, `pii_zero_leakage`, `injection_accuracy`) no Langfuse.

## 3. Rodar o Promptfoo contra a API local

Com a API Go rodando (ex.: `http://localhost:8081`):

```powershell
$env:PIXGUARD_API_URL = "http://localhost:8081"
npx.cmd promptfoo@latest eval -c evals/promptfoo/promptfooconfig.yaml --no-cache
```
