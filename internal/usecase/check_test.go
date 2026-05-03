package usecase_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/usecase"
)

// --- Test doubles for RunCheck ---

type fakeCheckPlanStore struct {
	saved   []domain.Plan
	scripts []string
}

func (f *fakeCheckPlanStore) SavePlan(p domain.Plan) error {
	f.saved = append(f.saved, p)
	return nil
}

func (f *fakeCheckPlanStore) LoadPlan(_ domain.PlanID) (*domain.Plan, error)    { return nil, nil }
func (f *fakeCheckPlanStore) LoadLatestPlan() (*domain.Plan, error)             { return nil, nil }
func (f *fakeCheckPlanStore) ApprovePlan(_ domain.PlanID) (*domain.Plan, error) { return nil, nil }
func (f *fakeCheckPlanStore) ListScripts() ([]string, error)                    { return f.scripts, nil }

type fakeConfigLoader struct {
	cfg domain.Config
}

func (f *fakeConfigLoader) Load(_ string) (domain.Config, error) { return f.cfg, nil }

type fakeCheckEventStore struct {
	events []domain.Event
}

func (f *fakeCheckEventStore) Append(events ...domain.Event) (domain.AppendResult, error) {
	f.events = append(f.events, events...)
	return domain.AppendResult{BytesWritten: 1}, nil
}

func (f *fakeCheckEventStore) LoadAll() ([]domain.Event, domain.LoadResult, error) {
	return f.events, domain.LoadResult{}, nil
}

func (f *fakeCheckEventStore) LoadSince(_ time.Time) ([]domain.Event, domain.LoadResult, error) {
	return f.events, domain.LoadResult{}, nil
}

// fakeContractReader returns a fixed CurrentContract or nil to simulate
// "no contract present in inbox".
type fakeContractReader struct {
	contract *domain.CurrentContract
	err      error
}

func (f *fakeContractReader) LoadCurrentContract() (*domain.CurrentContract, error) {
	return f.contract, f.err
}

// fakeCheckDMailEmitter records design-feedback emissions for missing
// contract NFR thresholds.
type fakeCheckDMailEmitter struct {
	designFeedbackCount   int
	designFeedbackContent string
}

func (f *fakeCheckDMailEmitter) EmitDesignFeedbackMissingNfr(missing []string, contractID string) error {
	f.designFeedbackCount++
	f.designFeedbackContent = strings.Join(missing, ",") + "|" + contractID
	return nil
}

// --- Tests ---

func TestCheckPlan_UsesContractNfrDefaults(t *testing.T) {
	// given: a config with default NFR thresholds and a Rival Contract v1
	// with concrete nfr.* evidence; the contract values must override the
	// config defaults in the resulting plan.
	cfg := domain.DefaultConfig()
	cfg.Target.URL = "http://localhost:8080"
	planStore := &fakeCheckPlanStore{scripts: []string{"load.js"}}
	configLoader := &fakeConfigLoader{cfg: cfg}
	eventStore := &fakeCheckEventStore{}
	contract := &domain.CurrentContract{
		DMailName: "spec-checkout-budget_a3f2b7c4",
		Metadata: domain.RivalContractMetadata{
			Schema:   domain.SchemaRivalContractV1,
			ID:       "checkout-budget",
			Revision: 1,
		},
		Contract: domain.RivalContract{
			Evidence: strings.Join([]string{
				"- nfr.p95_latency_ms: <= 250",
				"- nfr.error_rate_percent: <= 0.5",
				"- nfr.success_rate_percent: >= 99.9",
				"- nfr.target_rps: >= 200",
			}, "\n"),
		},
	}
	contractReader := &fakeContractReader{contract: contract}
	dmailEmitter := &fakeCheckDMailEmitter{}

	// when
	plan, err := usecase.RunCheck(
		context.Background(),
		t.TempDir(),
		planStore,
		configLoader,
		eventStore,
		contractReader,
		dmailEmitter,
		&domain.NopLogger{},
	)

	// then
	if err != nil {
		t.Fatalf("RunCheck: %v", err)
	}
	if plan.Nfr.Performance.P95LatencyMs != 250 {
		t.Errorf("plan.Nfr.Performance.P95LatencyMs = %d, want 250 (from contract)", plan.Nfr.Performance.P95LatencyMs)
	}
	if plan.Nfr.Performance.ErrorRatePercent != 0.5 {
		t.Errorf("plan.Nfr.Performance.ErrorRatePercent = %v, want 0.5 (from contract)", plan.Nfr.Performance.ErrorRatePercent)
	}
	if plan.Nfr.Reliability.SuccessRatePercent != 99.9 {
		t.Errorf("plan.Nfr.Reliability.SuccessRatePercent = %v, want 99.9 (from contract)", plan.Nfr.Reliability.SuccessRatePercent)
	}
	if plan.Nfr.Scalability.TargetRPS != 200 {
		t.Errorf("plan.Nfr.Scalability.TargetRPS = %d, want 200 (from contract)", plan.Nfr.Scalability.TargetRPS)
	}
	if dmailEmitter.designFeedbackCount != 0 {
		t.Errorf("expected no design-feedback (contract has thresholds), got %d", dmailEmitter.designFeedbackCount)
	}
}

func TestCheckPlan_MissingContractNfrRequestsDesignFeedback(t *testing.T) {
	// given: a contract with zero nfr.* evidence keys (only check/test
	// non-NFR bullets and prose) — the load test cannot determine
	// thresholds deterministically, so a design-feedback D-Mail must be
	// emitted asking sightjack to add NFR thresholds to the contract.
	cfg := domain.DefaultConfig()
	cfg.Target.URL = "http://localhost:8080"
	planStore := &fakeCheckPlanStore{scripts: []string{"load.js"}}
	configLoader := &fakeConfigLoader{cfg: cfg}
	eventStore := &fakeCheckEventStore{}
	contract := &domain.CurrentContract{
		DMailName: "spec-checkout-budget_a3f2b7c4",
		Metadata: domain.RivalContractMetadata{
			Schema:   domain.SchemaRivalContractV1,
			ID:       "checkout-budget",
			Revision: 1,
		},
		Contract: domain.RivalContract{
			Evidence: strings.Join([]string{
				"- check: just check",
				"- test: just test",
				"- Add a regression test under 100 VUs.",
			}, "\n"),
		},
	}
	contractReader := &fakeContractReader{contract: contract}
	dmailEmitter := &fakeCheckDMailEmitter{}

	// when
	_, err := usecase.RunCheck(
		context.Background(),
		t.TempDir(),
		planStore,
		configLoader,
		eventStore,
		contractReader,
		dmailEmitter,
		&domain.NopLogger{},
	)

	// then
	if err != nil {
		t.Fatalf("RunCheck: %v", err)
	}
	if dmailEmitter.designFeedbackCount != 1 {
		t.Errorf("expected exactly 1 design-feedback emission for missing NFR, got %d", dmailEmitter.designFeedbackCount)
	}
	if !strings.Contains(dmailEmitter.designFeedbackContent, "nfr.p95_latency_ms") {
		t.Errorf("design-feedback should mention required missing key 'nfr.p95_latency_ms', got %q", dmailEmitter.designFeedbackContent)
	}
	if !strings.Contains(dmailEmitter.designFeedbackContent, "checkout-budget") {
		t.Errorf("design-feedback should mention contract id, got %q", dmailEmitter.designFeedbackContent)
	}
}

func TestCheckPlan_NoContractFallsBackToConfig(t *testing.T) {
	// given: no contract is present (legacy/free-form mode); the plan
	// must use config NFR defaults verbatim, and no design-feedback is
	// emitted.
	cfg := domain.DefaultConfig()
	cfg.Target.URL = "http://localhost:8080"
	planStore := &fakeCheckPlanStore{scripts: []string{"load.js"}}
	configLoader := &fakeConfigLoader{cfg: cfg}
	eventStore := &fakeCheckEventStore{}
	contractReader := &fakeContractReader{contract: nil}
	dmailEmitter := &fakeCheckDMailEmitter{}

	// when
	plan, err := usecase.RunCheck(
		context.Background(),
		t.TempDir(),
		planStore,
		configLoader,
		eventStore,
		contractReader,
		dmailEmitter,
		&domain.NopLogger{},
	)

	// then
	if err != nil {
		t.Fatalf("RunCheck: %v", err)
	}
	if plan.Nfr.Performance.P95LatencyMs != cfg.Nfr.Performance.P95LatencyMs {
		t.Errorf("plan should fall back to config defaults; P95LatencyMs got %d want %d",
			plan.Nfr.Performance.P95LatencyMs, cfg.Nfr.Performance.P95LatencyMs)
	}
	if dmailEmitter.designFeedbackCount != 0 {
		t.Errorf("no contract -> no design-feedback; got %d", dmailEmitter.designFeedbackCount)
	}
}
