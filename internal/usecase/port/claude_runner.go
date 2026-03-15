package port

import (
	"context"
	"io"
)

// ClaudeRunner executes the Claude CLI and returns the result text.
type ClaudeRunner interface {
	Run(ctx context.Context, prompt string, w io.Writer) (string, error)
}
