package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jonatadev/pixguard-ai/internal/evalharness"
)

func main() {
	datasetPath := flag.String("dataset", "../evals/datasets/pixguard-golden-v1.json", "Path to evaluation dataset JSON")
	jsonOut := flag.String("json-out", "", "Optional path to write JSON evaluation report")
	markdownOut := flag.String("markdown-out", "", "Optional path to write Markdown evaluation summary")
	langfuseSync := flag.Bool("langfuse-sync", false, "Sync dataset, sanitized traces, experiment run, and scores to Langfuse")
	runName := flag.String("run-name", "", "Optional Langfuse experiment run name")
	flag.Parse()

	ds, err := evalharness.LoadDataset(*datasetPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading dataset: %v\n", err)
		os.Exit(1)
	}

	report := evalharness.Run(ds)
	for _, r := range report.Results {
		if r.Passed {
			fmt.Printf("[PASS] %s (%s): redactions=%d, injection=%v, med_window=%v\n",
				r.ID, r.Name, r.PIIRedactionsCount, r.InjectionDetected, r.MEDWindowValid)
		} else {
			fmt.Printf("[FAIL] %s (%s): %s\n", r.ID, r.Name, strings.Join(r.Failures, "; "))
		}
	}

	fmt.Printf("\nPixGuard AI Eval Harness (%s v%s): %d/%d passed (pii_leaks=%d, injection=%d/%d, med_window=%d/%d)\n",
		report.DatasetName,
		report.DatasetVersion,
		report.Summary.Passed,
		report.Summary.Total,
		report.Summary.PIILeaks,
		report.Summary.InjectionDetected,
		report.Summary.InjectionCases,
		report.Summary.MEDWindowMatches,
		report.Summary.Total,
	)

	if *jsonOut != "" {
		if err := evalharness.WriteJSON(*jsonOut, report); err != nil {
			fmt.Fprintf(os.Stderr, "error writing json report: %v\n", err)
			os.Exit(1)
		}
	}
	if *markdownOut != "" {
		if err := evalharness.WriteMarkdown(*markdownOut, report); err != nil {
			fmt.Fprintf(os.Stderr, "error writing markdown report: %v\n", err)
			os.Exit(1)
		}
	}

	if *langfuseSync {
		baseURL := strings.TrimSpace(os.Getenv("LANGFUSE_BASE_URL"))
		if baseURL == "" {
			baseURL = "http://localhost:3000"
		}
		client := evalharness.LangfuseClient{
			BaseURL:   baseURL,
			PublicKey: strings.TrimSpace(os.Getenv("LANGFUSE_PUBLIC_KEY")),
			SecretKey: strings.TrimSpace(os.Getenv("LANGFUSE_SECRET_KEY")),
		}
		name := *runName
		if name == "" {
			name = "eval-" + time.Now().UTC().Format("20060102-150405")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		syncRes, err := client.SyncReport(ctx, ds, report, name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error syncing eval run to Langfuse: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Langfuse sync complete: dataset=%s run=%s items=%d traces=%d scores=%d\n",
			syncRes.DatasetName, syncRes.RunName, syncRes.ItemsSynced, syncRes.TracesSent, syncRes.ScoresSent)
	}

	if report.Summary.Failed > 0 {
		os.Exit(1)
	}
}
