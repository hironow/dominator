# What / Why / How Conformance

This is the single source of truth for dominator's purpose, design rationale, and implementation approach.
Referenced from [README.md](../README.md) and [docs/README.md](README.md).

| Aspect | Description |
|--------|-------------|
| **What** | MCP server + data plane for NFR judgment: serves configured NFR thresholds and records externally judged k6 results |
| **Why** | Keep NFR evidence in the shared event store while ensuring k6 execution and any LLM reasoning happen inside a human-initiated claude-code session |
| **How** | `dominator mcp` serves MCP tools (`get_nfr`, `record_result`); the `/nfr-judge` skill drives mcp-k6 from the claude-code session and records the verdict through `record_result` |
| **Input** | `.pass/config.yaml`, event store, MCP tool arguments, externally judged k6 result summaries |
| **Output** | MCP tool responses and `EventJudgmentRecorded` entries in the event store |
| **Telemetry** | OTel spans on command roots and MCP tool handlers; MCP invocation metrics support post-jun15 cost verification |
| **External Systems** | Local filesystem, external mcp-k6 server attached to the claude-code session, OTel exporter (Jaeger/Weave), claude-code session as MCP client |

## Layer Architecture

```
cmd              --> usecase, session, usecase/port, platform, domain  (composition root)
usecase          --> usecase/port, domain                              (output port only)
usecase/port     --> domain (+ stdlib)                                 (interface contracts)
session          --> eventsource, usecase/port, platform, domain       (adapter impl)
eventsource      --> domain                                            (event persistence adapter)
platform         --> domain (+ stdlib)                                 (cross-cutting infra)
domain           --> (nothing internal, stdlib only)                   (pure types/logic)
```

`eventsource` is the event persistence adapter based on the [AWS Event Sourcing pattern](https://docs.aws.amazon.com/prescriptive-guidance/latest/cloud-design-patterns/event-sourcing.html).
Its responsibility is limited to append, load, and replay of domain events.
Event store implementation MUST NOT exist outside `internal/eventsource`.
`session` uses `eventsource` as a client but does not implement event persistence itself.

Key constraints enforced by semgrep (ERROR severity):

- `usecase --> session` PROHIBITED (must use output port interfaces)
- `cmd --> eventsource` PROHIBITED (ADR S0008)
- `domain` has no I/O, no `context.Context`

Ref: `.semgrep/layers.yaml`, ADR S0007

## Domain Primitives & Parse-Don't-Validate

Domain command types use the Parse-Don't-Validate pattern:

- Domain primitives (`RepoPath`, `SpecURL`, `Protocol`, `Days`) validate in `New*()` constructors — invalid values are rejected at parse time
- Command types use unexported fields with `New*Command()` constructors that accept only pre-validated primitives
- Commands are always-valid by construction — no `Validate() []error` methods exist
- Usecase layer receives always-valid commands with no validation boilerplate

Ref: `.semgrep/layers.yaml`, ADR S0029

## Cross-Tool Conformance

All tools (phonewave, sightjack, paintress, amadeus, dominator) maintain a What/Why/How conformance table in `docs/conformance.md` with the same structure. This prevents expression drift across README files.

## MCP Pivot Boundary

Dominator no longer starts Claude or k6 from the Go CLI. The old `generate`, `run`, and `validate` commands are deprecation stubs; `check` and `approve` remain local data-plane helpers for existing k6 scripts and plan metadata.

- `dominator mcp` implements the MCP lifecycle (`initialize`, `notifications/initialized`, `tools/list`, `tools/call`) over stdio.
- `get_nfr` reads the NFR thresholds from `.pass/config.yaml`.
- `record_result` records a pass/fail verdict from the claude-code session as `EventJudgmentRecorded`.
- The `/nfr-judge` skill drives mcp-k6 and performs result comparison from the claude-code session.

Ref: ADR 0003, ADR 0004, ADR 0005, `internal/session/mcp_server.go`, `plugins/dominator/skills/nfr-judge/SKILL.md`
