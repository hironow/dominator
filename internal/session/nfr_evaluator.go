package session

import "github.com/hironow/dominator/internal/domain"

// NfrEvaluator wraps domain.EvaluateNfr for use in the session layer.
type NfrEvaluator struct{}

// Evaluate compares K6Results against NfrConfig and returns deviations.
func (e *NfrEvaluator) Evaluate(results domain.K6Results, nfr domain.NfrConfig) (domain.Verdict, []domain.NfrDeviation) {
	return domain.EvaluateNfr(results, nfr)
}
