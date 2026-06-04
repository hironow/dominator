package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

// newRunCommand redirects operators away from the retired Go-owned k6
// execution + NFR judging loop. Judging now happens in a Claude Code
// session via mcp-k6 + the
// /nfr-judge skill, which records verdicts through dominator's MCP tools.
func newRunCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "run [path]",
		Short: "Retired: use Claude Code + mcp-k6 + /nfr-judge",
		Long: `Retired by the jun15 MCP pivot (2026-06-15 credit-pool split).

k6 load-test execution + NFR judging are no longer driven by a headless
'claude -p' loop. From a Claude Code session with the mcp-k6 server
attached, use the /nfr-judge skill: it runs k6 via mcp-k6 and records the
verdict through dominator's MCP tools (get_nfr, record_result). Start the
data-plane server with 'dominator mcp'.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, _ []string) error {
			return errors.New("dominator run is retired (jun15 MCP pivot): run k6 via Claude Code with mcp-k6 attached and the /nfr-judge skill; see 'dominator mcp'")
		},
	}
}
