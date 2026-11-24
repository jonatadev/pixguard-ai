package domain

import "time"

type RiskLevel string

const (
	RiskBaixo RiskLevel = "BAIXO"
	RiskMedio RiskLevel = "MEDIO"
	RiskAlto  RiskLevel = "ALTO"
)

type CaseStatus string

const (
	StatusPendente  CaseStatus = "PENDENTE"
	StatusEmAnalise CaseStatus = "EM_ANALISE"
	StatusAprovado  CaseStatus = "APROVADO"
	StatusRejeitado CaseStatus = "REJEITADO"
)

// PixCase represents a Brazilian Pix MED dispute submitted to the analysis desk.
type PixCase struct {
	ID                      string     `json:"id"`
	CustomerAlias           string     `json:"customerAlias"`
	AmountBRL               float64    `json:"amountBrl"`
	DaysSinceTx             int        `json:"daysSinceTx"`
	TransactionAt           time.Time  `json:"transactionAt"`
	PixKeyType              string     `json:"pixKeyType"`
	RecipientBank           string     `json:"recipientBank"`
	RecipientAccountAgeDays int        `json:"recipientAccountAgeDays"`
	RawDescription          string     `json:"rawDescription"`
	RiskLevel               RiskLevel  `json:"riskLevel"`
	Status                  CaseStatus `json:"status"`
	MedEligibleExpected     bool       `json:"medEligibleExpected"`
}

type GuardrailFinding struct {
	Category string `json:"category"`
	Rule     string `json:"rule"`
	Count    int    `json:"count"`
}

// GuardrailReport summarizes LGPD PII redactions and prompt-injection checks.
type GuardrailReport struct {
	PIIRedactionsCount      int                `json:"piiRedactionsCount"`
	PIICategories           map[string]int     `json:"piiCategories"`
	PromptInjectionDetected bool               `json:"promptInjectionDetected"`
	InjectionMatches        []string           `json:"injectionMatches"`
	Findings                []GuardrailFinding `json:"findings"`
	SafeForLLM              bool               `json:"safeForLlm"`
}

type CaseDetailResponse struct {
	Case                 PixCase         `json:"case"`
	SanitizedDescription string          `json:"sanitizedDescription"`
	GuardrailReport      GuardrailReport `json:"guardrailReport"`
	MedWindowValid       bool            `json:"medWindowValid"`
}

type GuardrailPreviewRequest struct {
	Text string `json:"text"`
}

type GuardrailPreviewResponse struct {
	OriginalLength int             `json:"originalLength"`
	SanitizedText  string          `json:"sanitizedText"`
	Report         GuardrailReport `json:"report"`
}

type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Env     string `json:"env"`
}

type ReadyResponse struct {
	Status       string            `json:"status"`
	Integrations map[string]bool   `json:"integrations"`
	Models       map[string]string `json:"models"`
}
