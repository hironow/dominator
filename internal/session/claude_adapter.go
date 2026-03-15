package session

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/hironow/dominator/internal/domain"
)

// ClaudeAdapter implements port.ClaudeRunner via exec.Command.
type ClaudeAdapter struct {
	ClaudeCmd  string
	Model      string
	TimeoutSec int
	Logger     domain.Logger
}

// Run executes the Claude CLI with the given prompt and writes progress to w.
// Returns the generated text from stdout.
func (a *ClaudeAdapter) Run(ctx context.Context, prompt string, w io.Writer) (string, error) {
	args := []string{
		"--print",
		"--verbose",
		"--output-format", "stream-json",
		"--model", a.Model,
		"-p", prompt,
	}

	// nosemgrep: lod-excessive-dot-chain -- false positive: dot chain is in nosemgrep rule ID below, not in code [permanent]
	cmd := exec.CommandContext(ctx, a.ClaudeCmd, args...) // nosemgrep: go.lang.security.audit.dangerous-exec-command.dangerous-exec-command -- ClaudeCmd comes from trusted configuration (config.yaml), not user input [permanent]
	// stdout captures the result, stderr goes to w for progress
	var stdout strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = w

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("claude: %w", err)
	}
	return stdout.String(), nil
}
