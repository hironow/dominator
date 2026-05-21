package session_test

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/session"
)

// TestClaudeAdapter_RunReturnsErrMCPPivotDeprecated is the canonical
// post jun15 MCP pivot assertion (refs/issues/0027): the previous
// exec.CommandContext path that drove `claude --print` has been
// removed. This test pins the behavior callers can rely on — every
// invocation short-circuits with session.ErrMCPPivotDeprecated so
// operators are routed to the human-initiated claude code /nfr-judge
// skill instead.
func TestClaudeAdapter_RunReturnsErrMCPPivotDeprecated(t *testing.T) {
	// given
	adapter := &session.ClaudeAdapter{
		ClaudeCmd:  "claude",
		Model:      "sonnet",
		TimeoutSec: 10,
		Logger:     &domain.NopLogger{},
	}

	// when
	_, err := adapter.Run(context.Background(), "anything", io.Discard)

	// then
	if !errors.Is(err, session.ErrMCPPivotDeprecated) {
		t.Errorf("Run() error = %v, want ErrMCPPivotDeprecated", err)
	}
}
