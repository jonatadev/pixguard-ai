---
name: pixguard-release-conventional
description: >-
  Use this skill when staging commits, writing Conventional Commit messages, bumping
  SemVer tags, updating CHANGELOG.md, or preparing releases and Pull Requests for
  github.com/jonatadev/pixguard-ai.
---

# PixGuard AI — Versionamento Profissional & Conventional Commits

Runbook para manter histórico Git limpo, auditável e alinhado ao padrão Conventional Commits + SemVer em `https://github.com/jonatadev/pixguard-ai`.

## Passos

1. **Validar Escopo e Limpeza Antes de Commitar:**
   - Verifique `git status` e `git diff` para garantir que nenhum `.env`, token, chave de API ou binário compilado (`backend/bin/`, `frontend/dist/`, `node_modules/`) esteja sendo versionado.
   - Rode os portões aplicáveis (`go vet ./...`, `go test -cover ./...`, `go run ./cmd/eval`, `npm.cmd run build`).

2. **Formato Obrigatório de Commit (Conventional Commits):**
   ```text
   <type>(<scope>): <descrição imperativa curta em minúsculo>
   ```
   - **Tipos permitidos:** `feat`, `fix`, `docs`, `refactor`, `test`, `chore`, `ci`, `perf`, `build`
   - **Escopos recomendados:** `repo`, `specs`, `agents`, `backend`, `guardrail`, `frontend`, `prompts`, `observability`, `ci`, `docs`
   - Quebre alterações grandes em **commits atômicos** por domínio lógico (ex.: `specs/` separado de `AGENTS.md`/`.agents/` e de código Go/Angular).

3. **Atualizar `CHANGELOG.md` e Criar Tag SemVer (em releases):**
   - Siga *Keep a Changelog* em [CHANGELOG.md](../../../CHANGELOG.md) agrupando itens sob `Added`, `Changed`, `Fixed`, `Security`.
   - Use SemVer (`v0.1.0`, `v0.2.0`, etc.):
     - `PATCH` (`0.2.x`): correções sem quebra de contrato.
     - `MINOR` (`0.x.0`): novas capacidades, specs, agentes ou endpoints retrocompatíveis.
     - `MAJOR` (`x.0.0`): mudanças incompatíveis de API/schema/prompt contract.
   - Crie sempre tag anotada:
     ```powershell
     git tag -a vX.Y.Z -m "Release vX.Y.Z: resumo da entrega"
     ```

4. **Checklist de PR / Push:**
   - Preencha os itens de [.github/PULL_REQUEST_TEMPLATE.md](../../../.github/PULL_REQUEST_TEMPLATE.md).
   - Ao publicar para o remoto: `git push -u origin main --tags`.
