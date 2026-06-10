---
name: nfr-judge
description: >-
  Run one NFR judgment (human-invoked /nfr-judge; 「NFR 判定して」):
  fetch thresholds, consult judgment history, run k6 through the
  external mcp-k6 server, judge pass/fail in-session, record the
  verdict, and emit feedback d-mails on fail — via the dominator MCP
  tools (ping / get_nfr / get_insights / record_result / dmail) plus
  mcp__k6__{validate_script,run_script}. One invocation = one
  judgment. All inference stays inside this interactive session
  (jun15 billing invariant; see body).
version: 0.3.2
argument-hint: "[target_id] (optional — lists configured targets if omitted)"
disable-model-invocation: true
allowed-tools:
  - Read
  - Edit
  - Write
  - Bash
  - Grep
  - Glob
  - mcp__dominator__ping
  - mcp__dominator__get_insights
  - mcp__dominator__get_nfr
  - mcp__dominator__record_result
  - mcp__dominator__dmail
  - mcp__k6__run_script
  - mcp__k6__validate_script
---

# /nfr-judge — dominator NFR judge

Human-initiated entry point. Drives the dominator MCP server's tools
plus the external mcp-k6 server's load-test tools without ever
invoking `claude -p`, so all inference happens inside this
interactive Claude Code session's subscription quota.

`disable-model-invocation: true` makes this skill human-invocable only
— the model cannot auto-trigger it. That is the mechanical form of the
jun15 invariant (LLM work fires only when a human asks).

## Execution principle: one invocation = one judgment

One `/nfr-judge` run judges **exactly one target**, then stops and
reports back to the human. Do not loop into the next target
automatically — the human re-invokes the slash command per target.
Load tests put real pressure on real systems; pacing them through a
human keeps the loop safe.

## Prerequisites

The session was launched with both MCP servers attached:

```bash
claude --mcp-config '{
  "dominator":{"command":"dominator","args":["mcp"]},
  "k6":{"command":"mcp-k6","args":["--transport","stdio"]}
}'
```

If `dominator mcp` is not on PATH, build it first:

```bash
cd path/to/dominator && go build -o ./dist/dominator ./cmd/dominator
```

`dominator mcp` must be started from the project root so it can resolve
the `.pass/` state dir (NFR config + event store). The MCP server
answers the `initialize` handshake, then exposes ping / get_nfr /
get_insights / record_result / dmail.

## Workflow

1. **Verify MCP wiring — both servers**. Call `mcp__dominator__ping`
   (must return `pong`) and confirm `mcp__k6__validate_script` /
   `mcp__k6__run_script` appear in the available tool list. If either
   server is missing, abort and ask the human to relaunch claude with
   both servers in `--mcp-config` — never substitute Bash k6.

2. **Determine the target**. The human normally supplies `target_id`.
   If they did not, `Read` `.pass/config.yaml`, list the configured
   target ids to the human, and ask them to pick one — do not pick a
   load-test target on your own.

3. **Fetch the NFR thresholds**. Call
   `mcp__dominator__get_nfr` with `{"target_id": "<id>"}`.
   The tool reads `.pass/config.yaml` and returns:

   ```json
   {
     "initialized": true,
     "target_id": "<id>",
     "nfr": {"performance": {...}, "reliability": {...}, "scalability": {...}},
     "target": {"url": "..."},
     "instruction": "Run k6 via mcp-k6, compare results against these thresholds, then call record_result with verdict='pass' or 'fail'."
   }
   ```

   If `initialized` is `false` (no `.pass/config.yaml` or load error),
   surface the `reason` and abort — the human must run `dominator mcp`
   from a project root with a valid NFR config.

4. **Consult the learning loop**. Call `mcp__dominator__get_insights`
   with no arguments. Repeated `fail` verdicts on the same target in
   `live_judgments` indicate a persistent NFR gap — apply extra
   scrutiny to that target's thresholds (from step 3) and prior
   deviations (legacy hue/coefficient ledgers) in the step 7
   comparison. A get_insights error is non-fatal: note it in the
   report and proceed with default scrutiny. Empty result = no
   history yet, proceed.

5. **Validate the k6 script**. Call `mcp__k6__validate_script` with the
   script for the target (from `.pass/k6-scripts/`; author or fix it in
   the session if missing/stale). Do not run an invalid script.

6. **Run the k6 load test**. Call `mcp__k6__run_script` with the
   validated script. Capture the metrics output.

7. **Judge pass/fail against thresholds**. Compare every captured
   metric against the `nfr` thresholds from step 3 inside this session
   — no `claude -p`. Build the comparison table for the report (one
   row per threshold):

   | NFR axis | metric | threshold | actual | verdict |
   |---|---|---|---|---|

   The overall verdict is `pass` only when every individual threshold
   passes; otherwise `fail` (list the failing rows).

8. **Record the verdict**. Call
   `mcp__dominator__record_result` with
   `{"target_id": ..., "verdict": "pass"|"fail", "summary": ...}` —
   put the failing-row digest in `summary`. The tool persists an
   `EventJudgmentRecorded` to the event store and returns
   `{"persisted": true, "recorded": true, "persistence": "event-store"}`.
   (If the MCP server was started without a writable `.pass/` dir it
   falls back to `persistence: "preview-only"` with `recorded: false` —
   report that explicitly; the judgment is then NOT durably recorded.)

9. **Emit feedback d-mails on fail**. When the verdict is `fail`,
   call `mcp__dominator__dmail` with
   `{kind: "implementation-feedback"|"design-feedback"|"report",
   name: "dom-<kind>-<target>" (stable per target/kind — a retried fail upserts instead of duplicating), description, body (the
   threshold-vs-actual table), severity}` so the implementer/designer
   receives the findings via phonewave. Re-sending the same name is an
   idempotent upsert.

10. **Report**. End with: target id, the threshold-vs-actual table, the
   recorded verdict + persistence mode, feedback d-mails emitted, and
   what the human should do next (fix and re-judge / move to the next
   target).

## Failure paths

- **mcp-k6 tools unavailable mid-run**: abort and ask the human to
  attach mcp-k6 — never substitute Bash k6.

- **MCP tool error mid-run**: report the tool name and the `reason`
  field, stop. Retry at most once for transient errors.
- **k6 validation failure**: fix the script in-session if the cause is
  obvious (syntax, missing import); otherwise report and stop. Never
  run an invalid script "to see what happens".
- **Load test aborts mid-run**: do NOT record a verdict from partial
  metrics — report the abort and stop (a judgment from incomplete
  evidence is worse than no judgment).

## Re-run idempotency

Re-invoking `/nfr-judge` after a partial run is safe: no state is
persisted until `record_result`, and recording again for the same
target appends a new judgment event (the event store is append-only;
`get_insights` aggregates the history). Avoid recording two verdicts for
the same k6 run — one run, one verdict.

## What this skill must NOT do

- Invoke `claude -p`, `claude --print`, the Anthropic Agent SDK, or
  any shell wrapper that does so (= billing boundary). The repo-wide
  semgrep gate (`.semgrep/jun15-no-headless-llm.yaml`) blocks these
  patterns in production code.
- Auto-trigger inference from a SessionStart hook or any other
  non-human-initiated path. The slash command typed by a human is
  the only valid entry to this workflow.
- Persist a result by writing to `.pass/events/` directly. The
  `record_result` MCP tool is the canonical path; it
  encapsulates the transactional event-source append (ADR 0005).
- Run k6 directly via `bash` / `Bash` — use `mcp__k6__run_script`
  exclusively so the audit trail (= OTel `messaging.*` attrs) stays
  consistent.
- Emit a feedback D-Mail by writing to `outbox/` directly with the
  Write tool (transactional outbox bypass). The `dmail` tool is the
  canonical emission path (refs issue 0031, resolved).

## Done criteria

A `/nfr-judge` run is complete when, in a real Claude Code session with
both MCP servers attached:

1. `ping` returns `pong` (handshake + tool dispatch verified).
2. `get_nfr` returns `initialized: true` with the NFR thresholds.
3. k6 validate + run succeed via mcp-k6.
4. `record_result` returns `persisted: true` / `persistence: "event-store"`,
   and the verdict shows up in `dominator status` (EventJudgmentRecorded
   read model).
5. On a `fail` verdict, the feedback d-mail is emitted via `dmail`
   (`persistence: "transactional-outbox"`).
6. The closing report (target / threshold table / verdict / d-mails /
   next step) is delivered to the human.

## Related

(Paths below live in the dominator source repo / the operator's local
refs server — not in the project this skill is materialized into.)

- Canonical pivot plan: refs archive 0027 (`http://localhost:8765/docs/archive/0027-jun15-mcp-pivot.html`, refs server)
- D-Mail emission tools: refs archive 0031 (`http://localhost:8765/docs/archive/0031-mcp-tool-surface-gaps.html`, refs server)
- Pattern reference (in the dominator repo): ADR 0003 "MCP pivot", ADR 0005 "record_result event store wiring", ADR 0007 "learning-loop read exposure"
- Self-improvement loop + semgrep gate: dominator repo docs/self-improvement-loop.md / .semgrep/jun15-no-headless-llm.yaml
- D-Mail envelope schema: see the `dmail` tool's input schema (tools/list)
