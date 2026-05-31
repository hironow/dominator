package cmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/session"
	"github.com/spf13/cobra"
)

func newInboxCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inbox [path]",
		Short: "Process incoming D-Mail messages",
		Long: `Read and process D-Mail files from .pass/inbox/.

Parses YAML frontmatter from each .md file, determines a suggested action
based on the D-Mail kind, archives processed files to .pass/archive/inbox/,
and outputs a JSON array to stdout.

Exit codes:
  0 — D-Mail found and processed
  1 — inbox empty or not initialized`,
		Example: `  # Process inbox for current directory
  dominator inbox

  # Process inbox for a specific project
  dominator inbox /path/to/project`,
		Args: cobra.MaximumNArgs(1),
		RunE: runInbox,
	}
	return cmd
}

func runInbox(cmd *cobra.Command, args []string) error {
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

	reader := &session.InboxReader{
		StateDir: stateDir,
		Logger:   logger,
	}

	entries, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("read inbox: %w", err)
	}

	if len(entries) == 0 {
		// Output empty array and signal empty inbox
		fmt.Fprintln(cmd.OutOrStdout(), "[]")
		return fmt.Errorf("inbox empty")
	}

	// Archive processed files
	for _, entry := range entries {
		if archErr := reader.Archive(entry.Name + ".md"); archErr != nil {
			logger.Warn("failed to archive %s: %v", entry.Name, archErr)
		}
	}

	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	if err := enc.Encode(entries); err != nil {
		return fmt.Errorf("encode inbox entries: %w", err)
	}

	logger.Info("Processed %d D-Mail(s)", len(entries))
	return nil
}
