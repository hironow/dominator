package domain_test

import (
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
