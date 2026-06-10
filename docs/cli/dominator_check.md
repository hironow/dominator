## dominator check

Retired: judge NFRs via Claude Code + mcp-k6 + /nfr-judge

### Synopsis

Retired by the jun15 MCP pivot (refs issue 0034).

Execution plans are no longer staged in the Go CLI: their only consumer
('dominator run') was retired with the pivot. Judge NFR targets from a
Claude Code session via the /nfr-judge skill, which reads thresholds with
get_nfr, runs k6 through mcp-k6, and records the verdict with
record_result.

```
dominator check [path] [flags]
```

### Options

```
  -h, --help   help for check
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

