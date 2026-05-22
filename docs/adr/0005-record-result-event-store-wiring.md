# 0005. record_result event store wiring (Phase 4 follow-up)

**Date:** 2026-05-22
**Status:** Accepted

## Context

The `dominator.record_result` MCP tool shipped in refs/issues/0027
Phase 2c (ADR 0003) as a preview-only stub (`persistence='preview-only'`,
`persisted: false`). It validated the input (`target_id`, `verdict`,
`summary`) but did not append a domain event. The Phase 3 note left an
explicit marker: "Phase 4 follow-up wires the aggregate + emitter chain
for EventJudgmentRecorded."

paintress Phase 4 #4 (PR #218) established the emitter-injection pattern:
the cmd composition root constructs the aggregate + event store + emitter
and injects a narrow `port.*EventEmitter` into the session MCP server,
keeping the `session-no-direct-new-aggregate` /
`session-no-direct-new-event` semgrep rules satisfied. This ADR applies
that pattern to dominator's record_result.

After the jun15 MCP pivot (ADR 0003) and the K6MCPAdapter stub-out
(ADR 0004), `dominator run` / `dominator validate` are deprecated, so
`record_result` is the only path that records a judgment into the event
store. A human-initiated claude code session runs the load test via
mcp-k6, compares against the NFR thresholds (from `dominator.get_nfr`),
and calls `record_result` with the verdict.

## Decision

`record_result` persists an `EventJudgmentRecorded` to the event store
when an emitter is wired:

1. **New event type `EventJudgmentRecorded`** (`"judgment.recorded"`) with
   a lightweight payload `JudgmentRecordedData{ TargetID, Verdict, Summary }`.
   This is distinct from `EventJudged` (which carries full `JudgedData`
   k6 metrics from a CLI-driven run). The session judged externally via
   mcp-k6 and reports only the verdict + summary — no metrics. Reusing
   `EventJudged` with empty metrics would mislead the `status.go` read
   model (= a judgment with zero metrics), so a separate event type is
   the correct event-sourcing semantics.

2. **`JudgeAggregate.RecordJudgment(targetID, verdict, summary, now)`**
   produces the event, validating `verdict ∈ {"pass","fail"}` and
   normalising it to `domain.Verdict` (pass→VerdictPass,
   fail→VerdictViolation). It does not consult the NFR config.

3. **`port.JudgmentEventEmitter` + `usecase.NewJudgmentEventEmitter`**:
   the emitter wraps `agg.RecordJudgment() → store.Append()`. The cmd
   composition root (`internal/cmd/mcp.go`) builds it and injects it via
   `MCPServer.WithEmitter`. A `NopJudgmentEventEmitter` null object
   covers the preview-only path.

4. **`status.go` read model** gains an `EventJudgmentRecorded` case that
   counts the judgment and reflects its verdict, mirroring the existing
   `EventPassConfirmed` / `EventViolationDetected` handling.

5. **Config-absent is acceptable**: `session.LoadConfig` returns
   `DefaultConfig()` when `config.yaml` is absent (no error), so the
   emitter is wired even without an NFR config. This is correct —
   `RecordJudgment` does not read the config. Only a *malformed*
   `config.yaml` (a real load error) disables persistence, falling back
   to preview-only.

## LLM firing boundary

`record_result` only persists an event; it never invokes the model. The
MCP tool call is driven by a human-initiated claude code session. The
jun15 invariant ("LLM firing is always human-initiated") is preserved:
the emitter fires only on a session-driven tool call, never from a
daemon or hook.

## Consequences

### Positive

- `record_result` is no longer preview-only; session-initiated
  judgments are durably recorded as `EventJudgmentRecorded` and surface
  in `dominator status`.
- The lightweight event type keeps the `JudgedData` metrics contract
  intact for any future CLI-driven judgment path.
- The emitter-injection pattern matches paintress Phase 4 #4, so the
  cross-tool MCP server architecture stays symmetric.

### Negative

- A second judgment event type (`EventJudged` vs `EventJudgmentRecorded`)
  means read-model consumers must handle both. `status.go` does; any
  future projection must too.

### Neutral

- The `JudgeAggregate` config field is unused by `RecordJudgment`. It is
  retained because the aggregate's other (CLI) responsibilities consult
  it, and the cmd composition root passes the loaded config regardless.

## References

- refs/issues/0027 — jun15 MCP pivot canonical plan (Phase 4 follow-up).
- [ADR 0003](0003-mcp-pivot.md) — the MCP pivot that introduced
  record_result as a preview-only stub.
- [ADR 0004](0004-mcp-pivot-k6-adapter-stub.md) — K6MCPAdapter stub-out
  that made record_result the sole judgment-recording path.
- paintress ADR 0016 / PR #218 — the emitter-injection pattern this ADR
  mirrors.
