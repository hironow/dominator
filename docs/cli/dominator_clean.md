## dominator clean

Remove state directory (.pass/)

### Synopsis

Delete the .pass/ directory to reset to a clean state. Use 'dominator init' to reinitialize.

```
dominator clean [path] [flags]
```

### Examples

```
  # Clean current directory (interactive confirmation)
  dominator clean

  # Clean a specific project directory
  dominator clean /path/to/project

  # Skip confirmation prompt
  dominator clean --yes
```

### Options

```
  -h, --help   help for clean
      --yes    Skip confirmation prompt
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

