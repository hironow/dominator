## dominator init

Initialize .pass directory

### Synopsis

Initialize the .pass/ state directory for dominator NFR judgment.

Creates the directory structure required by dominator: config.yaml,
events/, .run/, archive/, k6-scripts/, insights/, skills/. If [path] is
omitted, the current working directory is used. Use --force to reinitialize
an existing .pass/ directory.

Optionally configure an OpenTelemetry backend (Jaeger or Weave) with
the --otel-backend flag. The generated .otel.env file is written into
.pass/ and loaded automatically on subsequent runs.

```
dominator init [path] [flags]
```

### Examples

```
  # Initialize in current directory
  dominator init

  # Initialize a specific project directory
  dominator init /path/to/project

  # Reinitialize (overwrite existing .pass/)
  dominator init --force

  # Initialize with Jaeger OTel backend
  dominator init --otel-backend jaeger
```

### Options

```
      --force                 Overwrite existing state directory (re-initialize)
  -h, --help                  help for init
      --otel-backend string   OTel backend: jaeger, weave
      --otel-entity string    Weave entity/team (required for weave)
      --otel-project string   Weave project (required for weave)
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
