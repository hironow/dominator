package usecase

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/usecase/port"
)

// JudgeInsightWriter records judgment insights.
type JudgeInsightWriter interface {
	RecordHue(result domain.JudgedData) error
	RecordCoefficient(result domain.JudgedData) error
}

// JudgeDMailEmitter emits D-Mail notifications based on judgment results.
type JudgeDMailEmitter interface {
	EmitViolation(result domain.JudgedData) error
	EmitPass(result domain.JudgedData) error
}

// RunJudge loads an approved plan, executes k6, evaluates NFR thresholds,
// records events, writes insights, and emits D-Mail on violation.
func RunJudge(
	ctx context.Context,
	planID domain.PlanID,
	stateDir string,
	planStore port.PlanStore,
	k6Runner port.K6Runner,
	eventStore port.EventStore,
	insightWriter JudgeInsightWriter,
	dmailEmitter JudgeDMailEmitter,
	logger domain.Logger,
	stderrW io.Writer,
) (domain.JudgedData, error) {
	// 1. Load approved plan
	plan, err := planStore.LoadPlan(planID)
	if err != nil {
		return domain.JudgedData{}, fmt.Errorf("load plan: %w", err)
	}
	if !plan.Approved {
		return domain.JudgedData{}, fmt.Errorf("plan %s is not approved — run 'dominator approve --plan-id %s' first", planID, planID)
	}

	// Resolve script path relative to state dir (plan.Script is filename only)
	scriptPath := filepath.Join(stateDir, "k6-scripts", plan.Script)
	logger.Info("Executing plan %s (script: %s)", planID, scriptPath)

	// 2. Run k6
	results, err := k6Runner.Run(ctx, scriptPath, plan.Load, stderrW)
	if err != nil {
		return domain.JudgedData{}, fmt.Errorf("k6 run: %w", err)
	}

	// 3. Evaluate NFR
	verdict, deviations := domain.EvaluateNfr(results, plan.Nfr)

	// 4. Build JudgedData
	judged := domain.JudgedData{
		PlanID:     string(planID),
		ScriptPath: plan.Script,
		Duration:   plan.Load.Duration,
		VUs:        plan.Load.VUs,
		Results:    results,
		Verdict:    verdict,
		Deviations: deviations,
	}

	// 5. Write insights
	if insightErr := insightWriter.RecordHue(judged); insightErr != nil {
		logger.Warn("Failed to record hue insight: %v", insightErr)
	}
	if insightErr := insightWriter.RecordCoefficient(judged); insightErr != nil {
		logger.Warn("Failed to record coefficient insight: %v", insightErr)
	}

	// 6. Emit D-Mail
	if verdict == domain.VerdictViolation {
		if dmailErr := dmailEmitter.EmitViolation(judged); dmailErr != nil {
			logger.Warn("Failed to emit violation D-Mail: %v", dmailErr)
		}
	} else {
		if dmailErr := dmailEmitter.EmitPass(judged); dmailErr != nil {
			logger.Warn("Failed to emit pass D-Mail: %v", dmailErr)
		}
	}

	// 7. Record events (best-effort — log but do not fail judgment)
	now := time.Now()
	judgedEvent, err := domain.NewEvent(domain.EventJudged, judged, now)
	if err != nil {
		logger.Warn("Failed to create judged event: %v", err)
	} else if _, appendErr := eventStore.Append(judgedEvent); appendErr != nil {
		logger.Warn("Failed to append judged event: %v", appendErr)
	}

	if verdict == domain.VerdictViolation {
		violationEvent, err := domain.NewEvent(domain.EventViolationDetected, judged, now)
		if err != nil {
			logger.Warn("Failed to create violation event: %v", err)
		} else if _, appendErr := eventStore.Append(violationEvent); appendErr != nil {
			logger.Warn("Failed to append violation event: %v", appendErr)
		}
	} else {
		passEvent, err := domain.NewEvent(domain.EventPassConfirmed, judged, now)
		if err != nil {
			logger.Warn("Failed to create pass event: %v", err)
		} else if _, appendErr := eventStore.Append(passEvent); appendErr != nil {
			logger.Warn("Failed to append pass event: %v", appendErr)
		}
	}

	// 8. Return
	logger.Info("Judgment complete: %s (%d deviations)", verdict, len(deviations))
	return judged, nil
}
