// fake-claude is a test double for the Claude Code CLI.
//
// It mimics the subset of Claude CLI behaviour that dominator relies on:
//   - Responds to --version with a fake version string.
//   - Responds to `mcp list` with fake MCP server entries (including mcp-k6).
//   - Reads a prompt from stdin and returns canned responses.
//   - Accepts (and ignores) the flags dominator passes: --model, --output-format, --print, --verbose, etc.
//
// Install as /usr/local/bin/claude inside an E2E Docker container.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	for _, a := range os.Args[1:] {
		if a == "--version" || a == "-v" {
			fmt.Println("fake-claude 0.0.0-test")
			return
		}
	}

	// Handle `mcp list` subcommand (used by doctor's CheckMCPK6).
	if len(os.Args) >= 3 && os.Args[1] == "mcp" && os.Args[2] == "list" {
		fmt.Println("  linear        ✓  connected")
		fmt.Println("  k6            ✓  connected")
		return
	}

	// Read prompt from -p flag or stdin
	text := extractPrompt(os.Args[1:])
	if text == "" {
		prompt, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fake-claude: read stdin: %v\n", err)
			os.Exit(1)
		}
		text = string(prompt)
	}

	outputFormat := extractOutputFormat(os.Args[1:])

	// Match prompt content to a fixture.
	for _, f := range fixtures {
		if f.match(text) {
			emitResponse(f.content, outputFormat)
			return
		}
	}

	// Default: return a generic success response.
	emitResponse(defaultResponse, outputFormat)
}

// extractPrompt finds -p <prompt> in args.
func extractPrompt(args []string) string {
	for i, arg := range args {
		if arg == "-p" && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

func extractOutputFormat(args []string) string {
	for i, arg := range args {
		if arg == "--output-format" && i+1 < len(args) {
			return args[i+1]
		}
	}
	return "text"
}

func emitResponse(content, outputFormat string) {
	if outputFormat == "stream-json" {
		fmt.Fprint(os.Stdout, wrapStreamJSON(content))
	} else {
		fmt.Fprint(os.Stdout, content)
	}
}

func wrapStreamJSON(body string) string {
	escaped, _ := json.Marshal(body)
	escapedStr := string(escaped)

	initLine := `{"type":"system","subtype":"init","session_id":"fake-session"}`
	assistantLine := fmt.Sprintf(`{"type":"assistant","session_id":"fake-session","message":{"id":"msg_fake","type":"message","role":"assistant","content":[{"type":"text","text":%s}],"model":"claude-opus-4-6","stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":100,"output_tokens":50}},"parent_tool_use_id":null}`, escapedStr)
	resultLine := fmt.Sprintf(`{"type":"result","subtype":"success","session_id":"fake-session","result":%s,"is_error":false,"num_turns":1,"duration_ms":1000,"duration_api_ms":900,"total_cost_usd":0.01,"usage":{"input_tokens":100,"output_tokens":50},"stop_reason":"end_turn","uuid":"fake-uuid"}`, escapedStr)

	return initLine + "\n" + assistantLine + "\n" + resultLine + "\n"
}

type fixture struct {
	keyword string
	match   func(prompt string) bool
	content string
}

var fixtures = []fixture{
	{
		keyword: "validate",
		match:   func(p string) bool { return strings.Contains(p, "validate_script") },
		content: validateResponse,
	},
	{
		keyword: "run k6",
		match:   func(p string) bool { return strings.Contains(p, "run_script") },
		content: runK6Response,
	},
}

var defaultResponse = strings.TrimSpace(`
{
  "success": true,
  "message": "Operation completed successfully"
}
`)

var validateResponse = strings.TrimSpace(`
{
  "valid": true,
  "exit_code": 0,
  "stdout": "ok",
  "stderr": "",
  "error": ""
}
`)

var runK6Response = strings.TrimSpace(`
{
  "success": true,
  "exit_code": 0,
  "metrics": {
    "http_req_duration": {"avg": 50.0, "p95": 150.0},
    "http_reqs": {"count": 1000, "rate": 100.0},
    "http_req_failed": {"rate": 0.005}
  },
  "summary": "1000 requests, 0.5% errors, p95=150ms",
  "stdout": "",
  "stderr": "",
  "error": ""
}
`)
