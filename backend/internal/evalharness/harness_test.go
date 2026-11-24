package evalharness

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestLoadDatasetAndRunGoldenSuite(t *testing.T) {
	datasetPath := filepath.Join("..", "..", "..", "evals", "datasets", "pixguard-golden-v1.json")
	ds, err := LoadDataset(datasetPath)
	if err != nil {
		t.Fatalf("LoadDataset failed: %v", err)
	}
	if ds.Name != "pixguard-golden-v1" || len(ds.Cases) != 14 {
		t.Fatalf("unexpected dataset metadata: name=%s cases=%d", ds.Name, len(ds.Cases))
	}

	report := Run(ds)
	if report.Summary.Total != 14 || report.Summary.Failed != 0 || report.Summary.Passed != 14 {
		t.Fatalf("expected 14/14 pass, got passed=%d failed=%d", report.Summary.Passed, report.Summary.Failed)
	}
	if report.Summary.PIILeaks != 0 {
		t.Fatalf("expected 0 PII leaks, got %d", report.Summary.PIILeaks)
	}
	if report.Summary.InjectionCases != 5 || report.Summary.InjectionDetected != 5 || report.Summary.FalsePositives != 0 {
		t.Fatalf("unexpected injection summary: %+v", report.Summary)
	}
	if report.Summary.MEDWindowMatches != 14 {
		t.Fatalf("expected 14 MED window matches, got %d", report.Summary.MEDWindowMatches)
	}

	md := RenderMarkdown(report)
	if !strings.Contains(md, "PixGuard AI — Eval & Red-Team Harness (PASS)") || !strings.Contains(md, "LGPD-01") {
		t.Fatalf("unexpected markdown output: %s", md)
	}

	tmpDir := t.TempDir()
	jsonPath := filepath.Join(tmpDir, "out", "eval.json")
	mdPath := filepath.Join(tmpDir, "out", "eval.md")
	if err := WriteJSON(jsonPath, report); err != nil {
		t.Fatalf("WriteJSON failed: %v", err)
	}
	if err := WriteMarkdown(mdPath, report); err != nil {
		t.Fatalf("WriteMarkdown failed: %v", err)
	}
	if _, err := os.Stat(jsonPath); err != nil {
		t.Fatalf("missing json output: %v", err)
	}
	if _, err := os.Stat(mdPath); err != nil {
		t.Fatalf("missing markdown output: %v", err)
	}
}

func TestLoadDatasetValidationErrors(t *testing.T) {
	if _, err := LoadDataset(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Fatalf("expected error for missing file")
	}

	dupPath := filepath.Join(t.TempDir(), "dup.json")
	dupJSON := `{"name":"dup","cases":[{"id":"A","name":"1"},{"id":"A","name":"2"}]}`
	if err := os.WriteFile(dupPath, []byte(dupJSON), 0o644); err != nil {
		t.Fatalf("write temp dataset: %v", err)
	}
	if _, err := LoadDataset(dupPath); err == nil {
		t.Fatalf("expected error for duplicate case id")
	}
}

func TestRunDetectsFailures(t *testing.T) {
	ds := Dataset{
		Name:    "failure-check",
		Version: "1.0.0",
		Cases: []EvalCase{
			{
				ID:                     "FAIL-01",
				Category:               "med_business_rule",
				Name:                   "Expect invalid MED window failure",
				Input:                  "Contestacao de 95 dias",
				DaysSinceTx:            95,
				ExpectInjection:        true,
				ExpectSafeForLLM:       true,
				ExpectMEDWindowValid:   true,
				ExpectedMinRedactions:  1,
				ForbiddenRawSubstrings: []string{"95 dias"},
				RequiredTokens:         []string{"[CPF_REDACTED]"},
			},
		},
	}

	report := Run(ds)
	if report.Summary.Failed != 1 || report.Results[0].Passed {
		t.Fatalf("expected case FAIL-01 to fail, got %+v", report.Results[0])
	}
	if report.Results[0].PIILeaksCount != 1 {
		t.Fatalf("expected 1 leak match, got %d", report.Results[0].PIILeaksCount)
	}
	if !strings.Contains(RenderMarkdown(report), "FAIL") {
		t.Fatalf("expected FAIL badge in markdown")
	}
}

func TestLangfuseSyncReport(t *testing.T) {
	var mu sync.Mutex
	pathCounts := make(map[string]int)
	var authHeader string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		pathCounts[r.URL.Path]++
		authHeader = r.Header.Get("Authorization")
		mu.Unlock()

		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	ds := Dataset{
		Name:        "unit-ds",
		Version:     "1.0.0",
		Description: "Unit test dataset",
		Cases: []EvalCase{
			{
				ID:                   "U-01",
				Category:             "lgpd",
				Name:                 "Clean text",
				Input:                "Pix contestado",
				DaysSinceTx:          1,
				ExpectSafeForLLM:     true,
				ExpectMEDWindowValid: true,
			},
		},
	}
	report := Run(ds)

	client := LangfuseClient{
		BaseURL:    srv.URL,
		PublicKey:  "pk-test",
		SecretKey:  "sk-test",
		HTTPClient: srv.Client(),
	}

	summary, err := client.SyncReport(context.Background(), ds, report, "ci-run-1")
	if err != nil {
		t.Fatalf("SyncReport failed: %v", err)
	}
	if summary.ItemsSynced != 1 || summary.TracesSent != 1 || summary.ScoresSent != 3 {
		t.Fatalf("unexpected sync summary: %+v", summary)
	}
	if !strings.HasPrefix(authHeader, "Basic ") {
		t.Fatalf("expected Basic auth header, got %q", authHeader)
	}

	expectedPaths := []string{
		"/api/public/v2/datasets",
		"/api/public/dataset-items",
		"/api/public/ingestion",
		"/api/public/dataset-run-items",
		"/api/public/scores",
	}
	for _, p := range expectedPaths {
		if pathCounts[p] == 0 {
			t.Fatalf("expected endpoint %s to be called", p)
		}
	}
}
