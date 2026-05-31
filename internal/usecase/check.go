package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/usecase/port"
)

// CheckDMailEmitter emits a design-feedback D-Mail when the current
// Rival Contract v1 is missing required `nfr.*` thresholds. The
// dominator must not invent thresholds; instead it asks the contract
// author (sightjack) to add them via this design-feedback message.
type CheckDMailEmitter interface {
	EmitDesignFeedbackMissingNfr(missing []string, contractID string) error
}

// RunCheck loads config, lists k6 scripts, creates a Plan for the first script,
// saves it to the PlanStore, and records a plan.created event.
//
// When contractReader returns a Rival Contract v1, its `nfr.*` evidence
// values override config defaults. When the contract has no `nfr.*`
// evidence for required keys (currently `nfr.p95_latency_ms`), a
// design-feedback D-Mail is emitted via dmailEmitter and the plan still
// uses config defaults so the load test can proceed at-least-once.
//
// Both contractReader and dmailEmitter may be nil to preserve legacy
// behavior (config-only NFR defaults, no contract-aware emission).
func RunCheck(
	ctx context.Context,
	stateDir string,
	planStore port.PlanStore,
	configLoader port.ConfigLoader,
	eventStore port.EventStore,
	contractReader port.ContractReader,
	dmailEmitter CheckDMailEmitter,
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

	// Resolve NFR thresholds: contract overrides config defaults. When the
	// contract is missing required nfr.* keys, emit a design-feedback
	// D-Mail asking sightjack to add them rather than inventing values.
	planNfr := cfg.Nfr
	if contractReader != nil {
		current, ctErr := contractReader.LoadCurrentContract()
		if ctErr != nil {
			logger.Warn("Failed to load current contract; falling back to config NFR defaults: %v", ctErr)
		} else if current != nil {
			merged, missing := domain.MergeContractNfrIntoConfig(cfg.Nfr, current.Contract)
			planNfr = merged
			if len(missing) > 0 && dmailEmitter != nil {
				if emitErr := dmailEmitter.EmitDesignFeedbackMissingNfr(missing, current.Metadata.ID); emitErr != nil {
					logger.Warn("Failed to emit design-feedback for missing contract NFR keys: %v", emitErr)
				}
			}
		}
	}

	// Create a plan from the first available script
	plan := domain.NewPlan(scripts[0], cfg.Target, cfg.Load, planNfr)

	if err := planStore.SavePlan(plan); err != nil {
		return nil, fmt.Errorf("save plan: %w", err)
	}

	// Record event
	ev, evErr := domain.NewEvent(domain.EventPlanCreated, plan, time.Now())
	if evErr == nil {
		if _, err := eventStore.Append(ev); err != nil {
			return nil, fmt.Errorf("append event: %w", err)
		}
	}

	logger.Info("Plan created: %s (script: %s)", plan.ID, plan.Script)
	return &plan, nil
}
