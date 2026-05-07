## dominator validate

Validate k6 scripts via mcp-k6

### Synopsis

Validate k6 scripts in .pass/k6-scripts/ using Claude Code + mcp-k6 validate_script.

Uses the mcp-k6 MCP tool through Claude Code to check script syntax and validity.

Exit codes:
  0 = valid
  1 = invalid or error

```
dominator validate [path] [flags]
```

### Examples

```
  # Validate scripts in current directory
  dominator validate

  # Validate scripts in a specific project
  dominator validate /path/to/project
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

* [dominator](dominator.md)  - NFR Judge for your system
