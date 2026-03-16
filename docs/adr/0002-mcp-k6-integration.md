# 0002. mcp-k6 Integration via Claude Code

**Date:** 2026-03-16
**Status:** Accepted

## Context

dominator previously invoked the k6 binary directly via `exec.Command("k6", ...)` to
run load tests. This required users to install k6 separately and created a tight
coupling between dominator and the k6 CLI interface.

The mcp-k6 MCP server provides k6 functionality (run_script, validate_script,
get_documentation) as MCP tools accessible through Claude Code. Since dominator
already depends on Claude Code for script generation, routing k6 execution through
Claude Code + mcp-k6 unifies the external dependency model.

## Decision

Replace direct k6 binary invocation with Claude Code + mcp-k6 MCP tools.

- dominator builds prompts instructing Claude Code to use mcp-k6 tools
- Claude Code stream-json output is parsed to extract tool_result events
- k6 metrics are machine-parsed from tool_result JSON, not from LLM text
- mcp-k6 is configured in the user's global Claude Code MCP settings,
  not in the repository's `.mcp.json`
- The `doctor` command checks mcp-k6 availability instead of k6 binary presence
- A new `validate` command uses mcp-k6's validate_script tool
- The `generate` command instructs Claude Code to use mcp-k6's get_documentation
  and validate_script tools

## Consequences

### Positive

- Single external dependency (Claude Code) instead of two (Claude Code + k6)
- k6 script validation available as a first-class command
- Script generation benefits from mcp-k6's API documentation lookup
- Users do not need to install k6 locally

### Negative

- k6 execution is mediated by an LLM, adding latency and cost
- Metrics extraction depends on Claude Code's stream-json format stability
- mcp-k6 must be configured in the user's global Claude Code environment

### Neutral

- ParseK6Summary remains unchanged (same JSON format regardless of source)
- The port.K6Runner interface gains a Validate method
