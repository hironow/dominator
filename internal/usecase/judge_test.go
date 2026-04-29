package usecase_test

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/usecase"
)

// --- Test doubles ---

type fakePlanStore struct {
	plan *domain.Plan
	err  error
}

func (f *fakePlanStore) SavePlan(_ domain.Plan) error                      { return nil }
func (f *fakePlanStore) LoadPlan(_ domain.PlanID) (*domain.Plan, error)    { return f.plan, f.err }
func (f *fakePlanStore) LoadLatestPlan() (*domain.Plan, error)             { return f.plan, f.err }
func (f *fakePlanStore) ApprovePlan(_ domain.PlanID) (*domain.Plan, error) { return f.plan, f.err }
func (f *fakePlanStore) ListScripts() ([]string, error)                    { return nil, nil }

type fakeK6Runner struct {
	results domain.K6Results
	err     error
}

func (f *fakeK6Runner) Run(_ context.Context, _ string, _ domain.LoadConfig, _ io.Writer) (domain.K6Results, error) {
	return f.results, f.err
}

func (f *fakeK6Runner) Validate(_ context.Context, _ string) error {
	return nil
}

type fakeEventStore struct {
	events []domain.Event
}

func (f *fakeEventStore) Append(events ...domain.Event) (domain.AppendResult, error) {
	f.events = append(f.events, events...)
	return domain.AppendResult{BytesWritten: 1}, nil
}

func (f *fakeEventStore) LoadAll() ([]domain.Event, domain.LoadResult, error) {
	return f.events, domain.LoadResult{}, nil
}

func (f *fakeEventStore) LoadSince(_ time.Time) ([]domain.Event, domain.LoadResult, error) {
	return f.events, domain.LoadResult{}, nil
}

type fakeInsightWriter struct {
	hueCount         int
	coefficientCount int
}

func (f *fakeInsightWriter) RecordHue(_ domain.JudgedData) error {
	f.hueCount++
	return nil
}

func (f *fakeInsightWriter) RecordCoefficient(_ domain.JudgedData) error {
	f.coefficientCount++
	return nil
}

type fakeDMailEmitter struct {
	violationCount int
	passCount      int
}

func (f *fakeDMailEmitter) EmitViolation(_ domain.JudgedData) error {
	f.violationCount++
	return nil
}

func (f *fakeDMailEmitter) EmitPass(_ domain.JudgedData) error {
	f.passCount++
	return nil
}

// --- Tests ---

func TestRunJudge_Pass(t *testing.T) {
	// given
	plan := &domain.Plan{
		ID:       "plan-001",
		Script:   "test.js",
		Approved: true,
		Load:     domain.LoadConfig{VUs: 10, Duration: "30s"},
		Nfr: domain.NfrConfig{
			Performance: domain.PerformanceNfr{P95LatencyMs: 500, ErrorRatePercent: 1.0},
			Reliability: domain.ReliabilityNfr{SuccessRatePercent: 99.0},
			Scalability: domain.ScalabilityNfr{TargetRPS: 100},
		},
	}
	planStore := &fakePlanStore{plan: plan}
	k6Runner := &fakeK6Runner{results: domain.K6Results{
		P95LatencyMs:     200,
		ErrorRatePercent: 0.5,
		SuccessRate:      99.5,
		ActualRPS:        150,
	}}
	eventStore := &fakeEventStore{}
	insightWriter := &fakeInsightWriter{}
	dmailEmitter := &fakeDMailEmitter{}
	buf := new(bytes.Buffer)

	// when
	judged, err := usecase.RunJudge(
		context.Background(),
		"plan-001",
		t.TempDir(),
		planStore,
		k6Runner,
		eventStore,
		insightWriter,
		dmailEmitter,
		&domain.NopLogger{},
		buf,
	)

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if judged.Verdict != domain.VerdictPass {
		t.Errorf("verdict = %s, want pass", judged.Verdict)
	}
	if len(judged.Deviations) != 0 {
		t.Errorf("deviations = %d, want 0", len(judged.Deviations))
	}
	if insightWriter.hueCount != 1 {
		t.Errorf("hue recordings = %d, want 1", insightWriter.hueCount)
	}
	if dmailEmitter.passCount != 1 {
		t.Errorf("pass D-Mails = %d, want 1", dmailEmitter.passCount)
	}
	if dmailEmitter.violationCount != 0 {
		t.Errorf("violation D-Mails = %d, want 0", dmailEmitter.violationCount)
	}
	// Should have 2 events: judged + pass.confirmed
	if len(eventStore.events) != 2 {
		t.Errorf("events = %d, want 2", len(eventStore.events))
	}
}

func TestRunJudge_Violation(t *testing.T) {
	// given
	plan := &domain.Plan{
		ID:       "plan-002",
		Script:   "test.js",
		Approved: true,
		Load:     domain.LoadConfig{VUs: 10, Duration: "30s"},
		Nfr: domain.NfrConfig{
			Performance: domain.PerformanceNfr{P95LatencyMs: 500, ErrorRatePercent: 1.0},
			Reliability: domain.ReliabilityNfr{SuccessRatePercent: 99.0},
			Scalability: domain.ScalabilityNfr{TargetRPS: 100},
		},
	}
	planStore := &fakePlanStore{plan: plan}
	k6Runner := &fakeK6Runner{results: domain.K6Results{
		P95LatencyMs:     800, // over threshold of 500
		ErrorRatePercent: 0.5,
		SuccessRate:      99.5,
		ActualRPS:        150,
	}}
	eventStore := &fakeEventStore{}
	insightWriter := &fakeInsightWriter{}
	dmailEmitter := &fakeDMailEmitter{}
	buf := new(bytes.Buffer)

	// when
	judged, err := usecase.RunJudge(
		context.Background(),
		"plan-002",
		t.TempDir(),
		planStore,
		k6Runner,
		eventStore,
		insightWriter,
		dmailEmitter,
		&domain.NopLogger{},
		buf,
	)

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if judged.Verdict != domain.VerdictViolation {
		t.Errorf("verdict = %s, want violation", judged.Verdict)
	}
	if len(judged.Deviations) == 0 {
		t.Error("expected at least one deviation")
	}
	if insightWriter.hueCount != 1 {
		t.Errorf("hue recordings = %d, want 1", insightWriter.hueCount)
	}
	if insightWriter.coefficientCount != 1 {
		t.Errorf("coefficient recordings = %d, want 1", insightWriter.coefficientCount)
	}
	if dmailEmitter.violationCount != 1 {
		t.Errorf("violation D-Mails = %d, want 1", dmailEmitter.violationCount)
	}
	if dmailEmitter.passCount != 0 {
		t.Errorf("pass D-Mails = %d, want 0", dmailEmitter.passCount)
	}
	// Should have 2 events: judged + violation.detected
	if len(eventStore.events) != 2 {
		t.Errorf("events = %d, want 2", len(eventStore.events))
	}
}

func TestRunJudge_NotApproved(t *testing.T) {
	// given
	plan := &domain.Plan{
		ID:       "plan-003",
		Script:   "test.js",
		Approved: false,
		Load:     domain.LoadConfig{VUs: 10, Duration: "30s"},
	}
	planStore := &fakePlanStore{plan: plan}
	k6Runner := &fakeK6Runner{}
	eventStore := &fakeEventStore{}
	insightWriter := &fakeInsightWriter{}
	dmailEmitter := &fakeDMailEmitter{}
	buf := new(bytes.Buffer)

	// when
	_, err := usecase.RunJudge(
		context.Background(),
		"plan-003",
		t.TempDir(),
		planStore,
		k6Runner,
		eventStore,
		insightWriter,
		dmailEmitter,
		&domain.NopLogger{},
		buf,
	)

	// then
	if err == nil {
		t.Fatal("expected error for unapproved plan, got nil")
	}
}
