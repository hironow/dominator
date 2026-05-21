package cmd

import (
	"github.com/spf13/cobra"

	"github.com/hironow/dominator/internal/session"
)

// newMCPCommand exposes `dominator mcp` as a stdio MCP server entry
// point for the refs/issues/0027 jun15 MCP pivot Phase 2c. A claude
// code interactive session loads this binary via --mcp-config and
// calls dominator tools from inside the human-initiated subscription
// quota.
//
// Phase 2c MVP exposes dominator.ping + 2 stubs (get_nfr,
// record_result). Real tool wiring against the NFR config + k6
// event sourcing lands in subsequent commits on feat/jun15-mcp-pivot.
//
// Note: dominator uses mcp-k6 (an external MCP server) for k6 load
// test execution. That MCP server is attached to the claude code
// session directly (= `claude mcp add k6 -- mcp-k6 --transport
// stdio`); `dominator mcp` is a separate server exposing dominator's
// own data plane (NFR config + run history).
func newMCPCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Run dominator as an MCP server over stdio (refs/issues/0027 Phase 2c MVP)",
		Long: `Start a Model Context Protocol server reading JSON-RPC 2.0
messages on stdin and writing responses on stdout.

Designed for embedding in a claude code interactive session via
--mcp-config so inference stays on the session's subscription quota
rather than crossing into the Agent SDK credit pool that gates
'claude -p' from 2026-06-15.

Phase 2c MVP scope: dominator.ping + 2 stubs (dominator.get_nfr,
dominator.record_result). Real wiring against the NFR config and
k6 run history ships in subsequent commits on the
feat/jun15-mcp-pivot branch.

This MCP server exposes dominator's data plane (NFR config + run
history). For k6 load test execution, attach the external mcp-k6
server separately to the same claude code session.`,
		Example: `  # Launch claude code with the dominator MCP server attached
  claude --mcp-config '{"dominator":{"command":"dominator","args":["mcp"]}}'

  # Pipe a tools/list request manually (for debugging)
  echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | dominator mcp`,
		RunE: func(cmd *cobra.Command, args []string) error {
			srv := session.NewMCPServer(cmd.InOrStdin(), cmd.OutOrStdout(), nil)
			return srv.Serve(cmd.Context())
		},
	}
}
