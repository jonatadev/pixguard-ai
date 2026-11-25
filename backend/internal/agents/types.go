package agents

import (
	"time"

	"github.com/jonatadev/pixguard-ai/internal/domain"
	"github.com/jonatadev/pixguard-ai/internal/rag"
)

// TokenUsage tracks LLM prompt and completion token counts.
type TokenUsage struct {
	PromptTokens     int32 `json:"promptTokens"`
	CompletionTokens int32 `json:"completionTokens"`
	TotalTokens      int32 `json:"totalTokens"`
}

// ToolTrace records a deterministic read-only tool invocation made by FraudInvestigatorAgent.
type ToolTrace struct {
	Name       string         `json:"name"`
	Summary    string         `json:"summary"`
	Data       map[string]any `json:"data"`
	DurationMs int64          `json:"durationMs"`
}

// InvestigatorOutput matches prompts/fraud-investigator.v1.yaml output_contract.
type InvestigatorOutput struct {
	Summary              string           `json:"summary"`
	MedWindowValid       bool             `json:"med_window_valid"`
	Signals              []string         `json:"signals"`
	RecommendedRiskLevel domain.RiskLevel `json:"recommended_risk_level"`
	ToolTraces           []ToolTrace      `json:"tool_traces,omitempty"`
}

// CitedArticle identifies a regulatory document and section/article used for grounding.
type CitedArticle struct {
	DocID       string `json:"docId"`
	HeadingPath string `json:"headingPath"`
	Article     string `json:"article"`
}

// RegulatoryOutput matches prompts/regulatory-rag.v1.yaml output_contract.
type RegulatoryOutput struct {
	EligibleByRule  bool               `json:"eligible_by_rule"`
	CitedArticles   []CitedArticle     `json:"cited_articles"`
	RegulatoryNotes string             `json:"regulatory_notes"`
	RetrievedChunks []rag.SearchResult `json:"retrieved_chunks,omitempty"`
}

// AdjudicatorOutput matches prompts/adjudicator.v1.yaml output_contract.
type AdjudicatorOutput struct {
	ElegivelMED        bool             `json:"elegivel_med"`
	NivelRisco         domain.RiskLevel `json:"nivel_risco"`
	AcaoRecomendada    string           `json:"acao_recomendada"`
	JustificativaCitada string          `json:"justificativa_citada"`
	MensagemCliente    string           `json:"mensagem_cliente"`
	SinalizarPLDFT     bool             `json:"sinalizar_pldft"`
}

// CriticOutput matches prompts/compliance-critic.v1.yaml output_contract.
type CriticOutput struct {
	Approved   bool            `json:"approved"`
	Checks     map[string]bool `json:"checks"`
	Violations []string        `json:"violations"`
}

// AnalysisEvent represents a real-time SSE progress event for a pipeline stage.
type AnalysisEvent struct {
	Stage      string    `json:"stage"`
	Status     string    `json:"status"`
	Message    string    `json:"message"`
	LatencyMs  int64     `json:"latencyMs"`
	Timestamp  time.Time `json:"timestamp"`
}

// AnalysisResult is the complete multi-agent decision bundle returned by POST /analyze.
type AnalysisResult struct {
	CaseID                      string                 `json:"caseId"`
	TraceID                     string                 `json:"traceId"`
	Mode                        string                 `json:"mode"`
	SanitizedDescription        string                 `json:"sanitizedDescription"`
	GuardrailReport             domain.GuardrailReport `json:"guardrailReport"`
	RequiresPriorityHumanReview bool                   `json:"requiresPriorityHumanReview"`
	Investigator                InvestigatorOutput     `json:"investigator"`
	Regulatory                  RegulatoryOutput       `json:"regulatory"`
	Adjudicator                 AdjudicatorOutput      `json:"adjudicator"`
	Critic                      CriticOutput           `json:"critic"`
	TokenUsage                  TokenUsage             `json:"tokenUsage"`
	TotalDurationMs             int64                  `json:"totalDurationMs"`
}
