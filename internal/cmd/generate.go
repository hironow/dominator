package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

// newGenerateCommand redirects operators away from the retired Go-owned k6
// script generation loop. Generate scripts from a Claude Code session and
// write them under .pass/k6-scripts/.
func newGenerateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "generate [path]",
		Short: "Retired: generate k6 scripts via Claude Code",
		Long: `Retired by the jun15 MCP pivot.

k6 script generation from an API spec is no longer driven by a headless
'claude -p' loop. Generate scripts from a Claude Code session (which has
the spec-reading + authoring tools) and place them under .pass/k6-scripts/.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, _ []string) error {
			return errors.New("dominator generate is retired (jun15 MCP pivot): generate k6 scripts from a Claude Code session and write them to .pass/k6-scripts/")
		},
	}
}
