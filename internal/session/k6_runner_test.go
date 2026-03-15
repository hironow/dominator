package session_test

import (
	"math"
	"testing"

	"github.com/hironow/dominator/internal/session"
)

func TestParseK6Summary_ValidJSON(t *testing.T) {
	// given
	data := []byte(`{
		"metrics": {
			"http_req_duration": { "values": { "p(95)": 123.45, "avg": 50.0 } },
			"http_req_failed": { "values": { "rate": 0.01 } },
			"http_reqs": { "values": { "rate": 500.0, "count": 15000 } }
		}
	}`)

	// when
	results, err := session.ParseK6Summary(data)

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(results.P95LatencyMs-123.45) > 0.001 {
		t.Errorf("P95LatencyMs = %f, want 123.45", results.P95LatencyMs)
	}
	if math.Abs(results.ErrorRatePercent-1.0) > 0.001 {
		t.Errorf("ErrorRatePercent = %f, want 1.0", results.ErrorRatePercent)
	}
	if math.Abs(results.SuccessRate-99.0) > 0.001 {
		t.Errorf("SuccessRate = %f, want 99.0", results.SuccessRate)
	}
	if math.Abs(results.ActualRPS-500.0) > 0.001 {
		t.Errorf("ActualRPS = %f, want 500.0", results.ActualRPS)
	}
}

func TestParseK6Summary_MissingMetrics(t *testing.T) {
	// given: JSON with no recognized metrics
	data := []byte(`{
		"metrics": {
			"custom_metric": { "values": { "avg": 42.0 } }
		}
	}`)

	// when
	results, err := session.ParseK6Summary(data)

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// All should be zero defaults
	if results.P95LatencyMs != 0 {
		t.Errorf("P95LatencyMs = %f, want 0", results.P95LatencyMs)
	}
	if results.ErrorRatePercent != 0 {
		t.Errorf("ErrorRatePercent = %f, want 0", results.ErrorRatePercent)
	}
	if results.SuccessRate != 0 {
		t.Errorf("SuccessRate = %f, want 0 (unmeasured when http_req_failed absent)", results.SuccessRate)
	}
	if results.ActualRPS != 0 {
		t.Errorf("ActualRPS = %f, want 0", results.ActualRPS)
	}
}

func TestParseK6Summary_InvalidJSON(t *testing.T) {
	// given
	data := []byte(`not valid json`)

	// when
	_, err := session.ParseK6Summary(data)

	// then
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestParseK6Summary_EmptyMetrics(t *testing.T) {
	// given
	data := []byte(`{"metrics": {}}`)

	// when
	results, err := session.ParseK6Summary(data)

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results.SuccessRate != 0 {
		t.Errorf("SuccessRate = %f, want 0 (unmeasured — no http_req_failed metric)", results.SuccessRate)
	}
}
