package cmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/session"
	"github.com/spf13/cobra"
)

func newValidateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate [path]",
		Short: "Validate k6 scripts via mcp-k6",
		Long: `Validate k6 scripts in .pass/k6-scripts/ using Claude Code + mcp-k6 validate_script.

Uses the mcp-k6 MCP tool through Claude Code to check script syntax and validity.

Exit codes:
  0 = valid
  1 = invalid or error`,
		Example: `  # Validate scripts in current directory
  dominator validate

  # Validate scripts in a specific project
  dominator validate /path/to/project`,
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
			configPath := filepath.Join(stateDir, "config.yaml")

			cfg, err := session.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			logger := loggerFrom(cmd)
			adapter := &session.K6MCPAdapter{
				ClaudeCmd:  cfg.ClaudeCmd,
				Model:      cfg.Model,
				TimeoutSec: cfg.TimeoutSec,
				Logger:     logger,
			}

			// Find scripts to validate
			planStore := &session.PlanStore{StateDir: stateDir}
			scripts, err := planStore.ListScripts()
			if err != nil {
				return fmt.Errorf("list scripts: %w", err)
			}

			if len(scripts) == 0 {
				return fmt.Errorf("no k6 scripts found in %s/k6-scripts/", stateDir)
			}

			type validationResult struct {
				Script string `json:"script"`
				Valid  bool   `json:"valid"`
				Error  string `json:"error,omitempty"`
			}

			var results []validationResult
			hasInvalid := false

			for _, script := range scripts {
				scriptPath := filepath.Join(stateDir, "k6-scripts", script)
				vErr := adapter.Validate(cmd.Context(), scriptPath)
				r := validationResult{
					Script: script,
					Valid:  vErr == nil,
				}
				if vErr != nil {
					r.Error = vErr.Error()
					hasInvalid = true
				}
				results = append(results, r)
			}

			data, marshalErr := json.MarshalIndent(results, "", "  ")
			if marshalErr != nil {
				return fmt.Errorf("marshal validation results: %w", marshalErr)
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(data))

			if hasInvalid {
				return &domain.SilentError{Err: fmt.Errorf("one or more scripts are invalid")}
			}

			return nil
		},
	}

	return cmd
}
