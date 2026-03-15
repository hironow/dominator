package domain

// Severity levels for NFR deviations.
const (
	SeverityLow    = "low"
	SeverityMedium = "medium"
	SeverityHigh   = "high"
)

// K6Results holds the results from a k6 load test run.
type K6Results struct {
	P95LatencyMs     float64 `json:"p95_latency_ms"`
	ErrorRatePercent float64 `json:"error_rate_percent"`
	SuccessRate      float64 `json:"success_rate"`
	ActualRPS        float64 `json:"actual_rps"`
}

// NfrDeviation records a single NFR threshold violation.
type NfrDeviation struct {
	Metric    string  `json:"metric"`
	Threshold float64 `json:"threshold"`
	Actual    float64 `json:"actual"`
	Deviation float64 `json:"deviation_percent"` // (actual - threshold) / threshold * 100
	Severity  string  `json:"severity"`
}

// JudgedData holds the full judgment result.
type JudgedData struct {
	PlanID     string         `json:"plan_id"`
	ScriptPath string         `json:"script_path"`
	Duration   string         `json:"duration"`
	VUs        int            `json:"vus"`
	Results    K6Results      `json:"results"`
	Verdict    Verdict        `json:"verdict"`
	Deviations []NfrDeviation `json:"deviations,omitempty"`
}

// CalcSeverity determines severity from deviation percentage.
// < 10% -> low, 10-50% -> medium, > 50% -> high
func CalcSeverity(deviationPercent float64) string {
	if deviationPercent < 0 {
		deviationPercent = -deviationPercent
	}
	switch {
	case deviationPercent > 50:
		return SeverityHigh
	case deviationPercent >= 10:
		return SeverityMedium
	default:
		return SeverityLow
	}
}

// CalcDeviation computes the deviation percentage between actual and threshold.
// For metrics where lower is better (latency), positive deviation = over threshold.
// For metrics where higher is better (success rate), positive deviation = under threshold.
func CalcDeviation(actual, threshold float64) float64 {
	if threshold == 0 {
		return 0
	}
	return (actual - threshold) / threshold * 100
}

// EvaluateNfr compares K6Results against NfrConfig and returns deviations.
func EvaluateNfr(results K6Results, nfr NfrConfig) (Verdict, []NfrDeviation) {
	var deviations []NfrDeviation

	// p95 latency (lower is better — positive deviation = violation)
	if nfr.Performance.P95LatencyMs > 0 {
		dev := CalcDeviation(results.P95LatencyMs, float64(nfr.Performance.P95LatencyMs))
		if dev > 0 {
			deviations = append(deviations, NfrDeviation{
				Metric:    "p95_latency_ms",
				Threshold: float64(nfr.Performance.P95LatencyMs),
				Actual:    results.P95LatencyMs,
				Deviation: dev,
				Severity:  CalcSeverity(dev),
			})
		}
	}

	// error rate (lower is better — positive deviation = violation)
	if nfr.Performance.ErrorRatePercent > 0 {
		dev := CalcDeviation(results.ErrorRatePercent, nfr.Performance.ErrorRatePercent)
		if dev > 0 {
			deviations = append(deviations, NfrDeviation{
				Metric:    "error_rate_percent",
				Threshold: nfr.Performance.ErrorRatePercent,
				Actual:    results.ErrorRatePercent,
				Deviation: dev,
				Severity:  CalcSeverity(dev),
			})
		}
	}

	// success rate (higher is better — NEGATIVE deviation = violation)
	if nfr.Reliability.SuccessRatePercent > 0 {
		dev := CalcDeviation(nfr.Reliability.SuccessRatePercent, results.SuccessRate)
		if dev > 0 { // threshold > actual = violation
			deviations = append(deviations, NfrDeviation{
				Metric:    "success_rate_percent",
				Threshold: nfr.Reliability.SuccessRatePercent,
				Actual:    results.SuccessRate,
				Deviation: dev,
				Severity:  CalcSeverity(dev),
			})
		}
	}

	// target RPS (higher is better — NEGATIVE deviation = violation)
	if nfr.Scalability.TargetRPS > 0 {
		dev := CalcDeviation(float64(nfr.Scalability.TargetRPS), results.ActualRPS)
		if dev > 0 { // threshold > actual = violation
			deviations = append(deviations, NfrDeviation{
				Metric:    "target_rps",
				Threshold: float64(nfr.Scalability.TargetRPS),
				Actual:    results.ActualRPS,
				Deviation: dev,
				Severity:  CalcSeverity(dev),
			})
		}
	}

	if len(deviations) > 0 {
		return VerdictViolation, deviations
	}
	return VerdictPass, nil
}
