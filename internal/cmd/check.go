package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

// newCheckCommand redirects operators away from the retired plan-staging
// flow. Plans existed to gate the Go-owned `run` loop; with `run` retired
// by the jun15 MCP pivot, staging them has no consumer (refs issue 0034).
// The /nfr-judge skill in a Claude Code session drives judgment directly
// via mcp-k6 + dominator's MCP tools (get_nfr / get_insights /
// record_result / dmail).
func newCheckCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "check [path]",
		Short: "Retired: judge NFRs via Claude Code + mcp-k6 + /nfr-judge",
		Long: `Retired by the jun15 MCP pivot (refs issue 0034).

Execution plans are no longer staged in the Go CLI: their only consumer
('dominator run') was retired with the pivot. Judge NFR targets from a
Claude Code session via the /nfr-judge skill, which reads thresholds with
get_nfr, runs k6 through mcp-k6, and records the verdict with
record_result.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, _ []string) error {
			return errors.New("dominator check is retired (jun15 MCP pivot): plans have no consumer since 'run' was retired; judge via Claude Code + mcp-k6 + the /nfr-judge skill (see 'dominator mcp')")
		},
	}
}
