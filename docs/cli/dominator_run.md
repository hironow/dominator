## dominator run

Deprecated (jun15 MCP pivot): use claude code + mcp-k6 + /nfr-judge

### Synopsis

Deprecated by the jun15 MCP pivot (2026-06-15 credit-pool split).

k6 load-test execution + NFR judging are no longer driven by a headless
'claude -p' loop. From a claude code session with the mcp-k6 server
attached, use the /nfr-judge skill: it runs k6 via mcp-k6 and records the
verdict through dominator's MCP tools (get_nfr, record_result). Start the
data-plane server with 'dominator mcp'.

```
dominator run [path] [flags]
```

### Options

```
  -h, --help   help for run
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

