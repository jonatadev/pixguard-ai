export type RiskLevel = 'BAIXO' | 'MEDIO' | 'ALTO';
export type CaseStatus = 'PENDENTE' | 'EM_ANALISE' | 'APROVADO' | 'REJEITADO';
export type AnalystDecision = 'APROVADO' | 'EDITADO' | 'REJEITADO';

export interface PixCase {
  id: string;
  customerAlias: string;
  amountBrl: number;
  daysSinceTx: number;
  transactionAt: string;
  pixKeyType: string;
  recipientBank: string;
  recipientAccountAgeDays: number;
  rawDescription: string;
  riskLevel: RiskLevel;
  status: CaseStatus;
  medEligibleExpected: boolean;
}

export interface GuardrailFinding {
  category: string;
  rule: string;
  count: number;
}

export interface GuardrailReport {
  piiRedactionsCount: number;
  piiCategories: Record<string, number>;
  promptInjectionDetected: boolean;
  injectionMatches: string[];
  findings: GuardrailFinding[];
  safeForLlm: boolean;
}

export interface CaseListResponse {
  cases: PixCase[];
  total: number;
}

export interface CaseDetailResponse {
  case: PixCase;
  sanitizedDescription: string;
  guardrailReport: GuardrailReport;
  medWindowValid: boolean;
}

export interface GuardrailPreviewResponse {
  originalLength: number;
  sanitizedText: string;
  report: GuardrailReport;
}

export interface HealthResponse {
  status: string;
  service: string;
  env: string;
}

export interface ReadyResponse {
  status: string;
  integrations: {
    gemini: boolean;
    langfuse: boolean;
    database: boolean;
  };
  models: {
    chat: string;
    fast: string;
    embedding: string;
  };
}

export interface AnalysisEvent {
  stage: string;
  status: string;
  message: string;
  latencyMs: number;
  timestamp: string;
}

export interface CitedArticle {
  docId: string;
  headingPath: string;
  article: string;
}

export interface AdjudicatorOutput {
  elegivel_med: boolean;
  nivel_risco: RiskLevel;
  acao_recomendada: string;
  justificativa_citada: string;
  mensagem_cliente: string;
  sinalizar_pldft: boolean;
}

export interface CriticOutput {
  approved: boolean;
  checks: Record<string, boolean>;
  violations: string[];
}

export interface AnalysisResult {
  caseId: string;
  traceId: string;
  mode: string;
  sanitizedDescription: string;
  guardrailReport: GuardrailReport;
  requiresPriorityHumanReview: boolean;
  investigator: {
    summary: string;
    med_window_valid: boolean;
    signals: string[];
    recommended_risk_level: RiskLevel;
  };
  regulatory: {
    eligible_by_rule: boolean;
    cited_articles: CitedArticle[];
    regulatory_notes: string;
  };
  adjudicator: AdjudicatorOutput;
  critic: CriticOutput;
  tokenUsage: {
    promptTokens: number;
    completionTokens: number;
    totalTokens: number;
  };
  totalDurationMs: number;
}

export interface AnalystFeedbackRequest {
  traceId: string;
  decision: AnalystDecision;
  notes: string;
}

export interface AnalystFeedbackRecord {
  id: string;
  caseId: string;
  traceId: string;
  decision: AnalystDecision;
  notes: string;
  updatedStatus: CaseStatus;
  langfuseSynced: boolean;
  createdAt: string;
}
