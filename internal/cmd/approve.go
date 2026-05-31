package cmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/session"
	"github.com/spf13/cobra"
)

func newApproveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approve [path]",
		Short: "Approve an execution plan",
		Long: `Approve a previously created execution plan for load testing.

The --plan-id flag specifies which plan to approve. The approved plan JSON
is output to stdout.`,
		Example: `  # Approve a plan by ID
  dominator approve --plan-id abc123

  # Approve in a specific directory
  dominator approve --plan-id abc123 /path/to/project`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := resolveTargetDir(args)
			if err != nil {
				return err
			}

			err = session.ValidateStateDir(repoRoot)
			if err != nil {
				return err
			}

			planIDRaw, _ := cmd.Flags().GetString("plan-id") // nosemgrep: error-handling.ignored-error-go,error-handling.ignored-error-short-go -- cobra flag registered statically; GetString cannot fail at runtime [permanent]
			planID := domain.PlanID(planIDRaw)

			logger := loggerFrom(cmd)
			stateDir := filepath.Join(repoRoot, domain.StateDir)
			planStore := &session.PlanStore{StateDir: stateDir}

			plan, err := planStore.ApprovePlan(planID)
			if err != nil {
				return err
			}

			// Record event
			eventStore := session.NewEventStore(stateDir, logger)
			ev, evErr := domain.NewEvent(domain.EventPlanApproved, plan, time.Now())
			if evErr == nil {
				_, err = eventStore.Append(ev)
				if err != nil {
					return fmt.Errorf("append event: %w", err)
				}
			}

			data, err := json.MarshalIndent(plan, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal plan: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(data))
			fmt.Fprintf(cmd.ErrOrStderr(), "  Plan %s approved\n", planID)
			return nil
		},
	}
	cmd.Flags().String("plan-id", "", "Plan ID to approve (required)")
	if err := cmd.MarkFlagRequired("plan-id"); err != nil {
		panic(fmt.Sprintf("mark plan-id required: %v", err))
	}
	return cmd
}
