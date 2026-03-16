## dominator config show

Show current configuration

### Synopsis

Display the current dominator configuration as YAML to stdout.

```
dominator config show [path] [flags]
```

### Examples

```
  # Show config for current directory
  dominator config show

  # Show config for a specific project
  dominator config show /path/to/project
```

### Options

```
  -h, --help   help for show
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

* [dominator config](dominator_config.md)	 - Manage dominator configuration

