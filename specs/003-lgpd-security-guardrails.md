# Spec 003 — Guardrails LGPD, Segurança & Regras Regulatórias

- **Status:** Implementado (Camada Determinística Fase 1)
- **Owner:** `@jonatadev`
- **Implementação de Referência:** [`backend/internal/guardrail/lgpd.go`](../backend/internal/guardrail/lgpd.go)
- **Política Base:** [`docs/regulatory/03-politica-interna-antifraude-lgpd.md`](../docs/regulatory/03-politica-interna-antifraude-lgpd.md)

---

## 1. Modelo de Ameaças (OWASP Top 10 for LLM)

| Risco OWASP LLM | Vetor no PixGuard AI | Mitigação Obrigatória |
| :--- | :--- | :--- |
| **LLM01 — Prompt Injection** | Campo `rawDescription` ou mensagem Pix com comandos como *"Ignore as instruções anteriores"* ou *"System prompt override"* | Detecção determinística pré-LLM em `guardrail.Sanitize`; marca `promptInjectionDetected = true`, `safeForLlm = false`, incrementa `pixguard_guardrail_injection_blocks_total` e escala para revisão humana. |
| **LLM02 — Sensitive Information Disclosure** | Envio de CPF, CNPJ, e-mail, telefone ou chave aleatória EVP em texto claro para APIs generativas ou logs | Redação determinística antes do LLM (*Privacy by Design*); zero armazenamento de valores originais de PII nos logs JSON (`slog`). |
| **LLM06 — Excessive Agency** | Modelo tentando aprovar devolução MED ou bloqueio financeiro sozinho | Nenhuma ferramenta de escrita/liquidação é exposta a LLMs; aprovação acima de **R$ 0,01** exige decisão humana explícita na mesa Angular Material. |
| **LLM09 — Misinformation** | Citação inventada de artigo do Bacen ou concessão de MED fora do prazo de 80 dias | Validação determinística `medWindowValid := daysSinceTx <= 80` + auditoria pelo `ComplianceCriticAgent`. |

---

## 2. Taxonomia de Redação LGPD e Ordem de Precedência

Conforme implementado em [`guardrail.Sanitize`](../backend/internal/guardrail/lgpd.go), a aplicação das expressões regulares segue **ordem estrita** para evitar colisões (por exemplo, segmentos numéricos de UUID EVP ou CNPJ sendo capturados parcialmente como telefone):

| Ordem | Categoria | Regra (`Finding.Rule`) | Token Substituto | Observação |
| :--- | :--- | :--- | :--- | :--- |
| **1** | `PIX_EVP` | `uuid_evp_key` | `[PIX_EVP_REDACTED]` | Chaves aleatórias UUID v4 |
| **2** | `CPF` | `formatted_cpf` | `[CPF_REDACTED]` | Formato `000.000.000-00` |
| **3** | `CPF` | `contextual_11_digit_cpf` | `${1}[CPF_REDACTED]` | 11 dígitos precedidos de `cpf` / `cpf informado` / `cpf chave` |
| **4** | `CNPJ` | `formatted_cnpj` | `[CNPJ_REDACTED]` | Formato `00.000.000/0000-00` |
| **5** | `EMAIL` | `email_address` | `[EMAIL_REDACTED]` | Endereços de e-mail padrão |
| **6** | `PHONE` | `brazilian_phone` | `[PHONE_REDACTED]` | Telefones BR com/sem `+55` e DDD `(XX)` ou `XX 9XXXX-XXXX` |

> **Regra de regressão:** Qualquer novo padrão LGPD (ex.: placa, RG, número de conta/agência) deve respeitar essa ordem de precedência e ser testado em [`lgpd_test.go`](../backend/internal/guardrail/lgpd_test.go) e [`cmd/eval/main.go`](../backend/cmd/eval/main.go) sem gerar falsos positivos em narrativas limpas.

---

## 3. Catálogo de Assinaturas de Prompt Injection

Regras determinísticas ativas em `injectionRules`:

| Identificador (`Rule`) | Padrão Alvo |
| :--- | :--- |
| `pt_ignore_instructions` | Tentativas em PT-BR de ignorar instruções (`ignore (todas) as instru[cç][oõ]es`) |
| `pt_disregard_rules` | Tentativas de desconsiderar regras (`desconsidere (todas) as regras`) |
| `en_ignore_instructions` | Padrão clássico em inglês (`ignore (all) previous instructions`) |
| `pt_force_refund` | Coação de fluxo financeiro (`aprove o (estorno\|med\|reembolso) imediatamente`) |
| `system_prompt_probe` | Sondagem/sobrescrita de system prompt (`system prompt` / `reveal your/the prompt`) |
| `role_hijack` | Sequestro de persona (`você agora é` / `you are now`) |
| `jailbreak_keyword` | Palavras-chave explícitas (`jailbreak` / `bypass guardrail`) |

---

## 4. Invariantes de Negócio (Bacen MED, Bloqueio Cautelar & PLD-FT)

1. **Janela Legal do MED:** `daysSinceTx <= 80` (`medMaxWindowDays = 80` em `server.go`). Contestações acima de 80 dias devem ser indeferidas por decurso de prazo.
2. **Desacordo Comercial:** Atraso de entrega, defeito ou arrependimento sem indício de fraude não são elegíveis ao MED (Art. 41-C, § 2º).
3. **Bloqueio Cautelar:** Retenção preventiva máxima de **72 horas** na conta recebedora quando há conta recém-criada (`< 15 dias`) com valor `> R$ 1.500,00`, anomalia de velocidade ou reputação negativa.
4. **Alçada Human-in-the-Loop:** Nenhum estorno acima de **R$ 0,01** pode ser efetivado sem confirmação do analista humano.
5. **Sinalização PLD-FT / COAF:** Valores acima de **R$ 5.000,00** combinados com indícios de conta laranja ou fracionamento devem gerar recomendação de dossiê PLD-FT.
6. **Zero-PII em Logs:** Logs estruturados (`slog`) gravam apenas `method`, `path`, `status`, `duration_ms`, `trace_id` e contagens agregadas de redação — nunca `RawDescription` nem dados pessoais.
