## Resumo da Mudança
Descreva o objetivo técnico e regulatório deste Pull Request.

## Checklist de SDLC & AI Governance
- [ ] **Backend Go:** `go test ./...` e `go vet ./...` executados com sucesso.
- [ ] **Frontend Angular:** `npm.cmd run build` executado sem erros.
- [ ] **Guardrail LGPD:** Nenhum dado pessoal (CPF, CNPJ, e-mail, telefone, chave EVP) vaza sem redação.
- [ ] **Prompts Versionados:** Alterações em `prompts/*.yaml` incrementam a versão e atualizam o dataset de avaliação.
- [ ] **AI Eval Gate:** `go run ./cmd/eval` aprovado sem regressões de segurança ou precisão.
- [ ] **Segredos:** Nenhuma API key ou credencial commitada.
