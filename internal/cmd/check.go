package cmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/session"
	"github.com/hironow/dominator/internal/usecase"
	"github.com/spf13/cobra"
)

func newCheckCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check [path]",
		Short: "Create an execution plan from k6 scripts",
		Long: `Inspect available k6 scripts and create an execution plan for load testing.

Reads configuration and lists k6 scripts in .pass/k6-scripts/. If scripts are
found, a Plan is created and saved to .pass/.run/plans/. The plan JSON is
output to stdout for piping to other tools.`,
		Example: `  # Check current directory
  dominator check

  # Check a specific project directory
  dominator check /path/to/project`,
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

			logger := loggerFrom(cmd)
			stateDir := filepath.Join(repoRoot, domain.StateDir)

			planStore := &session.PlanStore{StateDir: stateDir}
			configLoader := &session.FileConfigLoader{}
			eventStore := session.NewEventStore(stateDir, logger)
			contractReader := &session.InboxContractReader{StateDir: stateDir, Logger: logger}
			dmailEmitter := &session.DMailEmitter{StateDir: stateDir, Logger: logger}

			plan, err := usecase.RunCheck(
				cmd.Context(),
				stateDir,
				planStore,
				configLoader,
				eventStore,
				contractReader,
				dmailEmitter,
				logger,
			)
			if err != nil {
				return err
			}

			data, err := json.MarshalIndent(plan, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal plan: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(data))
			return nil
		},
	}
	return cmd
}
