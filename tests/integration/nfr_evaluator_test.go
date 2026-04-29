//go:build integration

package integration_test

import (
	"testing"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/session"
)

// TestNfrEvaluation_AllMetrics tests NFR evaluation with all 4 dimensions.
func TestNfrEvaluation_AllMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testCases := []struct {
		name            string
		results         domain.K6Results
		nfr             domain.NfrConfig
		expectedVerdict domain.Verdict
		expectedDevs    int
	}{
		{
			name: "all_pass",
			results: domain.K6Results{
				P95LatencyMs:     100,
				ErrorRatePercent: 0.1,
				SuccessRate:      99.9,
				ActualRPS:        150,
			},
			nfr: domain.NfrConfig{
				Performance: domain.PerformanceNfr{P95LatencyMs: 500, ErrorRatePercent: 1.0},
				Reliability: domain.ReliabilityNfr{SuccessRatePercent: 99.0},
				Scalability: domain.ScalabilityNfr{TargetRPS: 100},
			},
			expectedVerdict: domain.VerdictPass,
			expectedDevs:    0,
		},
		{
			name: "latency_violation",
			results: domain.K6Results{
				P95LatencyMs:     800,
				ErrorRatePercent: 0.1,
				SuccessRate:      99.9,
				ActualRPS:        150,
			},
			nfr: domain.NfrConfig{
				Performance: domain.PerformanceNfr{P95LatencyMs: 500, ErrorRatePercent: 1.0},
				Reliability: domain.ReliabilityNfr{SuccessRatePercent: 99.0},
				Scalability: domain.ScalabilityNfr{TargetRPS: 100},
			},
			expectedVerdict: domain.VerdictViolation,
			expectedDevs:    1,
		},
		{
			name: "all_violation",
			results: domain.K6Results{
				P95LatencyMs:     1000,
				ErrorRatePercent: 5.0,
				SuccessRate:      90.0,
				ActualRPS:        50,
			},
			nfr: domain.NfrConfig{
				Performance: domain.PerformanceNfr{P95LatencyMs: 500, ErrorRatePercent: 1.0},
				Reliability: domain.ReliabilityNfr{SuccessRatePercent: 99.0},
				Scalability: domain.ScalabilityNfr{TargetRPS: 100},
			},
			expectedVerdict: domain.VerdictViolation,
			expectedDevs:    4,
		},
		{
			name: "rps_violation_only",
			results: domain.K6Results{
				P95LatencyMs:     100,
				ErrorRatePercent: 0.1,
				SuccessRate:      99.9,
				ActualRPS:        50,
			},
			nfr: domain.NfrConfig{
				Performance: domain.PerformanceNfr{P95LatencyMs: 500, ErrorRatePercent: 1.0},
				Reliability: domain.ReliabilityNfr{SuccessRatePercent: 99.0},
				Scalability: domain.ScalabilityNfr{TargetRPS: 100},
			},
			expectedVerdict: domain.VerdictViolation,
			expectedDevs:    1,
		},
	}

	evaluator := &session.NfrEvaluator{}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// when
			verdict, deviations := evaluator.Evaluate(tc.results, tc.nfr)

			// then
			if verdict != tc.expectedVerdict {
				t.Errorf("expected verdict %s, got %s", tc.expectedVerdict, verdict)
			}
			if len(deviations) != tc.expectedDevs {
				t.Errorf("expected %d deviations, got %d", tc.expectedDevs, len(deviations))
			}
		})
	}
}

// TestNfrEvaluation_EdgeCases tests boundary conditions.
func TestNfrEvaluation_EdgeCases(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testCases := []struct {
		name            string
		results         domain.K6Results
		nfr             domain.NfrConfig
		expectedVerdict domain.Verdict
		expectedDevs    int
	}{
		{
			name: "threshold_exactly_met",
			results: domain.K6Results{
				P95LatencyMs:     500,
				ErrorRatePercent: 1.0,
				SuccessRate:      99.0,
				ActualRPS:        100,
			},
			nfr: domain.NfrConfig{
				Performance: domain.PerformanceNfr{P95LatencyMs: 500, ErrorRatePercent: 1.0},
				Reliability: domain.ReliabilityNfr{SuccessRatePercent: 99.0},
				Scalability: domain.ScalabilityNfr{TargetRPS: 100},
			},
			expectedVerdict: domain.VerdictPass,
			expectedDevs:    0,
		},
		{
			name: "zero_thresholds_pass",
			results: domain.K6Results{
				P95LatencyMs:     100,
				ErrorRatePercent: 0.1,
				SuccessRate:      99.9,
				ActualRPS:        50,
			},
			nfr: domain.NfrConfig{
				Performance: domain.PerformanceNfr{P95LatencyMs: 0, ErrorRatePercent: 0},
				Reliability: domain.ReliabilityNfr{SuccessRatePercent: 0},
				Scalability: domain.ScalabilityNfr{TargetRPS: 0},
			},
			expectedVerdict: domain.VerdictPass,
			expectedDevs:    0,
		},
		{
			name: "zero_actual_values",
			results: domain.K6Results{
				P95LatencyMs:     0,
				ErrorRatePercent: 0,
				SuccessRate:      100.0,
				ActualRPS:        200,
			},
			nfr: domain.NfrConfig{
				Performance: domain.PerformanceNfr{P95LatencyMs: 500, ErrorRatePercent: 1.0},
				Reliability: domain.ReliabilityNfr{SuccessRatePercent: 99.0},
				Scalability: domain.ScalabilityNfr{TargetRPS: 100},
			},
			expectedVerdict: domain.VerdictPass,
			expectedDevs:    0,
		},
		{
			name: "barely_over_threshold",
			results: domain.K6Results{
				P95LatencyMs:     501,
				ErrorRatePercent: 0.1,
				SuccessRate:      99.9,
				ActualRPS:        150,
			},
			nfr: domain.NfrConfig{
				Performance: domain.PerformanceNfr{P95LatencyMs: 500, ErrorRatePercent: 1.0},
				Reliability: domain.ReliabilityNfr{SuccessRatePercent: 99.0},
				Scalability: domain.ScalabilityNfr{TargetRPS: 100},
			},
			expectedVerdict: domain.VerdictViolation,
			expectedDevs:    1,
		},
	}

	evaluator := &session.NfrEvaluator{}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// when
			verdict, deviations := evaluator.Evaluate(tc.results, tc.nfr)

			// then
			if verdict != tc.expectedVerdict {
				t.Errorf("expected verdict %s, got %s", tc.expectedVerdict, verdict)
			}
			if len(deviations) != tc.expectedDevs {
				t.Errorf("expected %d deviations, got %d", tc.expectedDevs, len(deviations))
			}
		})
	}
}

// TestNfrEvaluation_DeviationSeverity verifies correct severity classification.
func TestNfrEvaluation_DeviationSeverity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testCases := []struct {
		name             string
		deviationPercent float64
		expectedSeverity string
	}{
		{"low_5pct", 5.0, domain.SeverityLow},
		{"low_9pct", 9.0, domain.SeverityLow},
		{"medium_10pct", 10.0, domain.SeverityMedium},
		{"medium_30pct", 30.0, domain.SeverityMedium},
		{"medium_50pct", 50.0, domain.SeverityMedium},
		{"high_51pct", 51.0, domain.SeverityHigh},
		{"high_100pct", 100.0, domain.SeverityHigh},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// when
			severity := domain.CalcSeverity(tc.deviationPercent)

			// then
			if severity != tc.expectedSeverity {
				t.Errorf("expected severity %q for %.1f%%, got %q",
					tc.expectedSeverity, tc.deviationPercent, severity)
			}
		})
	}
}
