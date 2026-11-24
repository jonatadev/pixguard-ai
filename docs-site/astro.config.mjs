import { defineConfig } from 'astro/config';
import mermaid from 'astro-mermaid';
import starlight from '@astrojs/starlight';

export default defineConfig({
  site: 'https://jonatadev.github.io',
  base: '/pixguard-ai',
  integrations: [
    mermaid({ autoTheme: true }),
    starlight({
      title: 'PixGuard AI',
      description:
        'Documentação AI-Native, Specs, Guardrails LGPD, Multi-Agente e LLMOps do PixGuard AI',
      defaultLocale: 'root',
      locales: {
        root: {
          label: 'Português (BR)',
          lang: 'pt-BR',
        },
      },
      logo: {
        src: './src/assets/pixguard-logo.svg',
        replacesTitle: false,
      },
      social: [
        {
          icon: 'github',
          label: 'GitHub',
          href: 'https://github.com/jonatadev/pixguard-ai',
        },
      ],
      editLink: {
        baseUrl: 'https://github.com/jonatadev/pixguard-ai/edit/main/docs-site/',
      },
      customCss: ['./src/styles/custom.css'],
      sidebar: [
        {
          label: 'Início',
          items: [
            { label: 'Visão Geral', slug: 'index' },
            { label: 'Guia Rápido Local', slug: 'guides/getting-started' },
          ],
        },
        {
          label: 'Especificações (specs/)',
          items: [
            { label: 'Índice Spec-Driven', slug: 'specs/overview' },
            { label: '001 — Arquitetura do Sistema', slug: 'specs/001-system-architecture' },
            { label: '002 — Orquestração Multi-Agente', slug: 'specs/002-multi-agent-orchestration' },
            { label: '003 — Guardrails LGPD & Segurança', slug: 'specs/003-lgpd-security-guardrails' },
            { label: '004 — LLMOps, Eval & Observabilidade', slug: 'specs/004-llmops-eval-and-observability' },
            { label: 'ADR-0001 — Stack & AI SDLC', slug: 'specs/adr-0001-tech-stack-and-ai-sdlc' },
          ],
        },
        {
          label: 'Corpus Regulatório (RAG)',
          items: [
            { label: 'Manual PIX & MED (Bacen)', slug: 'regulatory/01-manual-pix-med-bacen' },
            { label: 'Bloqueio Cautelar (72h)', slug: 'regulatory/02-bloqueio-cautelar-fraude' },
            { label: 'Política Antifraude, PLD-FT & LGPD', slug: 'regulatory/03-politica-interna-antifraude-lgpd' },
          ],
        },
        {
          label: 'Engenharia AI-Native',
          items: [
            { label: 'AGENTS.md — Raiz', slug: 'ai-native/agents-root' },
            { label: 'AGENTS.md — Backend Go', slug: 'ai-native/agents-backend' },
            { label: 'AGENTS.md — Frontend Angular', slug: 'ai-native/agents-frontend' },
            { label: 'Skill: Spec-Driven Dev', slug: 'ai-native/skill-spec-driven-dev' },
            { label: 'Skill: Guardrail LGPD', slug: 'ai-native/skill-lgpd-guardrail' },
            { label: 'Skill: Prompt & Eval Ops', slug: 'ai-native/skill-prompt-eval-ops' },
            { label: 'Skill: Observability Ops', slug: 'ai-native/skill-observability-ops' },
            { label: 'Skill: Release & Conventional', slug: 'ai-native/skill-release-conventional' },
          ],
        },
        {
          label: 'Governança & Releases',
          items: [
            { label: 'Guia de Contribuição', slug: 'governance/contributing' },
            { label: 'Changelog (SemVer)', slug: 'governance/changelog' },
            { label: 'Índice AI-Native (llms.txt)', link: '/pixguard-ai/llms.txt' },
          ],
        },
      ],
    }),
  ],
});
