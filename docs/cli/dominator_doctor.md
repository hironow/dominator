## dominator doctor

Run health checks

### Synopsis

Run health checks on the dominator environment.

Each check reports one of four statuses: OK (passed), FAIL (exit 1),
SKIP (dependency missing), WARN (advisory, exit 0).

```
dominator doctor [path] [flags]
```

### Examples

```
  # Run environment check in current directory
  dominator doctor

  # Check a specific project directory
  dominator doctor /path/to/project

  # JSON output for scripting
  dominator doctor --json

  # Auto-fix repairable issues
  dominator doctor --repair
```

### Options

```
  -h, --help     help for doctor
  -j, --json     output as JSON
      --repair   Auto-fix repairable issues
```

### Options inherited from parent commands

```
  -c, --config string   config file path
  -l, --lang string     output language (ja, en)
      --no-color        Disable colored output (respects NO_COLOR env)
  -o, --output string   Output format: text, json (default "text")
  -v, --verbose         verbose output
```

### SEE ALSO

* [dominator](dominator.md)	 - NFR Judge for your system

