## dominator run

Execute k6 load test and judge NFR compliance

### Synopsis

Run loads an approved execution plan, executes the k6 load test,
evaluates NFR thresholds, records events, writes insights, and emits
D-Mail notifications on violation.

Exit codes:
  0 = pass (all NFRs met)
  1 = violation (one or more NFR thresholds exceeded)
  2 = error (plan not found, not approved, k6 failure, etc.)

```
dominator run [path] [flags]
```

### Examples

```
  # Run a specific plan
  dominator run --plan-id abc123

  # Run in a specific directory
  dominator run --plan-id abc123 /path/to/project
```

### Options

```
  -h, --help             help for run
      --plan-id string   Plan ID to execute (required)
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
