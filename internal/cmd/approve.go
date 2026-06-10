package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

// newApproveCommand redirects operators away from the retired
// plan-approval flow (see newCheckCommand for the full rationale; refs
// issue 0034).
func newApproveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approve [path]",
		Short: "Retired: judge NFRs via Claude Code + mcp-k6 + /nfr-judge",
		Long: `Retired by the jun15 MCP pivot (refs issue 0034).

Plan approval gated the retired 'dominator run' loop. The human-in-the-
loop control it provided lives in the Claude Code session now: the human
invokes /nfr-judge, and the session records the verdict via record_result.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, _ []string) error {
			return errors.New("dominator approve is retired (jun15 MCP pivot): plans have no consumer since 'run' was retired; judge via Claude Code + mcp-k6 + the /nfr-judge skill (see 'dominator mcp')")
		},
	}
	// --plan-id is kept registered so old invocations fail with the
	// retirement message instead of a flag-parse error.
	cmd.Flags().String("plan-id", "", "retired flag (plans are no longer staged)")
	return cmd
}
