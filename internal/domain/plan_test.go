package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/hironow/dominator/internal/domain"
)

func TestNewPlan_HasUUID(t *testing.T) {
	// given
	script := "import http from 'k6/http';"
	target := domain.TargetConfig{URL: "http://localhost:8080"}
	load := domain.LoadConfig{VUs: 10, Duration: "30s"}
	nfr := domain.NfrConfig{}

	// when
	plan := domain.NewPlan(script, target, load, nfr)

	// then
	if plan.ID == "" {
		t.Error("expected plan ID to be a non-empty UUID")
	}
	if len(string(plan.ID)) != 36 {
		t.Errorf("expected UUID format (36 chars), got %d chars: %s", len(string(plan.ID)), plan.ID)
	}
}

func TestNewPlan_NotApproved(t *testing.T) {
	// given
	script := "import http from 'k6/http';"
	target := domain.TargetConfig{URL: "http://localhost:8080"}
	load := domain.LoadConfig{VUs: 10, Duration: "30s"}
	nfr := domain.NfrConfig{}

	// when
	plan := domain.NewPlan(script, target, load, nfr)

	// then
	if plan.Approved {
		t.Error("expected new plan to not be approved")
	}
	if !plan.ApprovedAt.IsZero() {
		t.Error("expected ApprovedAt to be zero for unapproved plan")
	}
}

func TestPlan_Approve_SetsFlag(t *testing.T) {
	// given
	plan := domain.NewPlan("script", domain.TargetConfig{}, domain.LoadConfig{}, domain.NfrConfig{})

	// when
	plan.Approve()

	// then
	if !plan.Approved {
		t.Error("expected plan to be approved after Approve()")
	}
	if plan.ApprovedAt.IsZero() {
		t.Error("expected ApprovedAt to be set after Approve()")
	}
}

func TestNewPlanID_Unique(t *testing.T) {
	// given / when
	id1 := domain.NewPlanID()
	id2 := domain.NewPlanID()

	// then
	if id1 == id2 {
		t.Errorf("expected unique plan IDs, got %s and %s", id1, id2)
	}
}

func TestPlan_DoubleApprove_Idempotent(t *testing.T) {
	// given
	plan := domain.NewPlan("script", domain.TargetConfig{}, domain.LoadConfig{}, domain.NfrConfig{})

	// when: approve twice
	plan.Approve()
	firstApprovedAt := plan.ApprovedAt
	plan.Approve()

	// then: still approved, no panic
	if !plan.Approved {
		t.Error("expected plan to remain approved")
	}
	// ApprovedAt may be updated but should not be zero
	if plan.ApprovedAt.IsZero() {
		t.Error("expected ApprovedAt to remain set")
	}
	_ = firstApprovedAt // just verify no panic on double approve
}

func TestNewPlan_FieldsSet(t *testing.T) {
	// given
	script := "import http from 'k6/http'; export default function() {}"
	target := domain.TargetConfig{
		URL:      "http://localhost:8080",
		Protocol: "openapi",
		Spec:     "https://example.com/spec.json",
		Docs:     "https://example.com/docs",
	}
	load := domain.LoadConfig{VUs: 20, Duration: "60s", RampUp: "10s"}
	nfr := domain.NfrConfig{
		Performance: domain.PerformanceNfr{P95LatencyMs: 500, ErrorRatePercent: 1.0},
		Reliability: domain.ReliabilityNfr{SuccessRatePercent: 99.0},
		Scalability: domain.ScalabilityNfr{TargetRPS: 200},
	}

	// when
	plan := domain.NewPlan(script, target, load, nfr)

	// then
	if plan.ID == "" {
		t.Error("ID should be set")
	}
	if plan.Script != script {
		t.Errorf("Script = %q, want %q", plan.Script, script)
	}
	if plan.Target.URL != target.URL {
		t.Errorf("Target.URL = %q, want %q", plan.Target.URL, target.URL)
	}
	if plan.Target.Protocol != target.Protocol {
		t.Errorf("Target.Protocol = %q, want %q", plan.Target.Protocol, target.Protocol)
	}
	if plan.Load.VUs != load.VUs {
		t.Errorf("Load.VUs = %d, want %d", plan.Load.VUs, load.VUs)
	}
	if plan.Load.Duration != load.Duration {
		t.Errorf("Load.Duration = %q, want %q", plan.Load.Duration, load.Duration)
	}
	if plan.Load.RampUp != load.RampUp {
		t.Errorf("Load.RampUp = %q, want %q", plan.Load.RampUp, load.RampUp)
	}
	if plan.Nfr.Performance.P95LatencyMs != nfr.Performance.P95LatencyMs {
		t.Errorf("Nfr.Performance.P95LatencyMs = %d, want %d", plan.Nfr.Performance.P95LatencyMs, nfr.Performance.P95LatencyMs)
	}
	if plan.Nfr.Reliability.SuccessRatePercent != nfr.Reliability.SuccessRatePercent {
		t.Errorf("Nfr.Reliability.SuccessRatePercent = %f, want %f", plan.Nfr.Reliability.SuccessRatePercent, nfr.Reliability.SuccessRatePercent)
	}
	if plan.Nfr.Scalability.TargetRPS != nfr.Scalability.TargetRPS {
		t.Errorf("Nfr.Scalability.TargetRPS = %d, want %d", plan.Nfr.Scalability.TargetRPS, nfr.Scalability.TargetRPS)
	}
	if plan.Approved {
		t.Error("new plan should not be approved")
	}
	if plan.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestPlan_JSONRoundTrip(t *testing.T) {
	// given
	plan := domain.NewPlan(
		"test-script.js",
		domain.TargetConfig{URL: "http://localhost:8080", Protocol: "openapi"},
		domain.LoadConfig{VUs: 10, Duration: "30s", RampUp: "5s"},
		domain.NfrConfig{
			Performance: domain.PerformanceNfr{P95LatencyMs: 500, ErrorRatePercent: 1.0},
			Reliability: domain.ReliabilityNfr{SuccessRatePercent: 99.0},
			Scalability: domain.ScalabilityNfr{TargetRPS: 100},
		},
	)
	plan.Approve()

	// when
	data, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var restored domain.Plan
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// then
	if restored.ID != plan.ID {
		t.Errorf("ID = %s, want %s", restored.ID, plan.ID)
	}
	if restored.Script != plan.Script {
		t.Errorf("Script = %q, want %q", restored.Script, plan.Script)
	}
	if restored.Approved != plan.Approved {
		t.Errorf("Approved = %v, want %v", restored.Approved, plan.Approved)
	}
	if restored.Target.URL != plan.Target.URL {
		t.Errorf("Target.URL = %q, want %q", restored.Target.URL, plan.Target.URL)
	}
	if restored.Load.VUs != plan.Load.VUs {
		t.Errorf("Load.VUs = %d, want %d", restored.Load.VUs, plan.Load.VUs)
	}
	if restored.Nfr.Performance.P95LatencyMs != plan.Nfr.Performance.P95LatencyMs {
		t.Errorf("P95LatencyMs = %d, want %d", restored.Nfr.Performance.P95LatencyMs, plan.Nfr.Performance.P95LatencyMs)
	}
}

func TestNewPlan_CreatedAtIsUTC(t *testing.T) {
	// given / when
	plan := domain.NewPlan("test.js", domain.TargetConfig{}, domain.LoadConfig{}, domain.NfrConfig{})

	// then
	if plan.CreatedAt.Location().String() != "UTC" {
		t.Errorf("CreatedAt location = %q, want UTC", plan.CreatedAt.Location().String())
	}
}

func TestPlan_Approve_SetsUTCTime(t *testing.T) {
	// given
	plan := domain.NewPlan("test.js", domain.TargetConfig{}, domain.LoadConfig{}, domain.NfrConfig{})

	// when
	plan.Approve()

	// then
	if plan.ApprovedAt.Location().String() != "UTC" {
		t.Errorf("ApprovedAt location = %q, want UTC", plan.ApprovedAt.Location().String())
	}
}

func TestVerdictConstants(t *testing.T) {
	t.Parallel()

	if domain.VerdictPass != "pass" {
		t.Errorf("VerdictPass = %q, want %q", domain.VerdictPass, "pass")
	}
	if domain.VerdictViolation != "violation" {
		t.Errorf("VerdictViolation = %q, want %q", domain.VerdictViolation, "violation")
	}
}
