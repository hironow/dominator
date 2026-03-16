package session

import (
	"encoding/json"
	"fmt"

	"github.com/hironow/dominator/internal/domain"
)

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
	errorRateFound := false

	if m, ok := summary.Metrics["http_req_duration"]; ok {
		if v, ok := m.Values["p(95)"]; ok {
			results.P95LatencyMs = v
		}
	}

	if m, ok := summary.Metrics["http_req_failed"]; ok {
		if v, ok := m.Values["rate"]; ok {
			results.ErrorRatePercent = v * 100
			errorRateFound = true
		}
	}

	if m, ok := summary.Metrics["http_reqs"]; ok {
		if v, ok := m.Values["rate"]; ok {
			results.ActualRPS = v
		}
	}

	// Success rate derived from error rate only when error rate was measured.
	// For non-HTTP protocols (ws-json-rpc) where http_req_failed is absent,
	// SuccessRate stays 0 (unmeasured), and NFR evaluation skips it
	// unless the threshold is explicitly set.
	if errorRateFound {
		results.SuccessRate = 100 - results.ErrorRatePercent
	}

	return results, nil
}
