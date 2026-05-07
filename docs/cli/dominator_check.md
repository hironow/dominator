## dominator check

Create an execution plan from k6 scripts

### Synopsis

Inspect available k6 scripts and create an execution plan for load testing.

Reads configuration and lists k6 scripts in .pass/k6-scripts/. If scripts are
found, a Plan is created and saved to .pass/.run/plans/. The plan JSON is
output to stdout for piping to other tools.

```
dominator check [path] [flags]
```

### Examples

```
  # Check current directory
  dominator check

  # Check a specific project directory
  dominator check /path/to/project
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

* [dominator](dominator.md)  - NFR Judge for your system
