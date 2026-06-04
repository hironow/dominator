## dominator validate

Retired: validate k6 scripts via Claude Code + mcp-k6

### Synopsis

Retired by the jun15 MCP pivot.

k6 script validation is no longer driven by a headless 'claude -p' loop.
Validate scripts from a Claude Code session with the mcp-k6 server
attached (mcp-k6 exposes validate_script); see the /nfr-judge skill.

```
dominator validate [path] [flags]
```

### Options

```
  -h, --help   help for validate
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

