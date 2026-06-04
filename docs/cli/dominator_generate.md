## dominator generate

Retired: generate k6 scripts via Claude Code

### Synopsis

Retired by the jun15 MCP pivot.

k6 script generation from an API spec is no longer driven by a headless
'claude -p' loop. Generate scripts from a Claude Code session (which has
the spec-reading + authoring tools) and place them under .pass/k6-scripts/.

```
dominator generate [path] [flags]
```

### Options

```
  -h, --help   help for generate
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

