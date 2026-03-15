package domain

import (
	"time"

	"github.com/google/uuid"
)

// PlanID is a unique identifier for an execution plan.
type PlanID string

// NewPlanID generates a new UUID-based PlanID.
func NewPlanID() PlanID {
	return PlanID(uuid.New().String())
}

// Verdict represents the NFR judgment result.
type Verdict string

const (
	VerdictPass      Verdict = "pass"
	VerdictViolation Verdict = "violation"
)

// Plan is an execution plan for k6 load testing.
type Plan struct {
	ID         PlanID       `json:"plan_id"`
	Script     string       `json:"script"`
	Target     TargetConfig `json:"target"`
	Load       LoadConfig   `json:"load"`
	Nfr        NfrConfig    `json:"nfr"`
	Approved   bool         `json:"approved"`
	ApprovedAt time.Time    `json:"approved_at,omitzero"`
	CreatedAt  time.Time    `json:"created_at"`
}

// NewPlan creates a new Plan with a generated ID and timestamp.
func NewPlan(script string, target TargetConfig, load LoadConfig, nfr NfrConfig) Plan {
	return Plan{
		ID:        NewPlanID(),
		Script:    script,
		Target:    target,
		Load:      load,
		Nfr:       nfr,
		Approved:  false,
		CreatedAt: time.Now().UTC(),
	}
}

// Approve marks the plan as approved.
func (p *Plan) Approve() {
	p.Approved = true
	p.ApprovedAt = time.Now().UTC()
}
