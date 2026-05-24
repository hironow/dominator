package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

// newValidateCommand is a deprecation stub. k6 script validation was
// driven by the headless K6MCPAdapter, removed in the jun15 MCP pivot.
// Validate scripts from a claude code session with mcp-k6 attached
// (mcp-k6 exposes validate_script).
func newValidateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "validate [path]",
		Short: "Deprecated (jun15 MCP pivot): validate k6 scripts via claude code + mcp-k6",
		Long: `Deprecated by the jun15 MCP pivot.

k6 script validation is no longer driven by a headless 'claude -p' loop.
Validate scripts from a claude code session with the mcp-k6 server
attached (mcp-k6 exposes validate_script); see the /nfr-judge skill.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, _ []string) error {
			return errors.New("dominator validate is deprecated (jun15 MCP pivot): validate k6 scripts via claude code with mcp-k6 attached (validate_script)")
		},
	}
}
