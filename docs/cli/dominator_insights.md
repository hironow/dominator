## dominator insights

Display judgment insights (hue and coefficient)

### Synopsis

Read and display structured insight data from .pass/insights/.

Parses hue.md (judgment history) and coefficient.md (violation details)
into structured JSON output on stdout.

```
dominator insights [path] [flags]
```

### Examples

```
  # Show insights for current directory
  dominator insights

  # Show insights for a specific project
  dominator insights /path/to/project
```

### Options

```
  -h, --help   help for insights
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

