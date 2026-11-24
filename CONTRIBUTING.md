# Guia de Contribuição — PixGuard AI

Obrigado por contribuir com o **PixGuard AI** (`https://github.com/jonatadev/pixguard-ai`). Este projeto segue práticas de **Engenharia AI-Native (Spec-First + Eval-First + Privacy by Design)**.

---

## 1. Fluxo de Trabalho (Spec-First)

1. **Leia as regras do contexto:**
   - Regras globais: [`AGENTS.md`](./AGENTS.md)
   - Regras por camada: [`backend/AGENTS.md`](./backend/AGENTS.md) e [`frontend/AGENTS.md`](./frontend/AGENTS.md)
   - Workspace Skills do Antigravity: [`.agents/skills/`](./.agents/skills/)
2. **Comece pela especificação (`specs/`):**
   - Toda nova feature, endpoint, regra de guardrail ou alteração de prompt deve atualizar primeiro a spec correspondente em [`specs/README.md`](./specs/README.md) (ou criar uma ADR usando [`.agents/skills/pixguard-spec-driven-dev/resources/adr-template.md`](./.agents/skills/pixguard-spec-driven-dev/resources/adr-template.md)).
3. **Mantenha contratos sincronizados:**
   - Backend Go (`backend/internal/domain/models.go`), Frontend TypeScript (`frontend/src/app/core/models/pix.models.ts`), SQL (`backend/migrations/*.sql`) e Prompts (`prompts/*.yaml`).
4. **Valide antes de abrir PR:**
   - Execute todos os portões locais descritos na seção 4.

---

## 2. Convenção de Branches & Commits

### Branches
- `feat/<escopo-curto>` — novas funcionalidades ou agentes
- `fix/<escopo-curto>` — correções de bug ou guardrail
- `docs/<escopo-curto>` — especificações, ADRs e documentação
- `refactor/<escopo-curto>` — refatorações sem mudança funcional

### Conventional Commits (Obrigatório)

```text
<type>(<scope>): <descrição imperativa curta em minúsculo>
```

| Type | Quando usar | Exemplo no PixGuard AI |
| :--- | :--- | :--- |
| `feat` | Nova funcionalidade, agente, endpoint ou skill | `feat(agents): add hierarchical AGENTS.md rules and Antigravity workspace skills` |
| `fix` | Correção de bug, regex LGPD ou falha de UI | `fix(guardrail): prevent partial phone masking inside CNPJ` |
| `docs` | Specs, ADRs, README, CHANGELOG, guias | `docs(specs): add AI-native architecture, multi-agent, LGPD, and LLMOps specs` |
| `refactor` | Melhoria interna sem alterar comportamento | `refactor(repo): align Go module path, CODEOWNERS, and line endings with github.com/jonatadev` |
| `test` | Novos testes unitários ou casos de Eval | `test(eval): add red-team prompt injection fixture for MED override` |
| `ci` | Workflows GitHub Actions e DevSecOps | `ci(actions): add AI Eval Gate and Trivy image scan` |
| `chore` | Manutenção de dependências, build ou tooling | `chore(deps): update google.golang.org/genai sdk` |

Escopos recomendados: `repo`, `specs`, `agents`, `backend`, `guardrail`, `frontend`, `prompts`, `observability`, `ci`, `docs`.

---

## 3. Segurança & Privacidade (LGPD)

- **Nunca** commite `.env`, tokens de API (`GEMINI_API_KEY`, `LANGFUSE_SECRET_KEY`) ou CPFs/CNPJs/e-mails reais de clientes. Use apenas dados fictícios sintéticos em testes e fixtures.
- Toda narrativa enviada a LLMs deve passar antes por `guardrail.Sanitize`.

---

## 4. Portões de Verificação Local

### Windows PowerShell
```powershell
# 1. Backend Go + AI Eval Gate
cd backend
go vet ./...
go test -v -cover ./...
go run ./cmd/eval

# 2. Frontend Angular Material
cd ..\frontend
npm.cmd run build

# 3. Docker Compose (validação de sintaxe dos 3 perfis)
cd ..
docker compose -f deploy/docker-compose.yml --profile monitoring --profile langfuse config --quiet
```

### Linux / macOS / CI
```bash
cd backend && go vet ./... && go test -race -cover ./... && go run ./cmd/eval
cd ../frontend && npm run build
cd .. && docker compose -f deploy/docker-compose.yml --profile monitoring --profile langfuse config --quiet
```

---

## 5. Checklist de Pull Request

Todo PR deve seguir o template em [`.github/PULL_REQUEST_TEMPLATE.md`](./.github/PULL_REQUEST_TEMPLATE.md) e receber revisão dos responsáveis configurados em [`.github/CODEOWNERS`](./.github/CODEOWNERS).
