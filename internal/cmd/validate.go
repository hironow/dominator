package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

// newValidateCommand redirects operators away from the retired Go-owned k6
// validation loop. Validate scripts from a Claude Code session with mcp-k6
// attached (mcp-k6 exposes validate_script).
func newValidateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "validate [path]",
		Short: "Retired: validate k6 scripts via Claude Code + mcp-k6",
		Long: `Retired by the jun15 MCP pivot.

k6 script validation is no longer driven by a headless 'claude -p' loop.
Validate scripts from a Claude Code session with the mcp-k6 server
attached (mcp-k6 exposes validate_script); see the /nfr-judge skill.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, _ []string) error {
			return errors.New("dominator validate is retired (jun15 MCP pivot): validate k6 scripts via Claude Code with mcp-k6 attached (validate_script)")
		},
	}
}
