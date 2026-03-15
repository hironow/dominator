package session

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/usecase/port"
)

// Compile-time check that K6Runner implements port.K6Runner.
var _ port.K6Runner = (*K6Runner)(nil)

// K6Runner executes k6 load test scripts via the k6 CLI.
type K6Runner struct {
	Logger domain.Logger
}

// Run executes a k6 load test with the given script and load configuration.
// k6 stdout is discarded (not mixed with dominator stdout).
// k6 stderr is forwarded to stderrW for transparent passthrough.
// Results are parsed from k6's --summary-export JSON file.
func (r *K6Runner) Run(ctx context.Context, scriptPath string, load domain.LoadConfig, stderrW io.Writer) (domain.K6Results, error) {
	tmpFile, err := os.CreateTemp("", "dominator-k6-summary-*.json")
	if err != nil {
		return domain.K6Results{}, fmt.Errorf("create temp file for k6 summary: %w", err)
	}
	summaryPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(summaryPath)

	// k6 is a trusted load testing tool invoked by the user.
	// Script path and load parameters come from an approved plan stored in .pass/.run/plans/.
	// nosemgrep: dangerous-exec-command [permanent]
	cmd := exec.CommandContext(ctx, "k6", "run",
		"--summary-export", summaryPath,
		"--vus", strconv.Itoa(load.VUs),
		"--duration", load.Duration,
		scriptPath,
	)
	cmd.Stdout = io.Discard
	cmd.Stderr = stderrW

	r.Logger.Info("Running k6: %s (vus=%d, duration=%s)", scriptPath, load.VUs, load.Duration)

	if err := cmd.Run(); err != nil {
		return domain.K6Results{}, fmt.Errorf("k6 run failed: %w", err)
	}

	data, err := os.ReadFile(summaryPath)
	if err != nil {
		return domain.K6Results{}, fmt.Errorf("read k6 summary: %w", err)
	}

	return ParseK6Summary(data)
}

// k6Summary represents the structure of k6's --summary-export JSON output.
type k6Summary struct {
	Metrics map[string]k6Metric `json:"metrics"`
}

type k6Metric struct {
	Values map[string]float64 `json:"values"`
}

// ParseK6Summary parses k6's summary export JSON into domain.K6Results.
func ParseK6Summary(data []byte) (domain.K6Results, error) {
	var summary k6Summary
	if err := json.Unmarshal(data, &summary); err != nil {
		return domain.K6Results{}, fmt.Errorf("parse k6 summary JSON: %w", err)
	}

	var results domain.K6Results

	if m, ok := summary.Metrics["http_req_duration"]; ok {
		if v, ok := m.Values["p(95)"]; ok {
			results.P95LatencyMs = v
		}
	}

	if m, ok := summary.Metrics["http_req_failed"]; ok {
		if v, ok := m.Values["rate"]; ok {
			results.ErrorRatePercent = v * 100
		}
	}

	if m, ok := summary.Metrics["http_reqs"]; ok {
		if v, ok := m.Values["rate"]; ok {
			results.ActualRPS = v
		}
	}

	// Success rate is derived from error rate
	results.SuccessRate = 100 - results.ErrorRatePercent

	return results, nil
}
