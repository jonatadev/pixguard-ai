import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const docsSiteRoot = path.resolve(__dirname, '..');
const repoRoot = path.resolve(docsSiteRoot, '..');
const contentDocsRoot = path.join(docsSiteRoot, 'src', 'content', 'docs');
const publicRoot = path.join(docsSiteRoot, 'public');

const GITHUB_REPO = 'https://github.com/jonatadev/pixguard-ai';
const BASE_URL = '/pixguard-ai';
const SITE_ORIGIN = 'https://jonatadev.github.io';

const DOC_SOURCES = [
  {
    src: 'specs/README.md',
    slug: 'specs/overview',
    title: 'Índice Spec-Driven (specs/)',
    description: 'Índice mestre das especificações técnicas, ADRs e ciclo Spec-First do PixGuard AI.',
    section: 'Especificações',
  },
  {
    src: 'specs/001-system-architecture.md',
    slug: 'specs/001-system-architecture',
    title: '001 — Arquitetura do Sistema',
    description: 'Limites de pacotes Go, schema PostgreSQL + pgvector, contratos HTTP e metas SLI/SLO.',
    section: 'Especificações',
  },
  {
    src: 'specs/002-multi-agent-orchestration.md',
    slug: 'specs/002-multi-agent-orchestration',
    title: '002 — Orquestração Multi-Agente & RAG',
    description: 'DAG controlado em Go (errgroup), Tool Calling determinístico, RAG híbrido RRF e tiering Gemini.',
    section: 'Especificações',
  },
  {
    src: 'specs/003-lgpd-security-guardrails.md',
    slug: 'specs/003-lgpd-security-guardrails',
    title: '003 — Guardrails LGPD & Segurança',
    description: 'Mapeamento OWASP LLM Top 10, precedência de redação PII e regras Bacen MED/PLD-FT.',
    section: 'Especificações',
  },
  {
    src: 'specs/004-llmops-eval-and-observability.md',
    slug: 'specs/004-llmops-eval-and-observability',
    title: '004 — LLMOps, Eval & Observabilidade',
    description: 'AI Eval Gate, tracing OTLP para Langfuse v3, Prometheus/Grafana/Loki e pipeline CI/CD.',
    section: 'Especificações',
  },
  {
    src: 'specs/adr/0001-tech-stack-and-ai-sdlc.md',
    slug: 'specs/adr-0001-tech-stack-and-ai-sdlc',
    title: 'ADR-0001 — Stack Tecnológica & AI SDLC',
    description: 'Decisão arquitetural por Go 1.24, Angular 19 Material, pgvector e observabilidade open-source.',
    section: 'Especificações',
  },
  {
    src: 'docs/regulatory/01-manual-pix-med-bacen.md',
    slug: 'regulatory/01-manual-pix-med-bacen',
    title: 'Manual PIX & MED (Bacen)',
    description: 'Referência regulatória simulada sobre elegibilidade ao MED, prazos operacionais e priorização de risco.',
    section: 'Corpus Regulatório',
  },
  {
    src: 'docs/regulatory/02-bloqueio-cautelar-fraude.md',
    slug: 'regulatory/02-bloqueio-cautelar-fraude',
    title: 'Bloqueio Cautelar Preventivo (72h)',
    description: 'Gatilhos objetivos de suspeita, retenção cautelar de 72 horas e conversão em MED.',
    section: 'Corpus Regulatório',
  },
  {
    src: 'docs/regulatory/03-politica-interna-antifraude-lgpd.md',
    slug: 'regulatory/03-politica-interna-antifraude-lgpd',
    title: 'Política Interna Antifraude, PLD-FT & LGPD',
    description: 'Proteção de dados LGPD, proteção contra Prompt Injection e alçadas Human-in-the-Loop/COAF.',
    section: 'Corpus Regulatório',
  },
  {
    src: 'AGENTS.md',
    slug: 'ai-native/agents-root',
    title: 'AGENTS.md — Diretrizes Globais',
    description: 'Princípios AI-Native, invariantes de segurança, portões de verificação e convenções Git.',
    section: 'Engenharia AI-Native',
  },
  {
    src: 'backend/AGENTS.md',
    slug: 'ai-native/agents-backend',
    title: 'AGENTS.md — Backend Go 1.24',
    description: 'Limites de pacotes Go, isolamento de métricas Prometheus, Zero-PII em logs e timeouts.',
    section: 'Engenharia AI-Native',
  },
  {
    src: 'frontend/AGENTS.md',
    slug: 'ai-native/agents-frontend',
    title: 'AGENTS.md — Frontend Angular 19',
    description: 'Standalone components, Signals, Angular Material 3, proxy local e UX Human-in-the-Loop.',
    section: 'Engenharia AI-Native',
  },
  {
    src: '.agents/skills/pixguard-spec-driven-dev/SKILL.md',
    slug: 'ai-native/skill-spec-driven-dev',
    title: 'Skill: pixguard-spec-driven-dev',
    description: 'Runbook de evolução Spec-First, criação de ADRs e sincronia de contratos Go/TS/SQL/YAML.',
    section: 'Engenharia AI-Native',
  },
  {
    src: '.agents/skills/pixguard-lgpd-guardrail/SKILL.md',
    slug: 'ai-native/skill-lgpd-guardrail',
    title: 'Skill: pixguard-lgpd-guardrail',
    description: 'Runbook para manutenção de regex LGPD, detecção de Prompt Injection e Eval red-team.',
    section: 'Engenharia AI-Native',
  },
  {
    src: '.agents/skills/pixguard-prompt-eval-ops/SKILL.md',
    slug: 'ai-native/skill-prompt-eval-ops',
    title: 'Skill: pixguard-prompt-eval-ops',
    description: 'Runbook para versionamento de prompts YAML, corpus regulatório e AI Eval Gate.',
    section: 'Engenharia AI-Native',
  },
  {
    src: '.agents/skills/pixguard-observability-ops/SKILL.md',
    slug: 'ai-native/skill-observability-ops',
    title: 'Skill: pixguard-observability-ops',
    description: 'Runbook para Docker Compose profiles, OTLP Langfuse v3, Prometheus, Grafana e Loki.',
    section: 'Engenharia AI-Native',
  },
  {
    src: '.agents/skills/pixguard-release-conventional/SKILL.md',
    slug: 'ai-native/skill-release-conventional',
    title: 'Skill: pixguard-release-conventional',
    description: 'Runbook de Conventional Commits, SemVer, CHANGELOG.md e checklist de Pull Request.',
    section: 'Engenharia AI-Native',
  },
  {
    src: 'CONTRIBUTING.md',
    slug: 'governance/contributing',
    title: 'Guia de Contribuição',
    description: 'Fluxo Spec-First, convenção de branches, Conventional Commits e portões de verificação.',
    section: 'Governança',
  },
  {
    src: 'CHANGELOG.md',
    slug: 'governance/changelog',
    title: 'Changelog (SemVer)',
    description: 'Histórico de versões e mudanças do PixGuard AI no padrão Keep a Changelog.',
    section: 'Governança',
  },
];

const ROUTE_MAP = new Map(DOC_SOURCES.map((doc) => [doc.src, `${BASE_URL}/${doc.slug}/`]));
const DIRECTORY_SHORTCUTS = new Map([
  ['specs', `${BASE_URL}/specs/overview/`],
  ['docs/regulatory', `${BASE_URL}/regulatory/01-manual-pix-med-bacen/`],
]);

function stripFrontmatterAndFirstH1(raw) {
  let text = raw.replace(/\r\n/g, '\n');
  let extractedDescription = '';

  const fmMatch = text.match(/^---\n([\s\S]*?)\n---\n?/);
  if (fmMatch) {
    const descMatch = fmMatch[1].match(/description:\s*(?:>-\s*\n([\s\S]*?)|["']?(.+?)["']?)(?:\n\S|\n*$)/);
    if (descMatch) {
      extractedDescription = (descMatch[1] || descMatch[2] || '')
        .replace(/\n\s+/g, ' ')
        .trim();
    }
    text = text.slice(fmMatch[0].length);
  }

  text = text.replace(/^\s*#\s+[^\n]+\n+/, '');
  return { body: text.trimStart(), extractedDescription };
}

function rewriteRelativeLink(rawTarget, sourceRepoPath) {
  if (
    !rawTarget ||
    rawTarget.startsWith('http://') ||
    rawTarget.startsWith('https://') ||
    rawTarget.startsWith('mailto:') ||
    rawTarget.startsWith('#')
  ) {
    return rawTarget;
  }

  const hashIdx = rawTarget.indexOf('#');
  const cleanTarget = hashIdx >= 0 ? rawTarget.slice(0, hashIdx) : rawTarget;
  const anchor = hashIdx >= 0 ? rawTarget.slice(hashIdx) : '';

  if (!cleanTarget) {
    return rawTarget;
  }

  const sourceDir = path.posix.dirname(sourceRepoPath);
  const resolved = path.posix.normalize(
    sourceDir === '.' ? cleanTarget : path.posix.join(sourceDir, cleanTarget)
  );
  const trimmed = resolved.replace(/\/$/, '');

  if (ROUTE_MAP.has(trimmed)) {
    return `${ROUTE_MAP.get(trimmed)}${anchor}`;
  }
  if (DIRECTORY_SHORTCUTS.has(trimmed)) {
    return `${DIRECTORY_SHORTCUTS.get(trimmed)}${anchor}`;
  }

  const absOnDisk = path.join(repoRoot, trimmed);
  if (fs.existsSync(absOnDisk)) {
    const isDir = fs.statSync(absOnDisk).isDirectory();
    const kind = isDir ? 'tree' : 'blob';
    return `${GITHUB_REPO}/${kind}/main/${trimmed}${anchor}`;
  }

  return `${GITHUB_REPO}/blob/main/${trimmed}${anchor}`;
}

function rewriteLinksOutsideCodeFences(markdown, sourceRepoPath) {
  const lines = markdown.split('\n');
  let inFence = false;
  let fenceChar = '';

  return lines
    .map((line) => {
      const fenceMatch = line.match(/^\s*(```+|~~~+)/);
      if (fenceMatch) {
        const marker = fenceMatch[1][0];
        if (!inFence) {
          inFence = true;
          fenceChar = marker;
        } else if (marker === fenceChar) {
          inFence = false;
          fenceChar = '';
        }
        return line;
      }

      if (inFence) {
        return line;
      }

      return line.replace(/(!?\[[^\]]*\]\()([^)\s]+)(\))/g, (full, prefix, url, suffix) => {
        if (prefix.startsWith('!')) {
          return full;
        }
        const rewritten = rewriteRelativeLink(url, sourceRepoPath);
        return `${prefix}${rewritten}${suffix}`;
      });
    })
    .join('\n');
}

function buildFrontmatter({ title, description, sourcePath }) {
  return [
    '---',
    `title: ${JSON.stringify(title)}`,
    `description: ${JSON.stringify(description)}`,
    'editUrl: false',
    '---',
    '',
    `> **Fonte Canônica:** [\`${sourcePath}\`](${GITHUB_REPO}/blob/main/${sourcePath}) — sincronizado automaticamente por \`docs-site/scripts/sync-docs.mjs\`.`,
    '',
  ].join('\n');
}

function generateLlmsTxt() {
  const sections = new Map();
  for (const doc of DOC_SOURCES) {
    if (!sections.has(doc.section)) {
      sections.set(doc.section, []);
    }
    sections.get(doc.section).push(doc);
  }

  const lines = [
    '# PixGuard AI',
    '',
    '> Copiloto Multi-Agente de Contestação PIX (MED), Bloqueio Cautelar, PLD-FT e Guardrail LGPD em Go 1.24 + Angular 19 Material + PostgreSQL/pgvector + Langfuse v3.',
    '',
    `- Repositório: ${GITHUB_REPO}`,
    `- Portal Web: ${SITE_ORIGIN}${BASE_URL}/`,
    `- Guia Rápido: ${SITE_ORIGIN}${BASE_URL}/guides/getting-started/`,
    '',
  ];

  for (const [sectionName, docs] of sections.entries()) {
    lines.push(`## ${sectionName}`);
    lines.push('');
    for (const doc of docs) {
      lines.push(`- [${doc.title}](${SITE_ORIGIN}${BASE_URL}/${doc.slug}/): ${doc.description}`);
    }
    lines.push('');
  }

  lines.push('## Prompts & Schemas no Repositório');
  lines.push('');
  lines.push(`- [prompts/](${GITHUB_REPO}/tree/main/prompts): Prompts YAML versionados com output_contract`);
  lines.push(`- [backend/migrations/001_init.sql](${GITHUB_REPO}/blob/main/backend/migrations/001_init.sql): Schema PostgreSQL + pgvector + FTS`);
  lines.push(`- [deploy/docker-compose.yml](${GITHUB_REPO}/blob/main/deploy/docker-compose.yml): Stack core, monitoring (Prometheus/Grafana/Loki) e Langfuse v3`);
  lines.push('');

  fs.mkdirSync(publicRoot, { recursive: true });
  fs.writeFileSync(path.join(publicRoot, 'llms.txt'), lines.join('\n'), 'utf8');
}

function main() {
  const generatedDirs = ['specs', 'regulatory', 'ai-native', 'governance'];
  for (const dir of generatedDirs) {
    fs.rmSync(path.join(contentDocsRoot, dir), { recursive: true, force: true });
  }

  for (const doc of DOC_SOURCES) {
    const absSrc = path.join(repoRoot, doc.src);
    if (!fs.existsSync(absSrc)) {
      throw new Error(`Missing source document: ${absSrc}`);
    }

    const raw = fs.readFileSync(absSrc, 'utf8');
    const { body, extractedDescription } = stripFrontmatterAndFirstH1(raw);
    const rewrittenBody = rewriteLinksOutsideCodeFences(body, doc.src);
    const description = doc.description || extractedDescription || doc.title;

    const absDest = path.join(contentDocsRoot, `${doc.slug}.md`);
    fs.mkdirSync(path.dirname(absDest), { recursive: true });
    const output = `${buildFrontmatter({
      title: doc.title,
      description,
      sourcePath: doc.src,
    })}${rewrittenBody.trimEnd()}\n`;
    fs.writeFileSync(absDest, output, 'utf8');
  }

  generateLlmsTxt();
  console.log(`Synced ${DOC_SOURCES.length} docs and generated public/llms.txt`);
}

main();
