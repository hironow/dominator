package session

import (
	"context"
	"io"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/usecase/port"
)

// Compile-time check that K6MCPAdapter implements port.K6Runner.
var _ port.K6Runner = (*K6MCPAdapter)(nil)

// K6MCPAdapter previously routed k6 load test execution through
// Claude Code + mcp-k6 MCP tools by invoking `claude --print` as a
// subprocess so that Claude could call the mcp-k6 run_script /
// validate_script tools as an intermediary. Post jun15 MCP pivot
// (refs/issues/0027 Phase 2c + 0028 §4.3 residue cleanup), that
// headless invocation is forbidden.
//
// To execute k6 load tests post pivot, attach the mcp-k6 server
// directly to a human-initiated claude code session and invoke the
// /nfr-judge slash command (defined in
// plugins/dominator/skills/nfr-judge/SKILL.md). Example:
//
//	claude --plugin-dir ./plugins/dominator \
//	       --mcp-config '{"dominator":{"command":"dominator","args":["mcp"]},"mcp-k6":{"command":"mcp-k6"}}'
//
// All Run / Validate calls on this adapter now short-circuit to
// ErrMCPPivotDeprecated (declared in claude_adapter.go and shared
// across the session package).
//
// The struct retains its config fields so existing composition roots
// (cmd/run.go, cmd/validate.go) that construct it via
// K6MCPAdapter{...} still compile.
type K6MCPAdapter struct {
	ClaudeCmd  string
	Model      string
	TimeoutSec int
	Logger     domain.Logger
}

// Run returns ErrMCPPivotDeprecated. The previous implementation
// invoked `claude --print` to call the mcp-k6 run_script tool via
// Claude Code as an intermediary; that headless invocation path is
// forbidden post jun15 MCP pivot (refs/issues/0027 §5 billing
// boundary, 0028 §4.3 residue cleanup).
func (a *K6MCPAdapter) Run(_ context.Context, _ string, _ domain.LoadConfig, _ io.Writer) (domain.K6Results, error) {
	if a.Logger != nil {
		a.Logger.Warn("dominator: K6MCPAdapter.Run() is deprecated (refs/issues/0027 jun15 MCP pivot, 0028 residue cleanup); use claude code with mcp-k6 attached and the /nfr-judge skill instead.")
	}
	return domain.K6Results{}, ErrMCPPivotDeprecated
}

// Validate returns ErrMCPPivotDeprecated. See Run for context.
func (a *K6MCPAdapter) Validate(_ context.Context, _ string) error {
	if a.Logger != nil {
		a.Logger.Warn("dominator: K6MCPAdapter.Validate() is deprecated (refs/issues/0027 jun15 MCP pivot, 0028 residue cleanup); use claude code with mcp-k6 attached and the /nfr-judge skill instead.")
	}
	return ErrMCPPivotDeprecated
}
