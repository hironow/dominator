## dominator approve

Approve an execution plan

### Synopsis

Approve a previously created execution plan for load testing.

The --plan-id flag specifies which plan to approve. The approved plan JSON
is output to stdout.

```
dominator approve [path] [flags]
```

### Examples

```
  # Approve a plan by ID
  dominator approve --plan-id abc123

  # Approve in a specific directory
  dominator approve --plan-id abc123 /path/to/project
```

### Options

```
  -h, --help             help for approve
      --plan-id string   Plan ID to approve (required)
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
