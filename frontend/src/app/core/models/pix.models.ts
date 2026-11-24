export type RiskLevel = 'BAIXO' | 'MEDIO' | 'ALTO';
export type CaseStatus = 'PENDENTE' | 'EM_ANALISE' | 'APROVADO' | 'REJEITADO';

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
