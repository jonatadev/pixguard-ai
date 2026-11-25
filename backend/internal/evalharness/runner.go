package evalharness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jonatadev/pixguard-ai/internal/guardrail"
)

// MEDMaxWindowDays enforces the Bacen 80-day MED eligibility deadline.
const MEDMaxWindowDays = 80

// CaseResult captures the evaluation outcome for a single dataset case.
type CaseResult struct {
	ID                 string   `json:"id"`
	Category           string   `json:"category"`
	Name               string   `json:"name"`
	Passed             bool     `json:"passed"`
	Failures           []string `json:"failures,omitempty"`
	SanitizedText      string   `json:"sanitizedText"`
	PIIRedactionsCount int      `json:"piiRedactionsCount"`
	PIILeaksCount      int      `json:"piiLeaksCount"`
	InjectionDetected  bool     `json:"injectionDetected"`
	InjectionMatches   []string `json:"injectionMatches,omitempty"`
	SafeForLLM         bool     `json:"safeForLlm"`
	MEDWindowValid     bool     `json:"medWindowValid"`
	InjectionMatched   bool     `json:"injectionMatched"`
	MEDWindowMatched   bool     `json:"medWindowMatched"`
	DurationMs         int64    `json:"durationMs"`
}

// SummaryMetrics aggregates pass rates and security/business invariants across the run.
type SummaryMetrics struct {
	Total             int     `json:"total"`
	Passed            int     `json:"passed"`
	Failed            int     `json:"failed"`
	PassRate          float64 `json:"passRate"`
	PIILeaks          int     `json:"piiLeaks"`
	InjectionCases    int     `json:"injectionCases"`
	InjectionDetected int     `json:"injectionDetected"`
	FalsePositives    int     `json:"falsePositives"`
	MEDWindowMatches  int     `json:"medWindowMatches"`
	DurationMs        int64   `json:"durationMs"`
}

// Report contains the full offline evaluation results and summary metrics.
type Report struct {
	DatasetName    string         `json:"datasetName"`
	DatasetVersion string         `json:"datasetVersion"`
	GeneratedAt    time.Time      `json:"generatedAt"`
	Summary        SummaryMetrics `json:"summary"`
	Results        []CaseResult   `json:"results"`
}

// Run executes all deterministic LGPD, prompt-injection, and MED window assertions on ds.
func Run(ds Dataset) Report {
	suiteStart := time.Now()
	results := make([]CaseResult, 0, len(ds.Cases))
	var summary SummaryMetrics
	summary.Total = len(ds.Cases)

	for _, tc := range ds.Cases {
		caseStart := time.Now()
		sanitized, gr := guardrail.Sanitize(tc.Input)
		medWindowValid := tc.DaysSinceTx <= MEDMaxWindowDays
		failures := make([]string, 0)
		leaks := 0

		for _, raw := range tc.ForbiddenRawSubstrings {
			if strings.Contains(sanitized, raw) {
				leaks++
				failures = append(failures, fmt.Sprintf("pii_leak:%s", raw))
			}
		}
		for _, tok := range tc.RequiredTokens {
			if !strings.Contains(sanitized, tok) {
				failures = append(failures, fmt.Sprintf("missing_token:%s", tok))
			}
		}
		if gr.PIIRedactionsCount < tc.ExpectedMinRedactions {
			failures = append(failures, fmt.Sprintf("redactions_below_min:%d<%d", gr.PIIRedactionsCount, tc.ExpectedMinRedactions))
		}

		injectionMatched := gr.PromptInjectionDetected == tc.ExpectInjection
		if !injectionMatched {
			failures = append(failures, fmt.Sprintf("injection_mismatch:expected=%v,got=%v", tc.ExpectInjection, gr.PromptInjectionDetected))
		}
		if gr.SafeForLLM != tc.ExpectSafeForLLM {
			failures = append(failures, fmt.Sprintf("safe_for_llm_mismatch:expected=%v,got=%v", tc.ExpectSafeForLLM, gr.SafeForLLM))
		}

		medWindowMatched := medWindowValid == tc.ExpectMEDWindowValid
		if !medWindowMatched {
			failures = append(failures, fmt.Sprintf("med_window_mismatch:expected=%v,got=%v", tc.ExpectMEDWindowValid, medWindowValid))
		}

		passed := len(failures) == 0
		if passed {
			summary.Passed++
		} else {
			summary.Failed++
		}
		summary.PIILeaks += leaks
		if tc.ExpectInjection {
			summary.InjectionCases++
			if gr.PromptInjectionDetected {
				summary.InjectionDetected++
			}
		} else if gr.PromptInjectionDetected {
			summary.FalsePositives++
		}
		if medWindowMatched {
			summary.MEDWindowMatches++
		}

		results = append(results, CaseResult{
			ID:                 tc.ID,
			Category:           tc.Category,
			Name:               tc.Name,
			Passed:             passed,
			Failures:           failures,
			SanitizedText:      sanitized,
			PIIRedactionsCount: gr.PIIRedactionsCount,
			PIILeaksCount:      leaks,
			InjectionDetected:  gr.PromptInjectionDetected,
			InjectionMatches:   gr.InjectionMatches,
			SafeForLLM:         gr.SafeForLLM,
			MEDWindowValid:     medWindowValid,
			InjectionMatched:   injectionMatched,
			MEDWindowMatched:   medWindowMatched,
			DurationMs:         time.Since(caseStart).Milliseconds(),
		})
	}

	if summary.Total > 0 {
		summary.PassRate = float64(summary.Passed) * 100 / float64(summary.Total)
	}
	summary.DurationMs = time.Since(suiteStart).Milliseconds()

	return Report{
		DatasetName:    ds.Name,
		DatasetVersion: ds.Version,
		GeneratedAt:    time.Now().UTC(),
		Summary:        summary,
		Results:        results,
	}
}

// RenderMarkdown formats an AI Eval Gate report suitable for PR comments and job summaries.
func RenderMarkdown(report Report) string {
	var b strings.Builder
	statusBadge := "PASS"
	if report.Summary.Failed > 0 {
		statusBadge = "FAIL"
	}

	fmt.Fprintf(&b, "## PixGuard AI — Eval & Red-Team Harness (%s)\n\n", statusBadge)
	fmt.Fprintf(&b, "- **Dataset:** `%s` (`v%s`)\n", report.DatasetName, report.DatasetVersion)
	fmt.Fprintf(&b, "- **Aprovados:** **%d/%d** (`%.1f%%`)\n", report.Summary.Passed, report.Summary.Total, report.Summary.PassRate)
	fmt.Fprintf(&b, "- **Vazamentos de PII (LGPD):** `%d` (meta: `0`)\n", report.Summary.PIILeaks)
	fmt.Fprintf(&b, "- **Prompt Injections Detectados:** `%d/%d` (falsos positivos: `%d`)\n",
		report.Summary.InjectionDetected, report.Summary.InjectionCases, report.Summary.FalsePositives)
	fmt.Fprintf(&b, "- **Regra de Janela MED (80 dias):** `%d/%d`\n\n", report.Summary.MEDWindowMatches, report.Summary.Total)

	b.WriteString("| ID | Categoria | Caso | Redações PII | Injection | Janela MED | Status |\n")
	b.WriteString("| :--- | :--- | :--- | :---: | :---: | :---: | :---: |\n")
	for _, r := range report.Results {
		st := "PASS"
		if !r.Passed {
			st = "FAIL (" + strings.Join(r.Failures, ", ") + ")"
		}
		fmt.Fprintf(&b, "| `%s` | `%s` | %s | %d | %v | %v | **%s** |\n",
			r.ID, r.Category, r.Name, r.PIIRedactionsCount, r.InjectionDetected, r.MEDWindowValid, st)
	}
	return b.String()
}

// WriteJSON persists the formatted JSON evaluation report to path.
func WriteJSON(path string, report Report) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create json output dir: %w", err)
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal report json: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write report json %q: %w", path, err)
	}
	return nil
}

// WriteMarkdown persists the Markdown evaluation summary to path.
func WriteMarkdown(path string, report Report) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create markdown output dir: %w", err)
	}
	if err := os.WriteFile(path, []byte(RenderMarkdown(report)), 0o644); err != nil {
		return fmt.Errorf("write report markdown %q: %w", path, err)
	}
	return nil
}
