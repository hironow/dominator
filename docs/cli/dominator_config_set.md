## dominator config set

Set a configuration value

### Synopsis

Set a single configuration key to a new value.

Supported keys:
  target.url, target.protocol, target.spec, target.docs,
  nfr.performance.p95_latency_ms, nfr.performance.error_rate_percent,
  nfr.reliability.success_rate_percent, nfr.scalability.target_rps,
  load.vus, load.duration, load.ramp_up,
  approval.required, lang, claude_cmd, model, timeout_sec

```
dominator config set <key> <value> [path] [flags]
```

### Examples

```
  # Set target URL
  dominator config set target.url https://example.com

  # Set virtual users
  dominator config set load.vus 50

  # Set config in a specific project directory
  dominator config set lang en /path/to/project
```

### Options

```
  -h, --help   help for set
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

* [dominator config](dominator_config.md)  - Manage dominator configuration
