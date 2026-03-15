# dominator

NFR (Non-Functional Requirements) Judge — validates system quality attributes via k6 load testing.

Inspired by PSYCHO-PASS Dominator.

## Identity

- **Module**: `github.com/hironow/dominator`
- **State dir**: `.pass/`
- **License**: Apache 2.0
- **Go**: 1.26

## Architecture

### Patterns

- **Hexagonal (Port-Adapter)**: `cmd -> usecase -> usecase/port <- session, eventsource, platform, domain`
- **Event Sourcing**: append-only JSONL in `.pass/events/YYYY-MM-DD.jsonl`
- **Transactional Outbox**: SQLite WAL in `.pass/.run/`, stage -> flush -> archive
- **D-Mail Protocol**: YAML frontmatter + Markdown, kinds: specification, report, feedback, convergence
- **Parse-Don't-Validate**: domain primitives validate in constructors, commands always-valid

### Layer Architecture

```
cmd/                  Entry point (composition root)
internal/cmd/         CLI command definitions (cobra)
internal/usecase/     Orchestration (COMMAND, Aggregate, POLICY dispatch)
internal/usecase/port/ Output port interfaces (context-aware contracts)
internal/session/     I/O adapters (filesystem, network, subprocess)
internal/eventsource/ Event store (JSONL persistence)
internal/platform/    Cross-cutting (logger, telemetry, metrics)
internal/domain/      Pure types/logic (no I/O, no context.Context)
```

### Layer Rules (enforced by semgrep)

- `cmd` -> usecase, session, usecase/port, platform, domain (composition root)
- `usecase` -> usecase/port, domain (output port interface only)
- `session` -> eventsource, usecase/port, platform, domain (adapter implementation)
- `eventsource` -> domain
- `usecase/port` -> domain (+ stdlib)
- `platform` -> domain (+ stdlib)
- `domain` -> (nothing internal, stdlib only)

**PROHIBITED**:
- `cmd` -> eventsource (ADR S0008)
- `usecase` -> session, cmd, eventsource
- `session` -> cmd, usecase (except usecase/port)
- `domain` -> any upper layer, stdlib I/O packages

## Conventions

### UNIX stdio

- **stdout**: machine-readable data only (JSON). No human messages.
- **stderr**: logs, progress, diagnostics. All human-readable output.
- No log files on local disk. Logs stream to stderr.

### OTel

- Noop by default. Opt-in via `OTEL_EXPORTER_OTLP_ENDPOINT`.
- TracerProvider/MeterProvider set only in `InitTracer()`.
- All spans must have `defer span.End()`.

### CLI (cobra)

- All commands use `RunE` (not `Run`).
- Use `cmd.OutOrStdout()`, `cmd.ErrOrStderr()`, `cmd.InOrStdin()` — never raw `os.Stdout/Stderr/Stdin`.
- Use `cmd.Context()` — never `context.Background()` in command handlers.
- No `os.Exit()` in command handlers.

### Testing

- External test packages (`_test` suffix) preferred over same-package tests.
- Same-package tests require `white-box-reason:` comment.
- E2E: strictly no mocks, all real dependencies.
- `nosemgrep:` annotations require `[permanent]` or `[expires: YYYY-MM-DD]` tag.

### SQLite

- WAL mode for concurrent access.
- Connections in `.pass/.run/`.
- All `sql.Open()` must have `defer db.Close()`.

### Idempotency

- All operations must be idempotent.
- Multiple concurrent CLI invocations must not interfere with each other.
- No directory-level blocking (directory A must not block directory B).

## Domain Concepts

- **Judge**: NFR validation aggregate (load test execution + result evaluation)
- **NFR**: non-functional requirement (latency, throughput, error rate, etc.)
- **k6**: load testing tool used as external system
- **Pass/Fail**: judgment result based on NFR thresholds

## Task Runner

```bash
just build     # Build binary
just test      # Run all tests
just semgrep   # Run semgrep layer rules
just lint      # Full lint (vet + semgrep + root-guard + markdown)
just check     # Full pre-commit check
```
