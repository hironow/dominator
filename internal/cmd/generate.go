package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

// newGenerateCommand is a deprecation stub. k6 script generation from an
// API spec was driven by the headless ClaudeAdapter, removed in the jun15
// MCP pivot. Generate scripts from a claude code session and write them
// under .pass/k6-scripts/.
func newGenerateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "generate [path]",
		Short: "Deprecated (jun15 MCP pivot): generate k6 scripts via claude code",
		Long: `Deprecated by the jun15 MCP pivot.

k6 script generation from an API spec is no longer driven by a headless
'claude -p' loop. Generate scripts from a claude code session (which has
the spec-reading + authoring tools) and place them under .pass/k6-scripts/.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, _ []string) error {
			return errors.New("dominator generate is deprecated (jun15 MCP pivot): generate k6 scripts from a claude code session and write them to .pass/k6-scripts/")
		},
	}
}
