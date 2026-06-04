package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/session"
	"github.com/hironow/dominator/internal/usecase"
)

// newMCPCommand exposes `dominator mcp` as a stdio MCP server entry
// point for the refs/issues/0027 jun15 MCP pivot Phase 2c. A claude
// code interactive session loads this binary via --mcp-config and
// calls dominator tools from inside the human-initiated subscription
// quota.
//
// Exposes dominator.ping + dominator.get_nfr (reads NFR thresholds
// from .pass/config.yaml) + dominator.record_result (persists an
// EventJudgmentRecorded event to the .pass event store when an
// emitter is wired).
//
// Note: dominator uses mcp-k6 (an external MCP server) for k6 load
// test execution. That MCP server is attached to the Claude Code
// session directly (= `claude mcp add k6 -- mcp-k6 --transport
// stdio`); `dominator mcp` is a separate server exposing dominator's
// own data plane (NFR config + run history).
func newMCPCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Run dominator as an MCP server over stdio (NFR data plane: get_nfr + record_result)",
		Long: `Start a Model Context Protocol server reading JSON-RPC 2.0
messages on stdin and writing responses on stdout.

Designed for embedding in a Claude Code interactive session via
--mcp-config so inference stays on the session's subscription quota
rather than crossing into the Agent SDK credit pool that gates
'claude -p' from 2026-06-15.

Exposes dominator.ping, dominator.get_nfr (NFR thresholds read from
.pass/config.yaml), and dominator.record_result (persists a verdict
as an EventJudgmentRecorded event to the .pass event store when an
emitter is wired; preview-only otherwise).

This MCP server exposes dominator's data plane (NFR config + run
history). For k6 load test execution, attach the external mcp-k6
server separately to the same Claude Code session.`,
		Example: `  # Launch Claude Code with the dominator MCP server attached
  claude --mcp-config '{"dominator":{"command":"dominator","args":["mcp"]}}'

  # Pipe a tools/list request manually (for debugging)
  echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | dominator mcp`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			passDir := filepath.Join(cwd, domain.StateDir)
			logger := loggerFrom(cmd)
			srv := session.NewMCPServer(cmd.InOrStdin(), cmd.OutOrStdout(), logger).WithPassDir(passDir)

			// Wire the judgment emitter so dominator.record_result persists
			// an EventJudgmentRecorded to the event store (refs/issues/0027
			// Phase 4 follow-up, ADR 0005). LoadConfig falls back to
			// DefaultConfig when config.yaml is absent — that is fine here:
			// RecordJudgment does not consult the NFR config (the session
			// judged externally via mcp-k6, only the verdict is recorded).
			// Only a malformed config.yaml disables persistence, falling
			// back to preview-only.
			cfg, cfgErr := session.LoadConfig(filepath.Join(passDir, "config.yaml"))
			if cfgErr != nil {
				logger.Warn("record_result persistence disabled (config load failed): %v", cfgErr)
			} else {
				agg := domain.NewJudgeAggregate(cfg)
				store := session.NewEventStore(passDir, logger)
				srv = srv.WithEmitter(usecase.NewJudgmentEventEmitter(agg, store, logger))
			}

			return srv.Serve(cmd.Context())
		},
	}
}
