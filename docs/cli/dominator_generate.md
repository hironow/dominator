## dominator generate

Generate k6 load test scripts from API spec

### Synopsis

Fetch an API specification and generate k6 load test scripts using Claude Code.

The --spec flag provides the URL of the API specification to generate tests for.
The --docs flag provides a documentation URL (required for http protocol).
The --protocol flag selects the protocol type (openapi, json-rpc, ws-json-rpc, http).
Generated scripts are written to .pass/k6-scripts/.

```
dominator generate [path] [flags]
```

### Examples

```
  # Generate from an OpenAPI spec
  dominator generate --spec https://petstore3.swagger.io/api/v3/openapi.json

  # Generate for a JSON-RPC API
  dominator generate --spec https://example.com/spec.json -p json-rpc

  # Generate from HTTP documentation (no spec)
  dominator generate --docs https://example.com/api-docs -p http

  # Overwrite existing scripts
  dominator generate --spec https://example.com/spec.json --force
```

### Options

```
      --docs string       API documentation URL (required for http protocol)
      --force             Overwrite existing scripts
  -h, --help              help for generate
  -p, --protocol string   Protocol: openapi, json-rpc, ws-json-rpc, http (default "openapi")
      --spec string       API spec URL (required for openapi/json-rpc/ws-json-rpc)
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
