package session

import (
	"context"
	"errors"
	"io"

	"github.com/hironow/dominator/internal/domain"
)

// ClaudeAdapter implements port.ClaudeRunner. Post jun15 MCP pivot
// (refs/issues/0027), Run returns ErrMCPPivotDeprecated rather than
// executing `claude --print` via exec.CommandContext. LLM inference
// now happens inside a human-initiated claude code interactive
// session driven by the dominator MCP server (`dominator mcp`
// subcommand) plus the /nfr-judge slash command defined in
// plugins/dominator/skills/nfr-judge/SKILL.md.
//
// The struct retains its config fields so call sites (cmd/generate
// composition root) that construct it via ClaudeAdapter{...} still
// compile. Composite roots can wire it in but every Run call
// short-circuits to ErrMCPPivotDeprecated until the MCP-driven
// pipeline replaces this adapter entirely.
type ClaudeAdapter struct {
	ClaudeCmd  string
	Model      string
	TimeoutSec int
	Logger     domain.Logger
}

// ErrMCPPivotDeprecated is returned by ClaudeAdapter.Run now that
// the LLM invocation layer has moved to a human-initiated claude
// code interactive session per refs/issues/0027 (jun15 MCP pivot).
// Callers should surface this error and direct the operator to
// launch claude code with:
//
//	claude --plugin-dir ./plugins/dominator \
//	       --mcp-config '{"dominator":{"command":"dominator","args":["mcp"]},"k6":{"command":"mcp-k6","args":["--transport","stdio"]}}'
//
// then invoke the /nfr-judge slash command.
var ErrMCPPivotDeprecated = errors.New(
	"dominator Go CLI claude adapter deprecated post jun15 MCP pivot: " +
		"use claude code /nfr-judge skill (refs/issues/0027)",
)

// Run returns ErrMCPPivotDeprecated. The previous implementation
// invoked `claude --print` via exec.CommandContext, which is
// forbidden post the jun15 MCP pivot (refs/issues/0027 §5 billing
// boundary).
func (a *ClaudeAdapter) Run(_ context.Context, _ string, _ io.Writer) (string, error) {
	if a.Logger != nil {
		a.Logger.Warn("dominator: ClaudeAdapter.Run() is deprecated (refs/issues/0027 jun15 MCP pivot); use the claude code /nfr-judge skill instead.")
	}
	return "", ErrMCPPivotDeprecated
}
