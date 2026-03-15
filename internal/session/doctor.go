package session

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/platform"
)

// CheckStateDir verifies that the .pass/ state directory exists.
// If repair is true and the directory is missing, it attempts to create it.
func CheckStateDir(repoRoot string, repair bool) domain.DoctorCheck {
	passRoot := filepath.Join(repoRoot, domain.StateDir)
	info, err := os.Stat(passRoot)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if repair {
				if mkErr := os.MkdirAll(passRoot, 0o755); mkErr != nil {
					return domain.DoctorCheck{
						Name:    "state-dir",
						Status:  domain.CheckFail,
						Message: fmt.Sprintf("repair failed: %v", mkErr),
					}
				}
				return domain.DoctorCheck{
					Name:    "state-dir",
					Status:  domain.CheckFixed,
					Message: fmt.Sprintf("created %s", passRoot),
				}
			}
			return domain.DoctorCheck{
				Name:    "state-dir",
				Status:  domain.CheckFail,
				Message: fmt.Sprintf("%s not found", passRoot),
				Hint:    "run 'dominator init' to create the state directory",
			}
		}
		return domain.DoctorCheck{
			Name:    "state-dir",
			Status:  domain.CheckFail,
			Message: fmt.Sprintf("stat error: %v", err),
		}
	}
	if !info.IsDir() {
		return domain.DoctorCheck{
			Name:    "state-dir",
			Status:  domain.CheckFail,
			Message: fmt.Sprintf("%s exists but is not a directory", passRoot),
		}
	}
	return domain.DoctorCheck{
		Name:    "state-dir",
		Status:  domain.CheckOK,
		Message: fmt.Sprintf("%s exists", passRoot),
	}
}

// CheckConfig verifies that config.yaml is parseable and valid.
func CheckConfig(configPath string) domain.DoctorCheck {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return domain.DoctorCheck{
			Name:    "config-valid",
			Status:  domain.CheckFail,
			Message: fmt.Sprintf("config parse error: %v", err),
			Hint:    "fix config.yaml syntax errors",
		}
	}

	// Check if file exists
	if _, statErr := os.Stat(configPath); errors.Is(statErr, fs.ErrNotExist) {
		return domain.DoctorCheck{
			Name:    "config-valid",
			Status:  domain.CheckWarn,
			Message: "config.yaml not found (using defaults)",
			Hint:    "run 'dominator init' to generate config.yaml",
		}
	}

	errs := domain.ValidateConfig(cfg)
	if len(errs) > 0 {
		return domain.DoctorCheck{
			Name:    "config-valid",
			Status:  domain.CheckFail,
			Message: fmt.Sprintf("config validation errors: %s", errs[0]),
			Hint:    "fix config.yaml values",
		}
	}

	return domain.DoctorCheck{
		Name:    "config-valid",
		Status:  domain.CheckOK,
		Message: "config.yaml is valid",
	}
}

// CheckMCPK6 verifies that mcp-k6 is available via Claude Code.
// It runs Claude Code and checks if mcp-k6 tools appear in the init message.
func CheckMCPK6(claudeCmd string, model string, logger domain.Logger) domain.DoctorCheck {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	args := []string{
		"--print",
		"--verbose",
		"--output-format", "stream-json",
		"--model", model,
		"-p", "List your available MCP tools. Just list the tool names, nothing else.",
	}

	// nosemgrep: go.lang.security.audit.dangerous-exec-command.dangerous-exec-command, lod-excessive-dot-chain -- claudeCmd comes from trusted configuration (config.yaml), not user input [permanent]
	cmd := exec.CommandContext(ctx, claudeCmd, args...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = io.Discard

	if err := cmd.Run(); err != nil {
		return domain.DoctorCheck{
			Name:    "mcp-k6",
			Status:  domain.CheckFail,
			Message: fmt.Sprintf("claude command failed: %v", err),
			Hint:    "ensure Claude Code CLI is installed and configured",
		}
	}

	// Parse stream-json to look for mcp-k6 tools in init or result
	reader := platform.NewStreamReader(bytes.NewReader(stdout.Bytes()))
	mcpK6Found := false

	for {
		msg, err := reader.Next()
		if err != nil {
			break
		}

		// Check init message for MCP servers
		if msg.Type == "system" && msg.Subtype == "init" {
			for _, server := range msg.MCPServers {
				serverName := strings.ToLower(server.Name)
				if strings.Contains(serverName, "k6") {
					mcpK6Found = true
					break
				}
			}
			// Also check tools list
			for _, tool := range msg.Tools {
				if strings.Contains(tool, "run_script") || strings.Contains(tool, "validate_script") {
					mcpK6Found = true
					break
				}
			}
		}

		// Check result text for mcp-k6 tool mentions
		if msg.Type == "result" && msg.Result != "" {
			if strings.Contains(msg.Result, "run_script") || strings.Contains(msg.Result, "validate_script") {
				mcpK6Found = true
			}
		}
	}

	if !mcpK6Found {
		return domain.DoctorCheck{
			Name:    "mcp-k6",
			Status:  domain.CheckFail,
			Message: "mcp-k6 tools not found in Claude Code environment",
			Hint:    "run: claude mcp add k6 -- mcp-k6 --transport stdio",
		}
	}

	return domain.DoctorCheck{
		Name:    "mcp-k6",
		Status:  domain.CheckOK,
		Message: "mcp-k6 available via Claude Code",
	}
}
