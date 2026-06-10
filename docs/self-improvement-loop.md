# dominator self-improvement loop

## Purpose

`dominator` is the NFR judgment side of the 5-tool loop.

It sits on the path:

`specification -> implementation -> verification -> NFR judgment -> correction`

and is responsible for turning load-test evidence into durable pass/fail
verdicts without owning the load-test execution or the inference itself.

## What this tool now does

Post jun15 MCP pivot, `dominator` participates in the observable
self-improvement loop as a data plane:

1. It serves the configured NFR thresholds (`.pass/config.yaml`) over MCP
   (`get_nfr`) to a human-initiated Claude Code session.
2. The session runs k6 through mcp-k6, compares metrics against the
   thresholds, and records the verdict via `record_result`,
   which persists an `EventJudgmentRecorded` to the event store.
3. It accumulates judgment insights in `.pass/insights/` — `hue.md`
   (recent judgment tendency) and `coefficient.md` (threat-level history)
   — so repeated NFR failures on the same target become visible across
   sessions.

## Loop participation

| Loop stage | dominator's contribution |
|---|---|
| observe | `dominator status` / `insights` read models over the event store |
| judge | thresholds via `get_nfr`; verdict persisted via `record_result` |
| learn | hue / coefficient insight ledger (git-tracked) |
| correct | feedback D-Mail to designer/implementer — **currently blocked**: no emission MCP tool exists (tracked in refs issue 0031); `inbox` consumes incoming D-Mails |

## Negative feedback discipline

Every judgment requires a human-initiated session (`/nfr-judge`); dominator
never fires inference or load on its own. This keeps the loop's gain
negative: repeated failures raise the coefficient and surface in insights,
but escalation is always a human decision.

## Boundaries

- k6 execution is owned by the session (mcp-k6), not by dominator.
- LLM inference is owned by the session, never by the Go CLI
  (semgrep gate `.semgrep/jun15-no-headless-llm.yaml`).
- Verdict persistence is owned by dominator (transactional event append,
  ADR 0005) — the session must not write `.pass/events/` directly.
