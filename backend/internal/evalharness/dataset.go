package evalharness

import (
	"encoding/json"
	"fmt"
	"os"
)

// Dataset represents a versioned collection of evaluation and red-team cases.
type Dataset struct {
	Name        string     `json:"name"`
	Version     string     `json:"version"`
	Description string     `json:"description"`
	Cases       []EvalCase `json:"cases"`
}

// EvalCase defines a single deterministic guardrail and business-invariant test case.
type EvalCase struct {
	ID                     string   `json:"id"`
	Category               string   `json:"category"`
	Name                   string   `json:"name"`
	Input                  string   `json:"input"`
	DaysSinceTx            int      `json:"days_since_tx"`
	ForbiddenRawSubstrings []string `json:"forbidden_raw_substrings"`
	RequiredTokens         []string `json:"required_tokens"`
	ExpectInjection        bool     `json:"expect_injection"`
	ExpectSafeForLLM       bool     `json:"expect_safe_for_llm"`
	ExpectMEDWindowValid   bool     `json:"expect_med_window_valid"`
	ExpectedMinRedactions  int      `json:"expected_min_redactions"`
}

// LoadDataset reads and validates a JSON evaluation dataset from disk.
func LoadDataset(path string) (Dataset, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Dataset{}, fmt.Errorf("read dataset %q: %w", path, err)
	}

	var ds Dataset
	if err := json.Unmarshal(raw, &ds); err != nil {
		return Dataset{}, fmt.Errorf("parse dataset %q: %w", path, err)
	}

	if ds.Name == "" {
		return Dataset{}, fmt.Errorf("dataset %q missing required field name", path)
	}
	if len(ds.Cases) == 0 {
		return Dataset{}, fmt.Errorf("dataset %q contains no eval cases", path)
	}

	seen := make(map[string]struct{}, len(ds.Cases))
	for i, tc := range ds.Cases {
		if tc.ID == "" || tc.Name == "" {
			return Dataset{}, fmt.Errorf("dataset %q case[%d] missing id or name", path, i)
		}
		if _, dup := seen[tc.ID]; dup {
			return Dataset{}, fmt.Errorf("dataset %q has duplicate case id %q", path, tc.ID)
		}
		seen[tc.ID] = struct{}{}
	}

	return ds, nil
}
