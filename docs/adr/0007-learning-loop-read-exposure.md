# 0007. Learning-loop read exposure (get_insights)

**Date:** 2026-06-10
**Status:** Accepted

## Context

The hue / coefficient insight ledgers were written by
`InsightWriter.RecordHue` / `RecordCoefficient` — and verification on
2026-06-10 (refs issue 0034) showed those writers have had **zero
production callers** since the jun15 MCP pivot retired the Go-owned
judging loop. Judgment history, however, accumulates live in the event
store as `EventJudgmentRecorded` (the record_result MCP tool, ADR
0005).

## Decision

1. **`get_insights` read-only MCP tool**: the live judgment summary
   derived from `EventJudgmentRecorded` (count / fail count / last
   target / last verdict) is the **source of truth**; `hue.md` /
   `coefficient.md` are surfaced via the existing `InsightReader` as
   the **legacy ledger** when present.
2. **The dormant writers are not revived**: the event store is the
   durable learning input; no write path is introduced.
3. **Empty state is not an error.**
4. **Skill v0.3.1**: `/nfr-judge` consults `get_insights` before
   judging (repeated fail verdicts on a target = persistent NFR gap).

## Consequences

### Positive

- The judge's loop closes: past verdicts shape the next judgment,
  with zero new persistence and no port wiring.

### Negative

- hue.md/coefficient.md stay frozen at their pre-pivot content until a
  future wave retires or revives the writers.

### Neutral

- `dominator insights` (CLI) continues to read the same legacy files.
