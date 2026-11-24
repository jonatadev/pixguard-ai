# PixGuard AI — Portal de Documentação (Astro Starlight)

Portal estático construído com **Astro 6 + Starlight** e renderização nativa de diagramas **Mermaid** (`astro-mermaid`).

## Arquitetura Single Source of Truth

Para evitar duplicação de Markdown no repositório, os arquivos em `specs/`, `docs/regulatory/`, `AGENTS.md`, `.agents/skills/`, `CONTRIBUTING.md` e `CHANGELOG.md` são a fonte canônica.

Antes de iniciar o servidor ou compilar o site, o script [`scripts/sync-docs.mjs`](./scripts/sync-docs.mjs):
1. Sincroniza os 19 documentos canônicos para as pastas geradas em `src/content/docs/{specs,regulatory,ai-native,governance}/` (ignoradas no Git);
2. Reescreve links relativos para rotas internas do Starlight ou links diretos para o repositório GitHub;
3. Gera [`public/llms.txt`](./public/llms.txt) no padrão AI-Native.

## Comandos Locais (Windows PowerShell)

```powershell
cd docs-site
npm.cmd install
npm.cmd run dev     # http://localhost:4321/pixguard-ai/
npm.cmd run build   # Gera dist/ com busca Pagefind e llms.txt
```
