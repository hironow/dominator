# Dominator

**An MCP server + data plane for NFR judgment: it serves configured NFR thresholds and records externally judged k6 results.**

Following the jun15 MCP pivot, LLM ownership and k6 execution moved to a human-initiated [Claude Code](https://docs.anthropic.com/en/docs/claude-code) session. Dominator the Go CLI is now a data plane: it serves NFR thresholds over MCP, records pass/fail verdicts in the event store, and provides supporting local state commands. The old headless generation / validation / run loop has been retired.

```bash
dominator mcp
```

`dominator mcp` starts the MCP server. Its tools expose:

1. `dominator.ping` — health check
2. `dominator.get_nfr` — read NFR thresholds from `.pass/config.yaml`
3. `dominator.record_result` — persist a pass/fail verdict as `EventJudgmentRecorded`

The claude-code session runs the `/nfr-judge` skill with both Dominator and mcp-k6 attached. That session validates/runs k6 through mcp-k6, compares the result against the thresholds returned by `dominator.get_nfr`, and records the final verdict through `dominator.record_result`.

## Why "Dominator"?

The system design is inspired by [PSYCHO-PASS](https://en.wikipedia.org/wiki/Psycho-Pass), a cyberpunk anime by Production I.G (2012).

In the story, the Sibyl System continuously monitors citizens' mental states to produce a Crime Coefficient — a numerical measure of criminal intent. Inspectors carry the Dominator, a weapon that reads these coefficients in real time and adjusts its response (from non-lethal paralyzer to lethal eliminator) based on the target's threat level. The system's judgment is absolute: if your Psycho-Pass hue is clouded, action is taken.

This structure maps directly to NFR validation:

| PSYCHO-PASS Concept | Dominator | Design Meaning |
|---|---|---|
| **Dominator** | This binary | The instrument that measures and judges system health |
| **Sibyl System** | k6 + Claude Code | External systems that produce raw measurements |
| **Crime Coefficient** | NFR deviation score | Numerical measure of how far metrics deviate from thresholds |
| **Psycho-Pass Hue** | Hue insight (`hue.md`) | Overall judgment history — clear or clouded |
| **Inspector** | Human operator | Reviews and approves execution plans before judgment |
| **Paralyzer Mode** | Low severity D-Mail | Minor deviation — informational notification |
| **Eliminator Mode** | High severity D-Mail | Critical violation — urgent corrective action required |
| **Area Stress** | Load test parameters (VUs, duration) | Environmental pressure applied to the system under test |
| **.pass/** | Public Safety Bureau | Persistent state directory that tracks all judgments |

### Three Design Principles

1. **Measure, don't assume** — Like the Sibyl System, quantify NFR compliance with objective load testing rather than subjective assessment.
2. **Graduated response** — Severity scales with deviation: low (<10%), medium (10-50%), high (>50%). Response intensity matches the threat level.
3. **Judgment requires approval** — The Dominator requires the Sibyl System's authorization. Load tests require explicit human approval before execution.

---

## CLI Flow

The current workflow follows the MCP pivot boundary:

```
claude-code session -> mcp-k6 -> dominator.record_result
```

1. **Start MCP** — `dominator mcp` serves NFR thresholds and result recording
2. **Run Skill** — Claude Code invokes `/nfr-judge`
3. **Execute k6** — The session calls mcp-k6, not the Go CLI
4. **Record Result** — The session calls `dominator.record_result`
5. **Inspect State** — `dominator status`, `dominator insights`, and event replay commands read the local state

## Quick Start

```bash
# Build from source
just install

# Initialize .pass/ with default config
dominator init

# Start the data-plane server for a claude-code session
dominator mcp

# In the claude-code session, attach dominator + mcp-k6 and run /nfr-judge

# View judgment insights
dominator insights
```

## Subcommands

Running `dominator` without a subcommand shows usage help.

| Command | Description |
|---------|-------------|
| `init` | Initialize `.pass/` directory |
| `mcp` | Start the MCP server (data plane: ping / get_nfr / record_result) |
| `generate` | Retired redirect; generate scripts from the Claude Code session |
| `check` | Local helper: create a plan from existing k6 scripts |
| `approve` | Local helper: approve an existing plan |
| `run` | Retired redirect; run k6 via Claude Code + mcp-k6 + `/nfr-judge` |
| `validate` | Retired redirect; validate k6 scripts via Claude Code + mcp-k6 |
| `insights` | Display judgment insights (hue and coefficient) |
| `inbox` | Process incoming D-Mail messages |
| `config show` / `config set` | View or update configuration |
| `doctor` | Run health checks |
| `status` | Show operational status |
| `archive-prune` | Prune old archived files |
| `clean` | Remove state directory (`.pass/`) |
| `update` | Self-update to the latest release |
| `version` | Print version, commit, and build information |

All commands accept an optional `[path]` argument (defaults to cwd). For flags, examples, and full reference per subcommand, see [docs/cli/](docs/cli/).

## Configuration

```yaml
# .pass/config.yaml
lang: ja
claude_cmd: claude
model: opus
timeout_sec: 1980

target:
  url: https://api.example.com
  protocol: openapi
  spec: https://api.example.com/openapi.json

nfr:
  performance:
    p95_latency_ms: 500
    error_rate_percent: 1.0
  reliability:
    success_rate_percent: 99.0
  scalability:
    target_rps: 100

load:
  vus: 10
  duration: "30s"
  ramp_up: "5s"

approval:
  required: true
```

## NFR Judgment

Dominator evaluates four metrics against configurable thresholds:

| Metric | Direction | Violation Condition |
|--------|-----------|-------------------|
| `p95_latency_ms` | Lower is better | Actual exceeds threshold |
| `error_rate_percent` | Lower is better | Actual exceeds threshold |
| `success_rate_percent` | Higher is better | Actual falls below threshold |
| `target_rps` | Higher is better | Actual falls below threshold |

### Severity Levels

Deviation percentage determines severity:

```
< 10%  deviation -> low    (informational)
10-50% deviation -> medium (elevated priority)
> 50%  deviation -> high   (critical — urgent action required)
```

### Verdict

- **pass** — All metrics within thresholds (no deviations)
- **violation** — One or more metrics exceed thresholds

## D-Mail Protocol

Dominator is the NFR judge in the D-Mail protocol ecosystem:

| Tool | Role | Endpoint |
|------|------|----------|
| **sightjack** | Designer / Protocol spec owner | `.siren/` |
| **paintress** | Implementer | `.expedition/` |
| **amadeus** | Verifier | `.gate/` |
| **dominator** | NFR Judge | `.pass/` |
| **phonewave** | Courier / Coordinator | (no endpoint — routes between others) |

### Produces

| Kind | Prefix | Trigger | Target |
|------|--------|---------|--------|
| `design-feedback` | `design-feedback-` | NFR violation | sightjack (design review) |
| `implementation-feedback` | `implementation-feedback-` | NFR violation | paintress (implementation review) |
| `verification-feedback` | `verification-feedback-` | NFR violation | amadeus (verification review) |
| `report` (nfr-pass) | `nfr-pass-` | All NFRs pass | Informational |

### Consumes

| Kind | Source | Action |
|------|--------|--------|
| `implementation-feedback` | amadeus / other | Informs next judgment context |
| `convergence` | amadeus | Convergence state update |

## Hue / Coefficient Insights

Dominator records judgment history in `.pass/insights/`:

- **hue.md** — Chronological judgment results (pass/violation) with details. Analogous to a citizen's Psycho-Pass hue: consistently clear results indicate a healthy system.
- **coefficient.md** — Detailed deviation tables per judgment. Analogous to the Crime Coefficient: specific metrics that caused the violation.

Read insights programmatically:

```bash
dominator insights | jq '.hue'
dominator insights | jq '.coefficient'
```

## Architecture

```
claude-code session
    |
    |  /nfr-judge skill
    |  +-- Call dominator.get_nfr
    |  +-- Call mcp-k6 validate_script / run_script
    |  +-- Compare metrics against thresholds
    |  +-- Call dominator.record_result
    |
    v
dominator mcp  (MCP server / data plane)
    |
    |  dominator.get_nfr        -> read .pass/config.yaml
    |  dominator.record_result  -> append EventJudgmentRecorded
    |
.pass/                   <- Persistent state
    +-- config.yaml           <- Target, NFR thresholds, load config
    +-- k6-scripts/           <- k6 scripts authored outside the Go CLI
    +-- insights/             <- Hue + coefficient insight ledger
    |   +-- hue.md            <- Judgment history
    |   +-- coefficient.md    <- Deviation details
    +-- .run/                 <- Ephemeral state (gitignored)
    |   +-- plans/            <- Execution plans (JSON)
    |   +-- latest.json       <- Latest judgment state
    +-- events/               <- Append-only event log (JSONL, daily rotation)
    +-- outbox/               <- Outgoing D-Mails (picked up by phonewave)
    +-- inbox/                <- Incoming D-Mails
    +-- archive/              <- Permanent D-Mail audit trail
    +-- skills/               <- Agent skill manifests (phonewave discovery)
```

## Protocol Support

| Protocol | Spec Source | k6 Module | Notes |
|----------|-----------|-----------|-------|
| `openapi` | OpenAPI JSON/YAML | `k6/http` | REST APIs |
| `json-rpc` | JSON-RPC spec | `k6/http` | HTTP POST with JSON-RPC 2.0 |
| `ws-json-rpc` | WebSocket JSON-RPC spec | `k6/ws` | WebSocket with JSON-RPC 2.0 |
| `http` | HTTP documentation | `k6/http` | Generic HTTP endpoints |

## Event Types

| Event Type | Trigger | Description |
|-----------|---------|-------------|
| `script.generated` | Legacy / replayed history | k6 script created from API spec before the MCP pivot |
| `generation.failed` | Legacy / replayed history | Script generation error before the MCP pivot |
| `plan.created` | `check` completes | Execution plan created |
| `plan.approved` | `approve` completes | Plan approved for execution |
| `judgment.recorded` | `dominator.record_result` MCP tool | Session-judged pass/fail result recorded |
| `judged` | Legacy / replayed history | CLI-driven judgment result before the MCP pivot |
| `violation.detected` | Legacy / replayed history | Violation with deviations before the MCP pivot |
| `pass.confirmed` | Legacy / replayed history | Clean judgment before the MCP pivot |

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | Runtime error |

## Tracing (OpenTelemetry)

Dominator instruments key operations with OpenTelemetry spans. Tracing is off by default (noop tracer) and activates when `OTEL_EXPORTER_OTLP_ENDPOINT` is set.

```bash
# Start Jaeger v2 (trace viewer)
just jaeger

# Run the MCP data plane with tracing enabled
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318 dominator mcp

# View traces at http://localhost:16686

# Stop Jaeger
just jaeger-down
```

## Development

All code lives in `internal/` (Go convention). See [docs/conformance.md](docs/conformance.md) for layer architecture and directory responsibilities. Run `just --list` for available tasks.

## The Ecosystem

Dominator is the NFR judge in a multi-tool AI development ecosystem:

```
Sightjack (design)        Paintress (implement)      Amadeus (verify)        Dominator (NFR judge)
    |                          |                          |                       |
    |  Issue architecture      |  Autonomous impl         |  Integrity verify     |  Load test + judge
    |  DoD, dependencies       |  Code, tests, PRs        |  Divergence scoring   |  NFR compliance
    |  Wave-by-wave approval   |  Expedition loop         |  D-Mail routing       |  Hue / coefficient
    |                          |                          |                       |
    v                          v                          v                       v
Linear Issues ---------> Git Repository -----------> .gate/              -> .pass/
                               |                         |                       |
                  D-Mail       |        D-Mail           |        D-Mail         |
                 (report) -----+----> inbox/        outbox/ ----> inbox/    outbox/ ---->
                                                                                   design-feedback
                                                                                   impl-feedback
                                                                                   verification-feedback
```

## Documentation

- [docs/](docs/README.md) — Full documentation index
- [docs/conformance.md](docs/conformance.md) — What/Why/How conformance table
- [docs/pass-directory.md](docs/pass-directory.md) — `.pass/` directory structure
- [docs/policies.md](docs/policies.md) — Event -> Policy mapping
- [docs/otel-backends.md](docs/otel-backends.md) — OTel backend configuration
- [docs/testing.md](docs/testing.md) — Test strategy and conventions
- [docs/adr/](docs/adr/README.md) — Architecture Decision Records
- [docs/shared-adr/](docs/shared-adr/README.md) — Cross-tool shared ADRs

## Prerequisites

- Go 1.26+
- [Claude Code CLI](https://docs.anthropic.com/en/docs/claude-code)
- mcp-k6 or an equivalent k6 MCP server attached to the claude-code session
- [Docker](https://www.docker.com/) (optional, for Jaeger tracing)

Run `dominator doctor` to verify all prerequisites.

## License

Apache License 2.0
See [LICENSE](./LICENSE) for details.
