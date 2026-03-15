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

func TestEvaluateNfr_AllCombinations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		results        domain.K6Results
		nfr            domain.NfrConfig
		wantVerdict    domain.Verdict
		wantMetrics    []string
		wantDevCount   int
	}{
		{
			name: "only_p95_latency_violation",
			results: domain.K6Results{
				P95LatencyMs:     600,
				ErrorRatePercent: 0.5,
				SuccessRate:      99.5,
				ActualRPS:        200,
			},
			nfr: domain.NfrConfig{
				Performance: domain.PerformanceNfr{P95LatencyMs: 500, ErrorRatePercent: 1.0},
				Reliability: domain.ReliabilityNfr{SuccessRatePercent: 99.0},
				Scalability: domain.ScalabilityNfr{TargetRPS: 100},
			},
			wantVerdict:  domain.VerdictViolation,
			wantMetrics:  []string{"p95_latency_ms"},
			wantDevCount: 1,
		},
		{
			name: "only_error_rate_violation",
			results: domain.K6Results{
				P95LatencyMs:     100,
				ErrorRatePercent: 5.0,
				SuccessRate:      99.5,
				ActualRPS:        200,
			},
			nfr: domain.NfrConfig{
				Performance: domain.PerformanceNfr{P95LatencyMs: 500, ErrorRatePercent: 1.0},
				Reliability: domain.ReliabilityNfr{SuccessRatePercent: 99.0},
				Scalability: domain.ScalabilityNfr{TargetRPS: 100},
			},
			wantVerdict:  domain.VerdictViolation,
			wantMetrics:  []string{"error_rate_percent"},
			wantDevCount: 1,
		},
		{
			name: "only_success_rate_violation",
			results: domain.K6Results{
				P95LatencyMs:     100,
				ErrorRatePercent: 0.5,
				SuccessRate:      95.0,
				ActualRPS:        200,
			},
			nfr: domain.NfrConfig{
				Performance: domain.PerformanceNfr{P95LatencyMs: 500, ErrorRatePercent: 1.0},
				Reliability: domain.ReliabilityNfr{SuccessRatePercent: 99.0},
				Scalability: domain.ScalabilityNfr{TargetRPS: 100},
			},
			wantVerdict:  domain.VerdictViolation,
			wantMetrics:  []string{"success_rate_percent"},
			wantDevCount: 1,
		},
		{
			name: "only_target_rps_violation",
			results: domain.K6Results{
				P95LatencyMs:     100,
				ErrorRatePercent: 0.5,
				SuccessRate:      99.5,
				ActualRPS:        50,
			},
			nfr: domain.NfrConfig{
				Performance: domain.PerformanceNfr{P95LatencyMs: 500, ErrorRatePercent: 1.0},
				Reliability: domain.ReliabilityNfr{SuccessRatePercent: 99.0},
				Scalability: domain.ScalabilityNfr{TargetRPS: 100},
			},
			wantVerdict:  domain.VerdictViolation,
			wantMetrics:  []string{"target_rps"},
			wantDevCount: 1,
		},
		{
			name: "all_four_violations",
			results: domain.K6Results{
				P95LatencyMs:     1000,
				ErrorRatePercent: 5.0,
				SuccessRate:      90.0,
				ActualRPS:        30,
			},
			nfr: domain.NfrConfig{
				Performance: domain.PerformanceNfr{P95LatencyMs: 500, ErrorRatePercent: 1.0},
				Reliability: domain.ReliabilityNfr{SuccessRatePercent: 99.0},
				Scalability: domain.ScalabilityNfr{TargetRPS: 100},
			},
			wantVerdict:  domain.VerdictViolation,
			wantMetrics:  []string{"p95_latency_ms", "error_rate_percent", "success_rate_percent", "target_rps"},
			wantDevCount: 4,
		},
		{
			name: "all_pass_exact_thresholds",
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
			wantVerdict:  domain.VerdictPass,
			wantDevCount: 0,
		},
		{
			name: "disabled_thresholds_pass",
			results: domain.K6Results{
				P95LatencyMs:     9999,
				ErrorRatePercent: 99.0,
				SuccessRate:      1.0,
				ActualRPS:        0,
			},
			nfr:          domain.NfrConfig{},
			wantVerdict:  domain.VerdictPass,
			wantDevCount: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// when
			verdict, deviations := domain.EvaluateNfr(tt.results, tt.nfr)

			// then
			if verdict != tt.wantVerdict {
				t.Errorf("verdict = %s, want %s", verdict, tt.wantVerdict)
			}
			if len(deviations) != tt.wantDevCount {
				t.Fatalf("deviations count = %d, want %d", len(deviations), tt.wantDevCount)
			}
			metrics := make(map[string]bool)
			for _, d := range deviations {
				metrics[d.Metric] = true
			}
			for _, m := range tt.wantMetrics {
				if !metrics[m] {
					t.Errorf("expected metric %q in deviations", m)
				}
			}
		})
	}
}

func TestEvaluateNfr_ZeroActual(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		results     domain.K6Results
		nfr         domain.NfrConfig
		wantVerdict domain.Verdict
		wantMetric  string
	}{
		{
			name:    "zero_latency_passes",
			results: domain.K6Results{P95LatencyMs: 0},
			nfr: domain.NfrConfig{
				Performance: domain.PerformanceNfr{P95LatencyMs: 500},
			},
			wantVerdict: domain.VerdictPass,
		},
		{
			name:    "zero_error_rate_passes",
			results: domain.K6Results{ErrorRatePercent: 0},
			nfr: domain.NfrConfig{
				Performance: domain.PerformanceNfr{ErrorRatePercent: 1.0},
			},
			wantVerdict: domain.VerdictPass,
		},
		{
			name:    "zero_success_rate_violates",
			results: domain.K6Results{SuccessRate: 0},
			nfr: domain.NfrConfig{
				Reliability: domain.ReliabilityNfr{SuccessRatePercent: 99.0},
			},
			wantVerdict: domain.VerdictViolation,
			wantMetric:  "success_rate_percent",
		},
		{
			name:    "zero_rps_violates",
			results: domain.K6Results{ActualRPS: 0},
			nfr: domain.NfrConfig{
				Scalability: domain.ScalabilityNfr{TargetRPS: 100},
			},
			wantVerdict: domain.VerdictViolation,
			wantMetric:  "target_rps",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verdict, deviations := domain.EvaluateNfr(tt.results, tt.nfr)
			if verdict != tt.wantVerdict {
				t.Errorf("verdict = %s, want %s", verdict, tt.wantVerdict)
			}
			if tt.wantMetric != "" && len(deviations) > 0 {
				if deviations[0].Metric != tt.wantMetric {
					t.Errorf("metric = %s, want %s", deviations[0].Metric, tt.wantMetric)
				}
			}
		})
	}
}

func TestEvaluateNfr_ThresholdExactlyMet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		results domain.K6Results
		nfr     domain.NfrConfig
	}{
		{
			name:    "latency_exactly_at_threshold",
			results: domain.K6Results{P95LatencyMs: 500},
			nfr:     domain.NfrConfig{Performance: domain.PerformanceNfr{P95LatencyMs: 500}},
		},
		{
			name:    "error_rate_exactly_at_threshold",
			results: domain.K6Results{ErrorRatePercent: 1.0},
			nfr:     domain.NfrConfig{Performance: domain.PerformanceNfr{ErrorRatePercent: 1.0}},
		},
		{
			name:    "success_rate_exactly_at_threshold",
			results: domain.K6Results{SuccessRate: 99.0},
			nfr:     domain.NfrConfig{Reliability: domain.ReliabilityNfr{SuccessRatePercent: 99.0}},
		},
		{
			name:    "rps_exactly_at_threshold",
			results: domain.K6Results{ActualRPS: 100},
			nfr:     domain.NfrConfig{Scalability: domain.ScalabilityNfr{TargetRPS: 100}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verdict, deviations := domain.EvaluateNfr(tt.results, tt.nfr)
			if verdict != domain.VerdictPass {
				t.Errorf("verdict = %s, want pass (threshold exactly met)", verdict)
			}
			if len(deviations) != 0 {
				t.Errorf("expected 0 deviations, got %d", len(deviations))
			}
		})
	}
}

func TestEvaluateNfr_BarelyOverThreshold(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		results    domain.K6Results
		nfr        domain.NfrConfig
		wantMetric string
	}{
		{
			name:       "latency_barely_over",
			results:    domain.K6Results{P95LatencyMs: 500.01},
			nfr:        domain.NfrConfig{Performance: domain.PerformanceNfr{P95LatencyMs: 500}},
			wantMetric: "p95_latency_ms",
		},
		{
			name:       "error_rate_barely_over",
			results:    domain.K6Results{ErrorRatePercent: 1.01},
			nfr:        domain.NfrConfig{Performance: domain.PerformanceNfr{ErrorRatePercent: 1.0}},
			wantMetric: "error_rate_percent",
		},
		{
			name:       "success_rate_barely_under",
			results:    domain.K6Results{SuccessRate: 98.99},
			nfr:        domain.NfrConfig{Reliability: domain.ReliabilityNfr{SuccessRatePercent: 99.0}},
			wantMetric: "success_rate_percent",
		},
		{
			name:       "rps_barely_under",
			results:    domain.K6Results{ActualRPS: 99.99},
			nfr:        domain.NfrConfig{Scalability: domain.ScalabilityNfr{TargetRPS: 100}},
			wantMetric: "target_rps",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verdict, deviations := domain.EvaluateNfr(tt.results, tt.nfr)
			if verdict != domain.VerdictViolation {
				t.Errorf("verdict = %s, want violation (barely over threshold)", verdict)
			}
			if len(deviations) != 1 {
				t.Fatalf("expected 1 deviation, got %d", len(deviations))
			}
			if deviations[0].Metric != tt.wantMetric {
				t.Errorf("metric = %s, want %s", deviations[0].Metric, tt.wantMetric)
			}
		})
	}
}

func TestCalcUnderDeviation_Parameterized(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		actual    float64
		threshold float64
		expected  float64
	}{
		{name: "50_percent_under", actual: 50, threshold: 100, expected: 50.0},
		{name: "25_percent_under", actual: 75, threshold: 100, expected: 25.0},
		{name: "100_percent_under", actual: 0, threshold: 100, expected: 100.0},
		{name: "10_percent_under", actual: 90, threshold: 100, expected: 10.0},
		{name: "zero_threshold", actual: 50, threshold: 0, expected: 0},
		{name: "exact_match", actual: 100, threshold: 100, expected: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domain.CalcUnderDeviation(tt.actual, tt.threshold)
			if got != tt.expected {
				t.Errorf("CalcUnderDeviation(%f, %f) = %f, want %f", tt.actual, tt.threshold, got, tt.expected)
			}
		})
	}
}

func TestCalcUnderDeviation_ActualExceedsThreshold(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		actual    float64
		threshold float64
	}{
		{name: "slightly_over", actual: 100.1, threshold: 100},
		{name: "double", actual: 200, threshold: 100},
		{name: "much_over", actual: 1000, threshold: 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domain.CalcUnderDeviation(tt.actual, tt.threshold)
			if got != 0 {
				t.Errorf("CalcUnderDeviation(%f, %f) = %f, want 0 (no violation)", tt.actual, tt.threshold, got)
			}
		})
	}
}

func TestCalcDeviation_Parameterized(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		actual    float64
		threshold float64
		expected  float64
	}{
		{name: "50_percent_over", actual: 300, threshold: 200, expected: 50.0},
		{name: "100_percent_over", actual: 400, threshold: 200, expected: 100.0},
		{name: "equal", actual: 200, threshold: 200, expected: 0},
		{name: "under_threshold", actual: 100, threshold: 200, expected: -50.0},
		{name: "zero_threshold", actual: 100, threshold: 0, expected: 0},
		{name: "zero_actual", actual: 0, threshold: 200, expected: -100.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domain.CalcDeviation(tt.actual, tt.threshold)
			if got != tt.expected {
				t.Errorf("CalcDeviation(%f, %f) = %f, want %f", tt.actual, tt.threshold, got, tt.expected)
			}
		})
	}
}

func TestCalcDeviation_NegativeValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		actual    float64
		threshold float64
	}{
		{name: "negative_actual", actual: -100, threshold: 200},
		{name: "negative_threshold", actual: 100, threshold: -200},
		{name: "both_negative", actual: -100, threshold: -200},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// should not panic
			got := domain.CalcDeviation(tt.actual, tt.threshold)
			_ = got // just verify no panic
		})
	}
}

func TestCalcSeverity_Parameterized(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{name: "zero", input: 0, expected: domain.SeverityLow},
		{name: "1_percent", input: 1, expected: domain.SeverityLow},
		{name: "5_percent", input: 5, expected: domain.SeverityLow},
		{name: "9.99_percent", input: 9.99, expected: domain.SeverityLow},
		{name: "10_percent", input: 10, expected: domain.SeverityMedium},
		{name: "25_percent", input: 25, expected: domain.SeverityMedium},
		{name: "49.99_percent", input: 49.99, expected: domain.SeverityMedium},
		{name: "50_percent", input: 50, expected: domain.SeverityMedium},
		{name: "50.01_percent", input: 50.01, expected: domain.SeverityHigh},
		{name: "75_percent", input: 75, expected: domain.SeverityHigh},
		{name: "100_percent", input: 100, expected: domain.SeverityHigh},
		{name: "200_percent", input: 200, expected: domain.SeverityHigh},
		{name: "negative_5", input: -5, expected: domain.SeverityLow},
		{name: "negative_25", input: -25, expected: domain.SeverityMedium},
		{name: "negative_75", input: -75, expected: domain.SeverityHigh},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domain.CalcSeverity(tt.input)
			if got != tt.expected {
				t.Errorf("CalcSeverity(%f) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestEvaluateNfr_DeviationSeverity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		latency      float64
		threshold    int
		wantSeverity string
	}{
		{name: "low_severity_5pct", latency: 525, threshold: 500, wantSeverity: domain.SeverityLow},
		{name: "medium_severity_20pct", latency: 600, threshold: 500, wantSeverity: domain.SeverityMedium},
		{name: "high_severity_80pct", latency: 900, threshold: 500, wantSeverity: domain.SeverityHigh},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := domain.K6Results{P95LatencyMs: tt.latency}
			nfr := domain.NfrConfig{
				Performance: domain.PerformanceNfr{P95LatencyMs: tt.threshold},
			}
			_, deviations := domain.EvaluateNfr(results, nfr)
			if len(deviations) != 1 {
				t.Fatalf("expected 1 deviation, got %d", len(deviations))
			}
			if deviations[0].Severity != tt.wantSeverity {
				t.Errorf("severity = %s, want %s", deviations[0].Severity, tt.wantSeverity)
			}
		})
	}
}

func TestEvaluateNfr_ErrorRateViolation(t *testing.T) {
	// given
	results := domain.K6Results{
		P95LatencyMs:     100,
		ErrorRatePercent: 5.0,
		SuccessRate:      99.9,
		ActualRPS:        200,
	}
	nfr := domain.NfrConfig{
		Performance: domain.PerformanceNfr{P95LatencyMs: 200, ErrorRatePercent: 1.0},
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
	if deviations[0].Metric != "error_rate_percent" {
		t.Errorf("metric = %s, want error_rate_percent", deviations[0].Metric)
	}
}

func TestEvaluateNfr_ThreeViolations(t *testing.T) {
	// given
	results := domain.K6Results{
		P95LatencyMs:     800,
		ErrorRatePercent: 3.0,
		SuccessRate:      95.0,
		ActualRPS:        200,
	}
	nfr := domain.NfrConfig{
		Performance: domain.PerformanceNfr{P95LatencyMs: 500, ErrorRatePercent: 1.0},
		Reliability: domain.ReliabilityNfr{SuccessRatePercent: 99.0},
		Scalability: domain.ScalabilityNfr{TargetRPS: 100},
	}

	// when
	verdict, deviations := domain.EvaluateNfr(results, nfr)

	// then
	if verdict != domain.VerdictViolation {
		t.Errorf("expected violation, got %s", verdict)
	}
	if len(deviations) != 3 {
		t.Fatalf("expected 3 deviations, got %d", len(deviations))
	}
}

func TestCalcUnderDeviation_ZeroThreshold(t *testing.T) {
	got := domain.CalcUnderDeviation(50, 0)
	if got != 0 {
		t.Errorf("CalcUnderDeviation(50, 0) = %f, want 0", got)
	}
}

func TestCalcDeviation_BothZero(t *testing.T) {
	got := domain.CalcDeviation(0, 0)
	if got != 0 {
		t.Errorf("CalcDeviation(0, 0) = %f, want 0", got)
	}
}

func TestNfrDeviation_Fields(t *testing.T) {
	d := domain.NfrDeviation{
		Metric:    "p95_latency_ms",
		Threshold: 500,
		Actual:    750,
		Deviation: 50.0,
		Severity:  domain.SeverityMedium,
	}
	if d.Metric != "p95_latency_ms" {
		t.Errorf("Metric = %q", d.Metric)
	}
	if d.Threshold != 500 {
		t.Errorf("Threshold = %f", d.Threshold)
	}
	if d.Actual != 750 {
		t.Errorf("Actual = %f", d.Actual)
	}
	if d.Deviation != 50.0 {
		t.Errorf("Deviation = %f", d.Deviation)
	}
	if d.Severity != domain.SeverityMedium {
		t.Errorf("Severity = %q", d.Severity)
	}
}

func TestJudgedData_Fields(t *testing.T) {
	j := domain.JudgedData{
		PlanID:     "plan-1",
		ScriptPath: "test.js",
		Duration:   "30s",
		VUs:        10,
		Results: domain.K6Results{
			P95LatencyMs: 100,
			ActualRPS:    200,
		},
		Verdict: domain.VerdictPass,
	}
	if j.PlanID != "plan-1" {
		t.Errorf("PlanID = %q", j.PlanID)
	}
	if j.ScriptPath != "test.js" {
		t.Errorf("ScriptPath = %q", j.ScriptPath)
	}
	if j.VUs != 10 {
		t.Errorf("VUs = %d", j.VUs)
	}
	if j.Verdict != domain.VerdictPass {
		t.Errorf("Verdict = %q", j.Verdict)
	}
}

func TestK6Results_Fields(t *testing.T) {
	r := domain.K6Results{
		P95LatencyMs:     123.45,
		ErrorRatePercent: 1.5,
		SuccessRate:      98.5,
		ActualRPS:        500,
	}
	if r.P95LatencyMs != 123.45 {
		t.Errorf("P95LatencyMs = %f", r.P95LatencyMs)
	}
	if r.ErrorRatePercent != 1.5 {
		t.Errorf("ErrorRatePercent = %f", r.ErrorRatePercent)
	}
	if r.SuccessRate != 98.5 {
		t.Errorf("SuccessRate = %f", r.SuccessRate)
	}
	if r.ActualRPS != 500 {
		t.Errorf("ActualRPS = %f", r.ActualRPS)
	}
}

func TestSeverityConstants(t *testing.T) {
	t.Parallel()

	if domain.SeverityLow != "low" {
		t.Errorf("SeverityLow = %q", domain.SeverityLow)
	}
	if domain.SeverityMedium != "medium" {
		t.Errorf("SeverityMedium = %q", domain.SeverityMedium)
	}
	if domain.SeverityHigh != "high" {
		t.Errorf("SeverityHigh = %q", domain.SeverityHigh)
	}
}

func TestEvaluateNfr_RPSViolation(t *testing.T) {
	// given
	results := domain.K6Results{
		P95LatencyMs:     100,
		ErrorRatePercent: 0.5,
		SuccessRate:      99.9,
		ActualRPS:        50,
	}
	nfr := domain.NfrConfig{
		Performance: domain.PerformanceNfr{P95LatencyMs: 200, ErrorRatePercent: 1.0},
		Reliability: domain.ReliabilityNfr{SuccessRatePercent: 99.0},
		Scalability: domain.ScalabilityNfr{TargetRPS: 100},
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
	if deviations[0].Metric != "target_rps" {
		t.Errorf("expected metric target_rps, got %s", deviations[0].Metric)
	}
}
