package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/usecase/port"
)

// RunCheck loads config, lists k6 scripts, creates a Plan for the first script,
// saves it to the PlanStore, and records a plan.created event.
func RunCheck(
	ctx context.Context,
	stateDir string,
	planStore port.PlanStore,
	configLoader port.ConfigLoader,
	eventStore port.EventStore,
	logger domain.Logger,
) (*domain.Plan, error) {
	cfg, err := configLoader.Load(stateDir)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	scripts, err := planStore.ListScripts()
	if err != nil {
		return nil, fmt.Errorf("list scripts: %w", err)
	}
	if len(scripts) == 0 {
		return nil, fmt.Errorf("no k6 scripts found in k6-scripts/")
	}

	// Create a plan from the first available script
	plan := domain.NewPlan(scripts[0], cfg.Target, cfg.Load, cfg.Nfr)

	if err := planStore.SavePlan(plan); err != nil {
		return nil, fmt.Errorf("save plan: %w", err)
	}

	// Record event
	ev, evErr := domain.NewEvent(domain.EventPlanCreated, plan, time.Now())
	if evErr == nil {
		eventStore.Append(ev)
	}

	logger.Info("Plan created: %s (script: %s)", plan.ID, plan.Script)
	return &plan, nil
}
