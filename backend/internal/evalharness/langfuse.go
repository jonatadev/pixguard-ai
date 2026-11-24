package evalharness

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// LangfuseClient synchronizes evaluation datasets, sanitized traces, experiment runs, and scores.
type LangfuseClient struct {
	BaseURL    string
	PublicKey  string
	SecretKey  string
	HTTPClient *http.Client
}

// SyncSummary reports how many items, traces, and scores were pushed to Langfuse.
type SyncSummary struct {
	DatasetName string
	RunName     string
	ItemsSynced int
	TracesSent  int
	ScoresSent  int
}

// SyncReport upserts the dataset and items, emits sanitized traces, links dataset-run-items, and records scores.
func (c LangfuseClient) SyncReport(ctx context.Context, ds Dataset, report Report, runName string) (SyncSummary, error) {
	if strings.TrimSpace(c.BaseURL) == "" || strings.TrimSpace(c.PublicKey) == "" || strings.TrimSpace(c.SecretKey) == "" {
		return SyncSummary{}, fmt.Errorf("langfuse base URL, public key, and secret key are required")
	}
	if strings.TrimSpace(runName) == "" {
		runName = "eval-" + time.Now().UTC().Format("20060102-150405")
	}

	if err := c.postJSON(ctx, "/api/public/v2/datasets", map[string]any{
		"name":        ds.Name,
		"description": ds.Description,
		"metadata": map[string]any{
			"version": ds.Version,
			"source":  "pixguard-eval-harness",
		},
	}); err != nil {
		return SyncSummary{}, fmt.Errorf("upsert dataset: %w", err)
	}

	caseMap := make(map[string]EvalCase, len(ds.Cases))
	for _, tc := range ds.Cases {
		caseMap[tc.ID] = tc
	}

	var summary SyncSummary
	summary.DatasetName = ds.Name
	summary.RunName = runName

	for _, res := range report.Results {
		tc := caseMap[res.ID]
		itemID := ds.Name + "-" + res.ID

		if err := c.postJSON(ctx, "/api/public/dataset-items", map[string]any{
			"id":          itemID,
			"datasetName": ds.Name,
			"input": map[string]any{
				"caseId":      tc.ID,
				"category":    tc.Category,
				"daysSinceTx": tc.DaysSinceTx,
				"text":        tc.Input,
			},
			"expectedOutput": map[string]any{
				"expectInjection":      tc.ExpectInjection,
				"expectSafeForLlm":     tc.ExpectSafeForLLM,
				"expectMedWindowValid": tc.ExpectMEDWindowValid,
				"requiredTokens":       tc.RequiredTokens,
			},
			"metadata": map[string]any{
				"name":                  tc.Name,
				"expectedMinRedactions": tc.ExpectedMinRedactions,
			},
		}); err != nil {
			return summary, fmt.Errorf("upsert dataset item %s: %w", itemID, err)
		}
		summary.ItemsSynced++

		traceID := deterministicTraceID(runName, res.ID)
		nowStr := time.Now().UTC().Format(time.RFC3339Nano)

		// Never publish unredacted PII in evaluation trace bodies; record sanitized text and counts only.
		if err := c.postJSON(ctx, "/api/public/ingestion", map[string]any{
			"batch": []map[string]any{
				{
					"id":        traceID + "-evt",
					"type":      "trace-create",
					"timestamp": nowStr,
					"body": map[string]any{
						"id":        traceID,
						"timestamp": nowStr,
						"name":      "eval:" + res.ID,
						"input": map[string]any{
							"caseId":        res.ID,
							"category":      res.Category,
							"sanitizedText": res.SanitizedText,
						},
						"output": map[string]any{
							"passed":             res.Passed,
							"piiRedactionsCount": res.PIIRedactionsCount,
							"piiLeaksCount":      res.PIILeaksCount,
							"injectionDetected":  res.InjectionDetected,
							"safeForLlm":         res.SafeForLLM,
							"medWindowValid":     res.MEDWindowValid,
							"failures":           res.Failures,
						},
						"metadata": map[string]any{
							"dataset":        ds.Name,
							"datasetVersion": ds.Version,
							"runName":        runName,
						},
						"tags": []string{"eval-harness", res.Category, runName},
					},
				},
			},
		}); err != nil {
			return summary, fmt.Errorf("ingest trace for %s: %w", res.ID, err)
		}
		summary.TracesSent++

		if err := c.postJSON(ctx, "/api/public/dataset-run-items", map[string]any{
			"runName":       runName,
			"datasetItemId": itemID,
			"traceId":       traceID,
		}); err != nil {
			return summary, fmt.Errorf("link dataset run item %s: %w", itemID, err)
		}

		scores := []struct {
			name  string
			value float64
		}{
			{name: "eval_passed", value: boolToFloat(res.Passed)},
			{name: "pii_zero_leakage", value: boolToFloat(res.PIILeaksCount == 0)},
			{name: "injection_accuracy", value: boolToFloat(res.InjectionMatched)},
		}
		for _, sc := range scores {
			if err := c.postJSON(ctx, "/api/public/scores", map[string]any{
				"traceId": traceID,
				"name":    sc.name,
				"value":   sc.value,
				"comment": fmt.Sprintf("dataset=%s run=%s case=%s", ds.Name, runName, res.ID),
			}); err != nil {
				return summary, fmt.Errorf("post score %s for %s: %w", sc.name, res.ID, err)
			}
			summary.ScoresSent++
		}
	}

	return summary, nil
}

func (c LangfuseClient) postJSON(ctx context.Context, path string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	endpoint := strings.TrimRight(c.BaseURL, "/") + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	creds := base64.StdEncoding.EncodeToString([]byte(c.PublicKey + ":" + c.SecretKey))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+creds)

	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("call %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("%s returned status %d: %s", path, resp.StatusCode, strings.TrimSpace(string(errBody)))
	}
	return nil
}

func deterministicTraceID(runName, caseID string) string {
	sum := sha256.Sum256([]byte(runName + ":" + caseID))
	return hex.EncodeToString(sum[:16])
}

func boolToFloat(v bool) float64 {
	if v {
		return 1.0
	}
	return 0.0
}
