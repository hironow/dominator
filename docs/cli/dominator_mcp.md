## dominator mcp

Run dominator as an MCP server over stdio (refs/issues/0027 Phase 2c MVP)

### Synopsis

Start a Model Context Protocol server reading JSON-RPC 2.0
messages on stdin and writing responses on stdout.

Designed for embedding in a claude code interactive session via
--mcp-config so inference stays on the session's subscription quota
rather than crossing into the Agent SDK credit pool that gates
'claude -p' from 2026-06-15.

Phase 2c MVP scope: dominator.ping + 2 stubs (dominator.get_nfr,
dominator.record_result). Real wiring against the NFR config and
k6 run history ships in subsequent commits on the
feat/jun15-mcp-pivot branch.

This MCP server exposes dominator's data plane (NFR config + run
history). For k6 load test execution, attach the external mcp-k6
server separately to the same claude code session.

```
dominator mcp [flags]
```

### Examples

```
  # Launch claude code with the dominator MCP server attached
  claude --mcp-config '{"dominator":{"command":"dominator","args":["mcp"]}}'

  # Pipe a tools/list request manually (for debugging)
  echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | dominator mcp
```

### Options

```
  -h, --help   help for mcp
```

### Options inherited from parent commands

```
  -c, --config string   config file path
  -l, --lang string     output language (ja, en)
      --no-color        Disable colored output (respects NO_COLOR env)
  -o, --output string   Output format: text, json (default "text")
  -q, --quiet           Suppress all stderr output
  -v, --verbose         verbose output
```

### SEE ALSO

* [dominator](dominator.md)	 - NFR Judge for your system

