package cmd

import (
	"bufio"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/platform"
	"github.com/spf13/cobra"
)

func newArchivePruneCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive-prune [path]",
		Short: "Prune old archived files",
		Long: `Prune archived d-mail files and expired event files.

By default, runs in dry-run mode showing what would be deleted.
Pass --execute to actually remove the files.`,
		Example: `  # Dry-run: list expired files (default 30 days)
  dominator archive-prune

  # Delete expired files (with confirmation)
  dominator archive-prune --execute

  # Delete without confirmation
  dominator archive-prune --execute --yes

  # Custom retention period
  dominator archive-prune --days 7 --execute

  # Rebuild archive index from existing files
  dominator archive-prune --rebuild-index`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			days, _ := cmd.Flags().GetInt("days")
			execute, _ := cmd.Flags().GetBool("execute")
			dryRunExplicit := cmd.Flags().Changed("dry-run")
			yes, _ := cmd.Flags().GetBool("yes")
			logger := platform.NewLogger(cmd.ErrOrStderr(), false)

			if execute && dryRunExplicit {
				return fmt.Errorf("--execute and --dry-run are mutually exclusive")
			}

			repoRoot, err := resolveTargetDir(args)
			if err != nil {
				return err
			}

			rebuildIndex, _ := cmd.Flags().GetBool("rebuild-index")
			if rebuildIndex {
				if execute || dryRunExplicit {
					return fmt.Errorf("--rebuild-index cannot be combined with --execute or --dry-run")
				}
				logger.Info("rebuild-index: not yet implemented")
				return nil
			}

			_, dErr := domain.NewDays(days)
			if dErr != nil {
				return dErr
			}

			passRoot := filepath.Join(repoRoot, domain.StateDir)

			// Placeholder: no archive files to prune yet
			if !execute {
				fmt.Fprintf(cmd.ErrOrStderr(), "No files older than %d days to prune under %s\n", days, passRoot)
				fmt.Fprintln(cmd.ErrOrStderr(), "(dry-run — pass --execute to delete)")
				return nil
			}

			if !yes {
				fmt.Fprintf(cmd.ErrOrStderr(), "\nDelete files? [y/N] ")
				scanner := bufio.NewScanner(cmd.InOrStdin())
				if !scanner.Scan() {
					fmt.Fprintln(cmd.ErrOrStderr(), "Cancelled.")
					return nil
				}
				answer := strings.TrimSpace(scanner.Text())
				if answer != "y" && answer != "Y" {
					fmt.Fprintln(cmd.ErrOrStderr(), "Cancelled.")
					return nil
				}
			}

			fmt.Fprintf(cmd.ErrOrStderr(), "No files to prune under %s\n", passRoot)
			return nil
		},
	}

	cmd.Flags().IntP("days", "d", 30, "Retention days")
	cmd.Flags().BoolP("execute", "x", false, "Execute pruning (default: dry-run)")
	cmd.Flags().BoolP("dry-run", "n", false, "Dry-run mode (default behavior, explicit for scripting)")
	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	cmd.Flags().Bool("rebuild-index", false, "Rebuild archive index from existing files without pruning")

	return cmd
}
