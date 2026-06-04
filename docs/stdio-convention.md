# stdio Convention

Dominator follows the Unix convention of separating machine-readable data from human-readable diagnostics across standard streams.

## Stream Assignment

| Stream | Purpose | Implementation |
|--------|---------|----------------|
| **stdout** | Machine-readable output (JSON, judgment results) | `cmd.OutOrStdout()` |
| **stderr** | Human-readable progress, logs, errors | `platform.Logger` via `cmd.ErrOrStderr()` |
| **stdin** | JSON-RPC input for `dominator mcp`; optional confirmations for selected commands | `cmd.InOrStdin()` / MCP server |

The core dominator library does not read arbitrary prompt text from stdin. The MCP server reads JSON-RPC 2.0 requests from stdin by design, and some CLI commands (such as `archive-prune`) optionally read from stdin for confirmations.

## Cobra Wiring

All cobra subcommands MUST use cobra's stream accessors:

```go
logger := platform.NewLogger(cmd.ErrOrStderr(), verbose)
```

Rules:

- Use `cmd.OutOrStdout()` for data output — never `os.Stdout` directly
- Use `cmd.ErrOrStderr()` for logging — never `os.Stderr` directly
- This enables cobra's `cmd.SetOut()` / `cmd.SetErr()` for testing

### Exceptions

Direct `os.Stderr` is acceptable only where cobra's `cmd` is unavailable:

| Location | Reason |
|----------|--------|
| `cmd/dominator/main.go` | Error handling after `root.ExecuteContext()` returns |
| `internal/tools/docgen/main.go` | Standalone tool outside cobra |

## Pipeline Compatibility

The stream separation ensures correct behavior in Unix pipelines:

```bash
dominator insights | jq '.hue'       # stdout = JSON only
dominator run 2>/dev/null             # suppress stderr logs
dominator status --output json 2>run.log  # split logs to file
```
