## dominator approve

Retired: judge NFRs via Claude Code + mcp-k6 + /nfr-judge

### Synopsis

Retired by the jun15 MCP pivot (refs issue 0034).

Plan approval gated the retired 'dominator run' loop. The human-in-the-
loop control it provided lives in the Claude Code session now: the human
invokes /nfr-judge, and the session records the verdict via record_result.

```
dominator approve [path] [flags]
```

### Options

```
  -h, --help             help for approve
      --plan-id string   retired flag (plans are no longer staged)
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

