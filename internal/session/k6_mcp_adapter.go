package session

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strconv"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/platform"
	"github.com/hironow/dominator/internal/usecase/port"
)

// Compile-time check that K6MCPAdapter implements port.K6Runner.
var _ port.K6Runner = (*K6MCPAdapter)(nil)

// K6MCPAdapter executes k6 via Claude Code + mcp-k6 MCP tools.
// Claude Code is the intermediary -- it calls mcp-k6's run_script tool.
// We extract the tool_result from Claude's stream-json output to get
// the actual k6 metrics (not LLM text interpretation).
type K6MCPAdapter struct {
	ClaudeCmd  string
	Model      string
	TimeoutSec int
	Logger     domain.Logger
}

// Run executes a k6 load test via Claude Code + mcp-k6 run_script tool.
// It parses the stream-json output to find tool_result containing k6 metrics.
func (a *K6MCPAdapter) Run(ctx context.Context, scriptPath string, load domain.LoadConfig, stderrW io.Writer) (domain.K6Results, error) {
	prompt := buildK6RunPrompt(scriptPath, load)

	args := []string{
		"--print",
		"--verbose",
		"--output-format", "stream-json",
		"--model", a.Model,
		"-p", prompt,
	}

	a.Logger.Info("Running k6 via mcp-k6: %s (vus=%d, duration=%s)", scriptPath, load.VUs, load.Duration)

	// nosemgrep: go.lang.security.audit.dangerous-exec-command.dangerous-exec-command, lod-excessive-dot-chain -- ClaudeCmd comes from trusted configuration (config.yaml), not user input [permanent]
	cmd := exec.CommandContext(ctx, a.ClaudeCmd, args...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = stderrW

	if err := cmd.Run(); err != nil {
		return domain.K6Results{}, fmt.Errorf("claude mcp-k6 run: %w", err)
	}

	// Parse stream-json to find tool_result containing k6 metrics
	metricsJSON, err := extractK6MetricsFromStream(stdout.Bytes())
	if err != nil {
		return domain.K6Results{}, fmt.Errorf("extract k6 metrics from stream: %w", err)
	}

	// Parse metrics using mcp-k6 aware parser (handles both
	// mcp-k6 {success, metrics} wrapper and raw k6 {metrics} format)
	return parseK6MCPRunResult(metricsJSON)
}

// Validate checks a k6 script using Claude Code + mcp-k6 validate_script tool.
func (a *K6MCPAdapter) Validate(ctx context.Context, scriptPath string) error {
	prompt := buildK6ValidatePrompt(scriptPath)

	args := []string{
		"--print",
		"--verbose",
		"--output-format", "stream-json",
		"--model", a.Model,
		"-p", prompt,
	}

	a.Logger.Info("Validating k6 script via mcp-k6: %s", scriptPath)

	// nosemgrep: go.lang.security.audit.dangerous-exec-command.dangerous-exec-command, lod-excessive-dot-chain -- ClaudeCmd comes from trusted configuration (config.yaml), not user input [permanent]
	cmd := exec.CommandContext(ctx, a.ClaudeCmd, args...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = io.Discard

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("claude mcp-k6 validate: %w", err)
	}

	// Check for validation errors in the stream output
	return checkValidationResult(stdout.Bytes())
}

// buildK6RunPrompt constructs the prompt for mcp-k6 run_script.
func buildK6RunPrompt(scriptPath string, load domain.LoadConfig) string {
	return fmt.Sprintf(`Use the mcp-k6 run_script tool to execute the k6 load test script at %q with %s VUs for %s duration. Return ONLY the raw metrics JSON from the tool result. Do not interpret or summarize the results.`,
		scriptPath, strconv.Itoa(load.VUs), load.Duration)
}

// buildK6ValidatePrompt constructs the prompt for mcp-k6 validate_script.
func buildK6ValidatePrompt(scriptPath string) string {
	return fmt.Sprintf(`Use the mcp-k6 validate_script tool to validate the k6 load test script at %q. Report whether the script is valid or invalid. If invalid, include the error details.`,
		scriptPath)
}

// extractK6MetricsFromStream parses Claude Code stream-json output
// to find tool_result events containing k6 metrics JSON.
func extractK6MetricsFromStream(data []byte) ([]byte, error) {
	reader := platform.NewStreamReader(bytes.NewReader(data))

	var lastToolResult []byte

	for {
		msg, err := reader.Next()
		if err != nil {
			break
		}

		// Look for assistant messages containing tool_result with k6 metrics
		if msg.Type == "assistant" {
			am, parseErr := msg.ParseAssistantMessage()
			if parseErr != nil || am == nil {
				continue
			}
			for _, block := range am.Content {
				if block.Type == "tool_result" && len(block.Input) > 0 {
					lastToolResult = block.Input
				}
			}
		}

		// Also check for tool_result type messages at the top level
		if msg.Type == "tool_result" && msg.Message != nil {
			lastToolResult = msg.Message
		}

		// Check result messages for embedded tool results
		if msg.Type == "result" && msg.Result != "" {
			// Try to extract JSON metrics from the result text
			metricsJSON := extractJSONFromText(msg.Result)
			if metricsJSON != nil {
				lastToolResult = metricsJSON
			}
		}
	}

	if lastToolResult == nil {
		return nil, fmt.Errorf("no k6 metrics found in Claude Code stream output")
	}

	return lastToolResult, nil
}

// extractJSONFromText attempts to find a JSON object containing k6 metrics
// within a text string. Supports both k6 native format ({metrics: ...})
// and mcp-k6 run_script format ({success, metrics, summary, ...}).
func extractJSONFromText(text string) []byte {
	start := -1
	depth := 0
	for i, c := range text {
		switch c {
		case '{':
			if depth == 0 {
				start = i
			}
			depth++
		case '}':
			depth--
			if depth == 0 && start >= 0 {
				candidate := []byte(text[start : i+1])
				var obj map[string]json.RawMessage
				if json.Unmarshal(candidate, &obj) == nil {
					// k6 native format: {metrics: {http_req_duration: ...}}
					if _, hasMetrics := obj["metrics"]; hasMetrics {
						return candidate
					}
					// mcp-k6 run_script format: {success, metrics, summary}
					if _, hasSuccess := obj["success"]; hasSuccess {
						return candidate
					}
				}
				start = -1
			}
		}
	}
	return nil
}

// parseK6MCPRunResult extracts K6Results from mcp-k6 run_script response.
// Handles both the mcp-k6 wrapper format and raw k6 summary format.
// Returns error if the run itself failed (success=false or non-zero exit code).
func parseK6MCPRunResult(data []byte) (domain.K6Results, error) {
	// Try mcp-k6 run_script format first: {success, metrics, ...}
	var runResult mcpK6RunResult
	if err := json.Unmarshal(data, &runResult); err == nil {
		// Check for run failure before parsing metrics
		if !runResult.Success {
			errMsg := runResult.Error
			if errMsg == "" {
				errMsg = runResult.Stderr
			}
			if errMsg == "" {
				errMsg = fmt.Sprintf("k6 run failed with exit code %d", runResult.ExitCode)
			}
			return domain.K6Results{}, fmt.Errorf("mcp-k6 run_script failed: %s", errMsg)
		}
		if runResult.Metrics != nil {
			return ParseK6Summary(fmt.Appendf(nil, `{"metrics":%s}`, string(runResult.Metrics)))
		}
		// success=true but no metrics — fall through to raw format
	}

	// Fall back to raw k6 summary format: {metrics: {http_req_duration: ...}}
	return ParseK6Summary(data)
}

// mcpK6ValidateResult represents the response from mcp-k6 validate_script tool.
type mcpK6ValidateResult struct {
	Valid    bool   `json:"valid"`
	ExitCode int    `json:"exit_code"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	Error    string `json:"error"`
}

// mcpK6RunResult represents the response from mcp-k6 run_script tool.
type mcpK6RunResult struct {
	Success  bool            `json:"success"`
	ExitCode int             `json:"exit_code"`
	Metrics  json.RawMessage `json:"metrics"`
	Summary  string          `json:"summary"`
	Stdout   string          `json:"stdout"`
	Stderr   string          `json:"stderr"`
	Error    string          `json:"error"`
}

// checkValidationResult checks stream-json output for validation errors.
// Parses mcp-k6 validate_script response: {valid, exit_code, stderr, error}.
// Returns error if validation failed OR if no explicit validation result was found
// (fail-closed: absence of proof of validity is treated as invalid).
func checkValidationResult(data []byte) error {
	reader := platform.NewStreamReader(bytes.NewReader(data))

	for {
		msg, err := reader.Next()
		if err != nil {
			break
		}

		// Check for Claude Code error
		if msg.Type == "result" && msg.IsError {
			return fmt.Errorf("k6 script validation failed: %s", msg.Result)
		}

		// Parse mcp-k6 validate_script response from result text
		if msg.Type == "result" && msg.Result != "" {
			resultJSON := extractJSONFromResultText(msg.Result)
			if resultJSON != nil {
				var vr mcpK6ValidateResult
				if json.Unmarshal(resultJSON, &vr) == nil {
					if !vr.Valid {
						return fmt.Errorf("k6 script validation failed (exit %d): %s", vr.ExitCode, validationErrorMessage(vr))
					}
					return nil // explicitly valid
				}
			}
		}

		// Also check tool_result blocks in assistant messages
		if msg.Type == "assistant" {
			am, parseErr := msg.ParseAssistantMessage()
			if parseErr != nil || am == nil {
				continue
			}
			for _, block := range am.Content {
				if block.Type == "tool_result" && len(block.Input) > 0 {
					var vr mcpK6ValidateResult
					if json.Unmarshal(block.Input, &vr) == nil {
						if !vr.Valid {
							return fmt.Errorf("k6 script validation failed (exit %d): %s", vr.ExitCode, validationErrorMessage(vr))
						}
						return nil // explicitly valid
					}
				}
			}
		}
	}

	// Fail-closed: reaching here means no validate_script result was found
	// (each successful match returns directly inside the loop).
	return fmt.Errorf("k6 script validation inconclusive: no validate_script result found in Claude Code output (is mcp-k6 configured?)")
}

// extractJSONFromResultText finds the first JSON object in a text string.
func extractJSONFromResultText(text string) []byte {
	start := -1
	depth := 0
	for i, c := range text {
		switch c {
		case '{':
			if depth == 0 {
				start = i
			}
			depth++
		case '}':
			depth--
			if depth == 0 && start >= 0 {
				candidate := []byte(text[start : i+1])
				var obj map[string]json.RawMessage
				if json.Unmarshal(candidate, &obj) == nil {
					return candidate
				}
				start = -1
			}
		}
	}
	return nil
}

// validationErrorMessage builds an error message from mcp-k6 validate result,
// preferring stderr over error field, with a default fallback.
func validationErrorMessage(vr mcpK6ValidateResult) string {
	if vr.Stderr != "" {
		return vr.Stderr
	}
	if vr.Error != "" {
		return vr.Error
	}
	return "script is invalid (no details available)"
}
