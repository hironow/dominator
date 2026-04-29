package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/platform"
	"github.com/hironow/dominator/internal/session"
)

// newStatusCommand creates the status subcommand that displays operational status.
func newStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status [path]",
		Short: "Show dominator operational status",
		Long: `Display operational status including judgment history, plans,
k6 scripts, and pending d-mail counts.

Output goes to stdout by default (human-readable text).
Use -o json for machine-readable JSON output to stdout.`,
		Example: `  # Show status for current directory
  dominator status

  # Show status for a specific project
  dominator status /path/to/project

  # JSON output for scripting
  dominator status -o json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := resolveTargetDir(args)
			if err != nil {
				return err
			}

			passRoot := filepath.Join(repoRoot, domain.StateDir)
			if _, err := os.Stat(passRoot); errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf(".pass/ not found. Run 'dominator init' first")
			}

			logger := platform.NewLogger(cmd.ErrOrStderr(), false)
			report := session.Status(cmd.Context(), passRoot, logger)

			outputFmt, _ := cmd.Flags().GetString("output") // nosemgrep: error-handling.ignored-error-go,error-handling.ignored-error-short-go -- cobra flag registered statically; GetString cannot fail at runtime [permanent]
			if outputFmt == "json" {
				data, jsonErr := json.Marshal(report)
				if jsonErr != nil {
					return fmt.Errorf("marshal status: %w", jsonErr)
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			// Text output to stdout (human-readable)
			fmt.Fprint(cmd.OutOrStdout(), report.FormatText())
			return nil
		},
	}
}
