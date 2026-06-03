# dominator claude code plugin (jun15 MCP pivot)

**Status:** Phase 2c in progress (= MCP server stub + skill skeleton).
Production target for the post-2026-06-15 architecture where claude
code interactive sessions own LLM inference and the dominator Go CLI
exposes its NFR config + k6 run history as an MCP server. Pattern
referenced from paintress Phase 1 (ADR 0017) + sightjack Phase 2a
(ADR 0018) + amadeus Phase 2b (ADR 0026).

## Layout

```
plugins/dominator/
├── README.md                    # this file
└── skills/
    └── nfr-judge/SKILL.md       # /nfr-judge slash command
```

## Loading the plugin

```bash
claude \
  --plugin-dir ./plugins/dominator \
  --mcp-config '{
    "dominator":{"command":"dominator","args":["mcp"]},
    "k6":{"command":"mcp-k6","args":["--transport","stdio"]}
  }'
```

The `--plugin-dir` flag registers the skills; the `--mcp-config`
flag attaches both the dominator MCP server (`dominator mcp`
subcommand, for NFR config + run history) **and** the external
mcp-k6 server (for k6 load test execution). The skill calls
`mcp__dominator__*` for its own data plane and `mcp__k6__*` for
load testing.

## Phase 2c MVP scope

Only `/nfr-judge` is wired. The slash command calls the dominator
MCP server's stub tools (dominator.ping, dominator.get_nfr,
dominator.record_result) and surfaces the stub contract to the
human. Real domain wiring lands in subsequent commits on
`feat/jun15-mcp-pivot`.

## Why two MCP servers

- **dominator** (this plugin) — dominator's own data plane:
  NFR specifications, run history, verdict persistence
- **mcp-k6** (external, <https://github.com/grafana/mcp-k6>) —
  k6 script validation + execution

The session attaches both, the skill orchestrates them. Separating
the data plane from the execution plane keeps each MCP server
narrowly scoped and independently versioned.
