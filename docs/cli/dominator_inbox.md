## dominator inbox

Process incoming D-Mail messages

### Synopsis

Read and process D-Mail files from .pass/inbox/.

Parses YAML frontmatter from each .md file, determines a suggested action
based on the D-Mail kind, archives processed files to .pass/archive/inbox/,
and outputs a JSON array to stdout.

Exit codes:
  0 — D-Mail found and processed
  1 — inbox empty or not initialized

```
dominator inbox [path] [flags]
```

### Examples

```
  # Process inbox for current directory
  dominator inbox

  # Process inbox for a specific project
  dominator inbox /path/to/project
```

### Options

```
  -h, --help   help for inbox
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

