---
name: nfr-judge
description: >-
  Slash command for the dominator NFR judge (jun15 MCP pivot). Triggers
  when the user types "/nfr-judge", asks to "run a k6 NFR judgement via
  dominator", "check NFR thresholds with dominator", "NFR 判定して", or
  "test the dominator MCP server end-to-end". Drives the dominator MCP
  server's tools (get_nfr / record_result) plus the external mcp-k6
  server's load-test tools from inside a human-initiated Claude Code
  interactive session so inference stays on the subscription quota
  rather than the Agent SDK credit pool that gates `claude -p` from
  2026-06-15.
version: 0.3.0
argument-hint: "(none) - fetches the NFR target from dominator MCP, runs k6 via mcp-k6, records the verdict back to dominator"
disable-model-invocation: true
allowed-tools:
  - Read
  - Edit
  - Write
  - Bash
  - Grep
  - Glob
  - Agent
  - mcp__dominator__ping
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
record_result.

## Workflow

1. **Verify MCP wiring**. Call `mcp__dominator__ping`. The
   tool must return `pong`. If it errors, the dominator MCP server
   is not attached — abort and ask the human to relaunch claude
   with `--mcp-config`.

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

4. **Validate the k6 script**. Call `mcp__k6__validate_script` with the
   script for the target (from `.pass/k6-scripts/`; author or fix it in
   the session if missing/stale). Do not run an invalid script.

5. **Run the k6 load test**. Call `mcp__k6__run_script` with the
   validated script. Capture the metrics output.

6. **Judge pass/fail against thresholds**. Compare every captured
   metric against the `nfr` thresholds from step 3 inside this session
   — no `claude -p`. Build the comparison table for the report (one
   row per threshold):

   | NFR axis | metric | threshold | actual | verdict |
   |---|---|---|---|---|

   The overall verdict is `pass` only when every individual threshold
   passes; otherwise `fail` (list the failing rows).

7. **Record the verdict**. Call
   `mcp__dominator__record_result` with
   `{"target_id": ..., "verdict": "pass"|"fail", "summary": ...}` —
   put the failing-row digest in `summary`. The tool persists an
   `EventJudgmentRecorded` to the event store and returns
   `{"persisted": true, "recorded": true, "persistence": "event-store"}`.
   (If the MCP server was started without a writable `.pass/` dir it
   falls back to `persistence: "preview-only"` with `recorded: false` —
   report that explicitly; the judgment is then NOT durably recorded.)

8. **Emit feedback d-mails on fail**. When the verdict is `fail`,
   call `mcp__dominator__dmail` with
   `{kind: "implementation-feedback"|"design-feedback"|"report",
   name: "dom-<kind>-<target>-<ts>", description, body (the
   threshold-vs-actual table), severity}` so the implementer/designer
   receives the findings via phonewave. Re-sending the same name is an
   idempotent upsert.

9. **Report**. End with: target id, the threshold-vs-actual table, the
   recorded verdict + persistence mode, feedback d-mails emitted, and
   what the human should do next (fix and re-judge / move to the next
   target).

## Failure paths

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
`insights` aggregates the history). Avoid recording two verdicts for
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
  `dominator.record_result` MCP tool is the canonical path; it
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
5. The closing report (target / threshold table / verdict / next step)
   is delivered to the human.

## Related

- Canonical plan: `http://localhost:8765/docs/archive/0027-jun15-mcp-pivot.html` (refs)
- refs restructure + skill review: `http://localhost:8765/docs/issues/0030-refs-attic-restructure.html`
- D-Mail emission tool gap: `http://localhost:8765/docs/issues/0031-mcp-tool-surface-gaps.html`
- Pattern reference:
    - dominator ADR 0003 (`docs/adr/0003-mcp-pivot.md`) — MCP pivot
    - dominator ADR 0004 (`docs/adr/0004-mcp-pivot-k6-adapter-stub.md`) — K6MCPAdapter stub
    - dominator ADR 0005 (`docs/adr/0005-record-result-event-store-wiring.md`) — record_result event store wiring
- Self-improvement loop: `docs/self-improvement-loop.md`
- Mechanical gate (semgrep rules): `.semgrep/jun15-no-headless-llm.yaml`
- D-Mail 9-field schema: `internal/domain/dmail_envelope.go`
