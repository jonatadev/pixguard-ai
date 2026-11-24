---
title: Guia Rápido Local
description: Como rodar o backend Go, frontend Angular Material, stack Docker de observabilidade, AI Eval Gate e o portal Astro Starlight.
---

Este guia resume os comandos de desenvolvimento e validação local do **PixGuard AI** (com atenção às particularidades do **Windows PowerShell** e **Linux/macOS**).

## Pré-requisitos

- **Go 1.24+** (`go version`)
- **Node.js 22+** (`node -v`)
- **Docker & Docker Compose v2** (para PostgreSQL + `pgvector`, monitoramento e Langfuse v3)

> **Importante no Windows PowerShell:** Caso a `ExecutionPolicy` bloqueie scripts `.ps1`, use sempre **`npm.cmd`** e **`ng.cmd`** no lugar de `npm` e `ng`.

---

## 1. Backend Go + AI Eval Gate

```powershell
cd backend
go vet ./...
go test -v -cover ./...
go run ./cmd/eval
go run ./cmd/api
```

Endpoints locais disponíveis em `http://localhost:8080`:
- `GET /healthz` — liveness probe
- `GET /readyz` — status das integrações (`gemini`, `langfuse`, `database`)
- `GET /metrics` — métricas Prometheus
- `GET /api/v1/cases` e `GET /api/v1/cases/{id}` — fila e detalhe sanitizado de contestações PIX
- `POST /api/v1/guardrail/preview` — simulador de redação LGPD + Prompt Injection

*(No CI Linux, execute também `go test -race -cover ./...`.)*

---

## 2. Frontend Angular 19 + Angular Material

```powershell
cd frontend
npm.cmd install
npm.cmd run start
```

- Mesa de análise disponível em `http://localhost:4200` (com proxy em `frontend/proxy.conf.json` redirecionando `/api`, `/healthz`, `/readyz` e `/metrics` para `:8080`).
- Build de produção: `npm.cmd run build`.

---

## 3. Stack Docker Compose (Core, Monitoring & Langfuse v3)

```powershell
# Apenas Core (PostgreSQL + pgvector :5432, API Go :8080, Frontend Nginx :4200)
docker compose -f deploy/docker-compose.yml up -d --build

# Core + Monitoramento (Prometheus :9090, Grafana :3001, Loki :3100)
docker compose -f deploy/docker-compose.yml --profile monitoring up -d --build

# Stack Completa (+ Langfuse v3 Self-Hosted :3000)
docker compose -f deploy/docker-compose.yml --profile monitoring --profile langfuse up -d --build
```

---

## 4. Portal de Documentação (Astro Starlight)

```powershell
cd docs-site
npm.cmd install
npm.cmd run dev
```

Os scripts `predev` e `prebuild` executam `node scripts/sync-docs.mjs` automaticamente para sincronizar `specs/`, `docs/regulatory/`, `AGENTS.md`, `.agents/skills/*/SKILL.md`, `CONTRIBUTING.md`, `CHANGELOG.md` e gerar `public/llms.txt`.
