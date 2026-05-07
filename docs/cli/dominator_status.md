## dominator status

Show dominator operational status

### Synopsis

Display operational status including judgment history, plans,
k6 scripts, and pending d-mail counts.

Output goes to stdout by default (human-readable text).
Use -o json for machine-readable JSON output to stdout.

```
dominator status [path] [flags]
```

### Examples

```
  # Show status for current directory
  dominator status

  # Show status for a specific project
  dominator status /path/to/project

  # JSON output for scripting
  dominator status -o json
```

### Options

```
  -h, --help   help for status
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
