## dominator

NFR Judge for your system

```
dominator [flags]
```

### Options

```
  -c, --config string   config file path
  -h, --help            help for dominator
  -l, --lang string     output language (ja, en)
      --no-color        Disable colored output (respects NO_COLOR env)
  -o, --output string   Output format: text, json (default "text")
  -q, --quiet           Suppress all stderr output
  -v, --verbose         verbose output
```

### SEE ALSO

* [dominator approve](dominator_approve.md)	 - Approve an execution plan
* [dominator archive-prune](dominator_archive-prune.md)	 - Prune old archived files
* [dominator check](dominator_check.md)	 - Create an execution plan from k6 scripts
* [dominator clean](dominator_clean.md)	 - Remove state directory (.pass/)
* [dominator config](dominator_config.md)	 - Manage dominator configuration
* [dominator doctor](dominator_doctor.md)	 - Run health checks
* [dominator generate](dominator_generate.md)	 - Generate k6 load test scripts from API spec
* [dominator inbox](dominator_inbox.md)	 - Process incoming D-Mail messages
* [dominator init](dominator_init.md)	 - Initialize .pass directory
* [dominator insights](dominator_insights.md)	 - Display judgment insights (hue and coefficient)
* [dominator mcp](dominator_mcp.md)	 - Run dominator as an MCP server over stdio (NFR data plane: get_nfr + record_result)
* [dominator run](dominator_run.md)	 - Execute k6 load test and judge NFR compliance
* [dominator status](dominator_status.md)	 - Show dominator operational status
* [dominator update](dominator_update.md)	 - Self-update dominator to the latest release
* [dominator validate](dominator_validate.md)	 - Validate k6 scripts via mcp-k6
* [dominator version](dominator_version.md)	 - Print version, commit, and build information

