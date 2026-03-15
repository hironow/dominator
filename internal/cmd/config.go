package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/session"
	"github.com/spf13/cobra"
)

func newConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage dominator configuration",
		Long:  `View or modify the dominator configuration stored in .pass/config.yaml.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("no subcommand specified. Run 'dominator config help' for usage")
		},
	}
	cmd.AddCommand(
		newConfigShowCommand(),
		newConfigSetCommand(),
	)
	return cmd
}

func newConfigShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show [path]",
		Short: "Show current configuration",
		Long:  `Display the current dominator configuration as YAML to stdout.`,
		Example: `  # Show config for current directory
  dominator config show

  # Show config for a specific project
  dominator config show /path/to/project`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := resolveTargetDir(args)
			if err != nil {
				return err
			}
			if err := session.ValidateStateDir(repoRoot); err != nil {
				return err
			}
			stateDir := filepath.Join(repoRoot, domain.StateDir)
			data, err := session.ShowConfig(stateDir)
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), string(data))
			return nil
		},
	}
}

func newConfigSetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value> [path]",
		Short: "Set a configuration value",
		Long: `Set a single configuration key to a new value.

Supported keys:
  target.url, target.protocol, target.spec, target.docs,
  nfr.performance.p95_latency_ms, nfr.performance.error_rate_percent,
  nfr.reliability.success_rate_percent, nfr.scalability.target_rps,
  load.vus, load.duration, load.ramp_up,
  approval.required, lang, claude_cmd, model, timeout_sec`,
		Example: `  # Set target URL
  dominator config set target.url https://example.com

  # Set virtual users
  dominator config set load.vus 50

  # Set config in a specific project directory
  dominator config set lang en /path/to/project`,
		Args: cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]

			var pathArgs []string
			if len(args) == 3 {
				pathArgs = []string{args[2]}
			}

			repoRoot, err := resolveTargetDir(pathArgs)
			if err != nil {
				return err
			}
			if err := session.ValidateStateDir(repoRoot); err != nil {
				return err
			}
			stateDir := filepath.Join(repoRoot, domain.StateDir)
			if err := session.UpdateConfig(stateDir, key, value); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "  Set %s = %s\n", key, value)
			return nil
		},
	}
}
