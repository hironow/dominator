package cmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/eventsource"
	"github.com/hironow/dominator/internal/session"
	"github.com/hironow/dominator/internal/usecase"
	"github.com/spf13/cobra"
)

func newRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [path]",
		Short: "Execute k6 load test and judge NFR compliance",
		Long: `Run loads an approved execution plan, executes the k6 load test,
evaluates NFR thresholds, records events, writes insights, and emits
D-Mail notifications on violation.

Exit codes:
  0 = pass (all NFRs met)
  1 = violation (one or more NFR thresholds exceeded)
  2 = error (plan not found, not approved, k6 failure, etc.)`,
		Example: `  # Run a specific plan
  dominator run --plan-id abc123

  # Run in a specific directory
  dominator run --plan-id abc123 /path/to/project`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := resolveTargetDir(args)
			if err != nil {
				return err
			}

			if err := session.ValidateStateDir(repoRoot); err != nil {
				return err
			}

			planIDRaw, _ := cmd.Flags().GetString("plan-id")
			planID := domain.PlanID(planIDRaw)

			logger := loggerFrom(cmd)
			stateDir := filepath.Join(repoRoot, domain.StateDir)

			planStore := &session.PlanStore{StateDir: stateDir}
			k6Runner := &session.K6Runner{Logger: logger}
			eventStore := eventsource.NewFileEventStore(filepath.Join(stateDir, "events"), logger)
			insightWriter := &session.InsightWriter{StateDir: stateDir, Logger: logger}
			dmailEmitter := &session.DMailEmitter{StateDir: stateDir, Logger: logger}

			judged, err := usecase.RunJudge(
				cmd.Context(),
				planID,
				planStore,
				k6Runner,
				eventStore,
				insightWriter,
				dmailEmitter,
				logger,
				cmd.ErrOrStderr(),
			)
			if err != nil {
				return err
			}

			// Output JudgedData as JSON to stdout
			data, marshalErr := json.MarshalIndent(judged, "", "  ")
			if marshalErr != nil {
				return fmt.Errorf("marshal judgment result: %w", marshalErr)
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(data))

			if judged.Verdict == domain.VerdictViolation {
				return &domain.JudgmentError{Violations: len(judged.Deviations)}
			}

			return nil
		},
	}

	cmd.Flags().String("plan-id", "", "Plan ID to execute (required)")
	if err := cmd.MarkFlagRequired("plan-id"); err != nil {
		panic(fmt.Sprintf("mark plan-id required: %v", err))
	}

	return cmd
}
