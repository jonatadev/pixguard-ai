package telemetry

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jonatadev/pixguard-ai/internal/config"
	"github.com/jonatadev/pixguard-ai/internal/domain"
)

// ScoreClient publishes Human-in-the-Loop analyst decisions as Scores to Langfuse.
type ScoreClient struct {
	BaseURL    string
	PublicKey  string
	SecretKey  string
	HTTPClient *http.Client
}

// NewScoreClient constructs a Langfuse ScoreClient from runtime configuration.
func NewScoreClient(cfg config.Config) ScoreClient {
	return ScoreClient{
		BaseURL:    strings.TrimRight(cfg.LangfuseBaseURL, "/"),
		PublicKey:  cfg.LangfusePublicKey,
		SecretKey:  cfg.LangfuseSecretKey,
		HTTPClient: &http.Client{Timeout: 8 * time.Second},
	}
}

// Configured reports whether Langfuse credentials and URL are available for score submission.
func (c ScoreClient) Configured() bool {
	return strings.TrimSpace(c.BaseURL) != "" &&
		strings.TrimSpace(c.PublicKey) != "" &&
		strings.TrimSpace(c.SecretKey) != ""
}

// PostAnalystDecision sends the analyst verdict (APROVADO=1.0, EDITADO=0.5, REJEITADO=0.0) to Langfuse.
func (c ScoreClient) PostAnalystDecision(ctx context.Context, traceID string, decision domain.AnalystDecision, notes string) (bool, error) {
	if !c.Configured() || strings.TrimSpace(traceID) == "" {
		return false, nil
	}

	var scoreVal float64
	switch decision {
	case domain.DecisionAprovado:
		scoreVal = 1.0
	case domain.DecisionEditado:
		scoreVal = 0.5
	case domain.DecisionRejeitado:
		scoreVal = 0.0
	default:
		return false, fmt.Errorf("unsupported analyst decision %q", decision)
	}

	payload := map[string]any{
		"traceId": traceID,
		"name":    "analyst_decision",
		"value":   scoreVal,
		"comment": strings.TrimSpace(notes),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("marshal score payload: %w", err)
	}

	endpoint := strings.TrimRight(c.BaseURL, "/") + "/api/public/scores"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return false, fmt.Errorf("create score request: %w", err)
	}

	creds := base64.StdEncoding.EncodeToString([]byte(c.PublicKey + ":" + c.SecretKey))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+creds)

	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 8 * time.Second}
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("send langfuse score: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return false, fmt.Errorf("langfuse score status %d: %s", resp.StatusCode, strings.TrimSpace(string(errBody)))
	}
	return true, nil
}
