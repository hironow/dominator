# What / Why / How Conformance

This is the single source of truth for dominator's purpose, design rationale, and implementation approach.
Referenced from [README.md](../README.md) and [docs/README.md](README.md).

| Aspect | Description |
|--------|-------------|
| **What** | NFR validation system that judges system quality via k6 load testing and routes corrective D-Mails |
| **Why** | Detect non-functional requirement violations early with objective, automated load testing rather than subjective assessment |
| **How** | Generate k6 scripts from API specs via Claude Code -> create execution plans -> human approval -> execute k6 -> evaluate NFR thresholds -> record hue/coefficient insights -> emit D-Mails by severity |
| **Input** | API specifications (OpenAPI, JSON-RPC, WebSocket, HTTP docs), NFR thresholds (config.yaml), inbox D-Mails |
| **Output** | Judgment results (pass/violation), D-Mails (design-feedback / implementation-feedback / verification-feedback / nfr-pass), insight ledger (hue.md, coefficient.md) |
| **Telemetry** | OTel spans: `dominator.run`, `dominator.generate`, `claude.invoke` (with `claude.model`, `claude.timeout_sec`, `gen_ai.*`) |
| **External Systems** | Claude Code subprocess, k6 load testing engine, OTel exporter (Jaeger/Weave) |

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
