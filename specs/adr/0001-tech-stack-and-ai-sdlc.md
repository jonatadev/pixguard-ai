# ADR-0001: Stack Tecnológica e Arquitetura AI-Native SDLC do PixGuard AI

- **Status:** Aceito
- **Data:** 2026-11-24
- **Decisores:** `@jonatadev`

---

## Contexto

O PixGuard AI é uma plataforma de engenharia de IA com **SDLC completo (End-to-End LLMOps)** voltada para o domínio de contestações **PIX / Mecanismo Especial de Devolução (MED)**, **Bloqueio Cautelar**, **PLD-FT** e **LGPD** em bancos digitais brasileiros.

O projeto precisa demonstrar práticas reais de produção corporativa:
1. Baixa latência, tipagem forte e concorrência previsível no backend de pagamentos/guardrails.
2. Interface operacional de mesa de análise (*back-office*) padronizada e responsiva.
3. Persistência transacional, busca vetorial e busca lexical (*Full-Text Search*) sem proliferação desnecessária de infraestrutura.
4. Observabilidade open-source completa cobrindo tanto infraestrutura quanto traces/avaliação de LLMs.

---

## Decisão

Adotar a seguinte stack oficial:
1. **Backend em Go 1.24** com SDK oficial **`google.golang.org/genai`**, roteador `net/http` padrão, concorrência via `errgroup`, logs estruturados em `log/slog` e container `distroless`.
2. **Frontend em Angular 19 + Angular Material (Material 3)** com *standalone components* e Signals.
3. **Banco de Dados em PostgreSQL 16 + `pgvector`** unificando estado transacional (`pix_cases`, `agent_runs`, `analyst_feedback`), embeddings `vector(768)` e busca lexical `tsvector('portuguese')`.
4. **Observabilidade LLM & Infra 100% Open-Source:** **Langfuse v3** (via OpenTelemetry OTLP HTTP) + **Prometheus + Grafana + Loki** orquestrados em `deploy/docker-compose.yml` com profiles modulares (`monitoring`, `langfuse`).
5. **Engenharia AI-Native Orientada a Specs:** Especificações em `specs/`, regras hierárquicas em `AGENTS.md`, skills do Antigravity em `.agents/skills/` e prompts versionados em `prompts/*.yaml`.

---

## Consequências

### Positivas
- **Concorrência e segurança em produção:** Go 1.24 executa o guardrail determinístico em microssegundos/milissegundos, orquestra agentes em paralelo com cancelamento via `context.Context` e gera binário estático `distroless` amigável ao `Trivy`.
- **Simplicidade operacional:** PostgreSQL + `pgvector` elimina sincronização eventual entre um banco relacional e um vector DB externo para o corpus regulatório e histórico transacional.
- **Portabilidade local:** Profiles no Docker Compose permitem rodar apenas o core em máquinas com menor memória RAM ou subir toda a observabilidade com um comando.
- **Auditabilidade bancária:** Toda execução produz `trace_id` correlacionado entre `slog`, Prometheus/Grafana e Langfuse.

### Negativas / Trade-offs
- **Ecossistema de agentes em Go:** Exige implementar explicitamente a orquestração (DAG, RRF, tool loop) em vez de depender de abstrações prontas de frameworks Python — escolha intencional para garantir controle determinístico e transparência.
- **Overhead do Langfuse v3 local:** Requer ClickHouse, Redis e MinIO no profile `langfuse`, exigindo isolamento em profile opcional do Docker Compose.

---

## Alternativas Consideradas e Descartadas

| Alternativa | Motivo da Recusa |
| :--- | :--- |
| **Python + LangChain / LlamaIndex** | Menor aderência ao ecossistema Go de motores Pix de alta concorrência e maior superfície/tamanho de imagem para DevSecOps, além de abstrações opacas que dificultam auditoria determinística. |
| **Qdrant / Pinecone dedicado junto com Postgres** | Adiciona mais um serviço stateful para manter consistência transacional com `regulatory_docs` e `pix_cases` na fase inicial. |
| **SaaS proprietário de observabilidade LLM** | Contraria o requisito de stack open-source self-hosted com controle total de dados sensíveis sob LGPD. |
