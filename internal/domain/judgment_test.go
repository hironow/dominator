package domain_test

import (
	"testing"

	"github.com/hironow/dominator/internal/domain"
)

func TestCalcSeverity_Low(t *testing.T) {
	// given
	deviationPercent := 5.0

	// when
	result := domain.CalcSeverity(deviationPercent)

	// then
	if result != domain.SeverityLow {
		t.Errorf("expected %q, got %q", domain.SeverityLow, result)
	}
}

func TestCalcSeverity_Medium(t *testing.T) {
	// given
	deviationPercent := 25.0

	// when
	result := domain.CalcSeverity(deviationPercent)

	// then
	if result != domain.SeverityMedium {
		t.Errorf("expected %q, got %q", domain.SeverityMedium, result)
	}
}

func TestCalcSeverity_High(t *testing.T) {
	// given
	deviationPercent := 75.0

	// when
	result := domain.CalcSeverity(deviationPercent)

	// then
	if result != domain.SeverityHigh {
		t.Errorf("expected %q, got %q", domain.SeverityHigh, result)
	}
}

func TestCalcSeverity_Boundary(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{name: "10_percent_is_medium", input: 10.0, expected: domain.SeverityMedium},
		{name: "50_percent_is_medium", input: 50.0, expected: domain.SeverityMedium},
		{name: "50.1_percent_is_high", input: 50.1, expected: domain.SeverityHigh},
		{name: "9.9_percent_is_low", input: 9.9, expected: domain.SeverityLow},
		{name: "negative_uses_absolute", input: -25.0, expected: domain.SeverityMedium},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// when
			result := domain.CalcSeverity(tt.input)

			// then
			if result != tt.expected {
				t.Errorf("CalcSeverity(%f): expected %q, got %q", tt.input, tt.expected, result)
			}
		})
	}
}

func TestCalcDeviation_Basic(t *testing.T) {
	// given
	actual := 300.0
	threshold := 200.0

	// when
	result := domain.CalcDeviation(actual, threshold)

	// then
	expected := 50.0
	if result != expected {
		t.Errorf("expected %f, got %f", expected, result)
	}
}

func TestCalcDeviation_ZeroThreshold(t *testing.T) {
	// given
	actual := 300.0
	threshold := 0.0

	// when
	result := domain.CalcDeviation(actual, threshold)

	// then
	if result != 0 {
		t.Errorf("expected 0 for zero threshold, got %f", result)
	}
}

func TestEvaluateNfr_AllPass(t *testing.T) {
	// given
	results := domain.K6Results{
		P95LatencyMs:     100,
		ErrorRatePercent: 0.5,
		SuccessRate:      99.9,
		ActualRPS:        150,
	}
	nfr := domain.NfrConfig{
		Performance: domain.PerformanceNfr{
			P95LatencyMs:     200,
			ErrorRatePercent: 1.0,
		},
		Reliability: domain.ReliabilityNfr{
			SuccessRatePercent: 99.0,
		},
		Scalability: domain.ScalabilityNfr{
			TargetRPS: 100,
		},
	}

	// when
	verdict, deviations := domain.EvaluateNfr(results, nfr)

	// then
	if verdict != domain.VerdictPass {
		t.Errorf("expected pass, got %s", verdict)
	}
	if len(deviations) != 0 {
		t.Errorf("expected no deviations, got %d", len(deviations))
	}
}

func TestEvaluateNfr_LatencyViolation(t *testing.T) {
	// given
	results := domain.K6Results{
		P95LatencyMs:     300,
		ErrorRatePercent: 0.5,
		SuccessRate:      99.9,
		ActualRPS:        150,
	}
	nfr := domain.NfrConfig{
		Performance: domain.PerformanceNfr{
			P95LatencyMs:     200,
			ErrorRatePercent: 1.0,
		},
		Reliability: domain.ReliabilityNfr{
			SuccessRatePercent: 99.0,
		},
		Scalability: domain.ScalabilityNfr{
			TargetRPS: 100,
		},
	}

	// when
	verdict, deviations := domain.EvaluateNfr(results, nfr)

	// then
	if verdict != domain.VerdictViolation {
		t.Errorf("expected violation, got %s", verdict)
	}
	if len(deviations) != 1 {
		t.Fatalf("expected 1 deviation, got %d", len(deviations))
	}
	if deviations[0].Metric != "p95_latency_ms" {
		t.Errorf("expected metric p95_latency_ms, got %s", deviations[0].Metric)
	}
	if deviations[0].Deviation != 50.0 {
		t.Errorf("expected 50%% deviation, got %f", deviations[0].Deviation)
	}
}

func TestEvaluateNfr_MultipleViolations(t *testing.T) {
	// given
	results := domain.K6Results{
		P95LatencyMs:     300,
		ErrorRatePercent: 3.0,
		SuccessRate:      99.9,
		ActualRPS:        150,
	}
	nfr := domain.NfrConfig{
		Performance: domain.PerformanceNfr{
			P95LatencyMs:     200,
			ErrorRatePercent: 1.0,
		},
		Reliability: domain.ReliabilityNfr{
			SuccessRatePercent: 99.0,
		},
		Scalability: domain.ScalabilityNfr{
			TargetRPS: 100,
		},
	}

	// when
	verdict, deviations := domain.EvaluateNfr(results, nfr)

	// then
	if verdict != domain.VerdictViolation {
		t.Errorf("expected violation, got %s", verdict)
	}
	if len(deviations) != 2 {
		t.Fatalf("expected 2 deviations (latency + error rate), got %d", len(deviations))
	}

	metrics := make(map[string]bool)
	for _, d := range deviations {
		metrics[d.Metric] = true
	}
	if !metrics["p95_latency_ms"] {
		t.Error("expected p95_latency_ms deviation")
	}
	if !metrics["error_rate_percent"] {
		t.Error("expected error_rate_percent deviation")
	}
}

func TestEvaluateNfr_SuccessRateViolation(t *testing.T) {
	// given
	results := domain.K6Results{
		P95LatencyMs:     100,
		ErrorRatePercent: 0.5,
		SuccessRate:      98.0,
		ActualRPS:        150,
	}
	nfr := domain.NfrConfig{
		Performance: domain.PerformanceNfr{
			P95LatencyMs:     200,
			ErrorRatePercent: 1.0,
		},
		Reliability: domain.ReliabilityNfr{
			SuccessRatePercent: 99.9,
		},
		Scalability: domain.ScalabilityNfr{
			TargetRPS: 100,
		},
	}

	// when
	verdict, deviations := domain.EvaluateNfr(results, nfr)

	// then
	if verdict != domain.VerdictViolation {
		t.Errorf("expected violation, got %s", verdict)
	}
	if len(deviations) != 1 {
		t.Fatalf("expected 1 deviation, got %d", len(deviations))
	}
	if deviations[0].Metric != "success_rate_percent" {
		t.Errorf("expected metric success_rate_percent, got %s", deviations[0].Metric)
	}
	if deviations[0].Actual != 98.0 {
		t.Errorf("expected actual 98.0, got %f", deviations[0].Actual)
	}
}
