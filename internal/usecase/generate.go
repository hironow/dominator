package usecase

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/usecase/port"
)

// RunGenerate orchestrates k6 script generation from an API spec via Claude Code.
func RunGenerate(
	ctx context.Context,
	cmd domain.GenerateCommand,
	specReader port.SpecReader,
	claudeRunner port.ClaudeRunner,
	eventStore port.EventStore,
	scriptWriter port.ScriptWriter,
	logger domain.Logger,
	stderrW io.Writer,
) (string, error) {
	// 1. Fetch spec
	specURL := cmd.SpecURL().String()
	logger.Info("Fetching spec from %s ...", specURL)
	specContent, err := specReader.Fetch(ctx, specURL)
	if err != nil {
		failEvent, evErr := domain.NewEvent(domain.EventGenerationFailed, domain.GenerationFailedData{
			SpecURL: specURL,
			Reason:  err.Error(),
		}, time.Now())
		if evErr == nil {
			eventStore.Append(failEvent)
		}
		return "", fmt.Errorf("fetch spec: %w", err)
	}

	// 2. Build prompt for Claude Code
	protocol := cmd.Protocol().String()
	prompt := buildK6Prompt(string(specContent), protocol)

	// 3. Generate k6 script via Claude Code
	logger.Info("Generating k6 script via Claude Code ...")
	result, err := claudeRunner.Run(ctx, prompt, stderrW)
	if err != nil {
		failEvent, evErr := domain.NewEvent(domain.EventGenerationFailed, domain.GenerationFailedData{
			SpecURL: specURL,
			Reason:  err.Error(),
		}, time.Now())
		if evErr == nil {
			eventStore.Append(failEvent)
		}
		return "", fmt.Errorf("claude generate: %w", err)
	}

	// 4. Extract script content and write
	scriptContent := extractScriptContent(result)
	scriptName := deriveScriptName(specURL, protocol)
	scriptPath, err := scriptWriter.Write(scriptName, scriptContent)
	if err != nil {
		return "", fmt.Errorf("write script: %w", err)
	}

	// 5. Record event
	successEvent, evErr := domain.NewEvent(domain.EventScriptGenerated, domain.ScriptGeneratedData{
		SpecURL:    specURL,
		Protocol:   protocol,
		ScriptPath: scriptPath,
	}, time.Now())
	if evErr == nil {
		eventStore.Append(successEvent)
	}

	return scriptPath, nil
}

func buildK6Prompt(specContent, protocol string) string {
	base := buildK6PromptForProtocol(specContent, protocol)
	return buildK6PromptWithMCPK6(base)
}

// buildK6PromptForProtocol returns a protocol-specific prompt for k6 script generation.
func buildK6PromptForProtocol(specOrDocs, protocol string) string {
	switch protocol {
	case "json-rpc":
		return fmt.Sprintf(`You are a k6 load testing expert. Generate a k6 load test script for the following JSON-RPC API specification.

Requirements:
- Use k6 JavaScript/TypeScript syntax
- Include proper imports (import http from 'k6/http', import { check, sleep } from 'k6')
- Define reasonable default options (vus: 10, duration: '30s')
- All requests should be HTTP POST to the JSON-RPC endpoint
- Use proper JSON-RPC 2.0 request format: {"jsonrpc": "2.0", "method": "...", "params": {...}, "id": N}
- Test all methods defined in the spec
- Validate JSON-RPC response structure (check for result/error fields)
- Add appropriate sleep between requests
- Output ONLY the k6 script code, no explanations

API Specification:
%s`, specOrDocs)

	case "ws-json-rpc":
		return fmt.Sprintf(`You are a k6 load testing expert. Generate a k6 load test script for the following WebSocket JSON-RPC API specification.

Requirements:
- Use k6 JavaScript/TypeScript syntax
- Include proper imports (import ws from 'k6/ws', import { check } from 'k6')
- Define reasonable default options (vus: 10, duration: '30s')
- Use the k6/ws module to establish WebSocket connections
- Send JSON-RPC 2.0 messages over WebSocket: {"jsonrpc": "2.0", "method": "...", "params": {...}, "id": N}
- Handle WebSocket events (open, message, close, error)
- Validate JSON-RPC response structure in message handlers
- Test all methods defined in the spec
- Output ONLY the k6 script code, no explanations

API Specification:
%s`, specOrDocs)

	case "http":
		return fmt.Sprintf(`You are a k6 load testing expert. Generate a k6 load test script based on the following HTTP API documentation.

Requirements:
- Use k6 JavaScript/TypeScript syntax
- Include proper imports (import http from 'k6/http', import { check, sleep } from 'k6')
- Define reasonable default options (vus: 10, duration: '30s')
- Infer API endpoints, methods, and expected responses from the documentation
- Test all discoverable endpoints
- Include response status checks
- Add appropriate sleep between requests
- Handle common HTTP methods (GET, POST, PUT, DELETE) as described in the docs
- Output ONLY the k6 script code, no explanations

API Documentation:
%s`, specOrDocs)

	default: // openapi
		return fmt.Sprintf(`You are a k6 load testing expert. Generate a k6 load test script for the following %s API specification.

Requirements:
- Use k6 JavaScript/TypeScript syntax
- Include proper imports (import http from 'k6/http', import { check, sleep } from 'k6')
- Define reasonable default options (vus: 10, duration: '30s')
- Test all endpoints defined in the spec
- Include response status checks
- Add appropriate sleep between requests
- Output ONLY the k6 script code, no explanations

API Specification:
%s`, protocol, specOrDocs)
	}
}

// buildK6PromptWithMCPK6 wraps a k6 generation prompt with mcp-k6 tool instructions.
// It instructs Claude Code to use mcp-k6's get_documentation tool before writing
// the script and validate_script tool after generating it.
func buildK6PromptWithMCPK6(basePrompt string) string {
	return fmt.Sprintf(`Before writing the script, use the mcp-k6 get_documentation tool to look up relevant k6 API documentation for the modules you will use.

%s

After generating the script, use the mcp-k6 validate_script tool to verify the generated script is valid k6 code. If validation fails, fix the issues and validate again.`, basePrompt)
}

func extractScriptContent(claudeOutput string) string {
	// Try to extract code block content
	// Look for ```javascript or ```typescript blocks
	lines := strings.Split(claudeOutput, "\n")
	var inCodeBlock bool
	var scriptLines []string

	for _, line := range lines {
		if strings.HasPrefix(line, "```javascript") || strings.HasPrefix(line, "```typescript") || strings.HasPrefix(line, "```js") {
			inCodeBlock = true
			continue
		}
		if inCodeBlock && strings.HasPrefix(line, "```") {
			break
		}
		if inCodeBlock {
			scriptLines = append(scriptLines, line)
		}
	}

	if len(scriptLines) > 0 {
		return strings.Join(scriptLines, "\n") + "\n"
	}
	// If no code block found, return raw output
	return claudeOutput
}

func deriveScriptName(specURL, protocol string) string {
	// Extract a meaningful name from the URL
	base := filepath.Base(specURL)
	base = strings.TrimSuffix(base, ".json")
	base = strings.TrimSuffix(base, ".yaml")
	base = strings.TrimSuffix(base, ".yml")
	if base == "" || base == "." || base == "/" {
		base = protocol + "-load-test"
	}
	return base + ".js"
}
