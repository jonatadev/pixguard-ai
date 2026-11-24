# Frontend Angular 19 + Angular Material — Diretrizes Locais (`frontend/AGENTS.md`)

Este arquivo complementa o [`AGENTS.md`](../AGENTS.md) da raiz para todo desenvolvimento Angular em `frontend/`.

---

## 1. Padrões de Arquitetura Angular

- **Angular 19 Standalone + Signals:** Use *standalone components*, `inject()` e estado reativo via `signal()` / `computed()`. Evite introduzir NgModules legados ou gerenciamento de estado externo desnecessário.
- **Angular Material (Material 3):** Reutilize componentes oficiais (`MatTable`, `MatTabs`, `MatChips`, `MatCard`, `MatFormField`, `MatProgressBar`, `MatToolbar`) mantendo consistência visual de back-office bancário.
- **Sincronia Estrita de Contratos:** Mantenha [`src/app/core/models/pix.models.ts`](./src/app/core/models/pix.models.ts) 100% alinhado aos campos JSON em camelCase de [`backend/internal/domain/models.go`](../backend/internal/domain/models.go). Todas as chamadas HTTP devem passar por [`src/app/core/services/pixguard-api.service.ts`](./src/app/core/services/pixguard-api.service.ts).

---

## 2. Regras de UX Bancária & Segurança (LGPD / Human-in-the-Loop)

1. **Visibilidade de Guardrails:** Na inspeção de caso e no simulador de guardrail, destaque visualmente:
   - O texto sanitizado (`sanitizedDescription` / `sanitizedText`) com tokens `[CPF_REDACTED]`, `[CNPJ_REDACTED]`, `[EMAIL_REDACTED]`, `[PHONE_REDACTED]`, `[PIX_EVP_REDACTED]`.
   - Alertas explícitos quando `promptInjectionDetected === true` ou `medWindowValid === false` (`daysSinceTx > 80`).
2. **Human-in-the-Loop:** Nunca implemente aprovação automática ou botão que efetive devolução financeira sem confirmação explícita do analista humano.
3. **Rotas de Proxy:** Em dev local use [`proxy.conf.json`](./proxy.conf.json) e em container preserve [`nginx.conf`](./nginx.conf) roteando `/api`, `/healthz`, `/readyz` e `/metrics` para o backend Go.

---

## 3. Comandos e Budget no Windows PowerShell

```powershell
cd frontend
npm.cmd install        # apenas quando package.json mudar
npm.cmd run build      # verifica compilação de produção (dist/frontend)
npm.cmd run start      # dev server em :4200 com proxy para :8080
```

- **Windows PowerShell:** Sempre utilize `npm.cmd` (ou `npx.cmd`) para evitar bloqueio de ExecutionPolicy em scripts `.ps1`.
- **Bundle Budget:** Monitore o aviso de budget inicial em `angular.json` ao adicionar novos módulos Material ou gráficos; prefira lazy loading de rotas/abas pesadas nas próximas fases.
