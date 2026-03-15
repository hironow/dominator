package port

import "github.com/hironow/dominator/internal/domain"

// PlanStore persists and retrieves execution plans.
type PlanStore interface {
	SavePlan(plan domain.Plan) error
	LoadPlan(planID domain.PlanID) (*domain.Plan, error)
	LoadLatestPlan() (*domain.Plan, error)
	ApprovePlan(planID domain.PlanID) (*domain.Plan, error)
	ListScripts() ([]string, error)
}
