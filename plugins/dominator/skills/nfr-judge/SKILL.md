---
name: nfr-judge
description: >-
  Slash command for the dominator NFR judge (refs/issues/0027 jun15
  MCP pivot). Triggers when the user types "/nfr-judge", asks to
  "run a k6 NFR judgement via dominator", "check NFR thresholds with
  dominator", or "test the dominator MCP server end-to-end". Drives the
  dominator MCP server's tools (get_nfr / record_result) plus the
  external mcp-k6 server's load-test tools from inside a
  human-initiated claude code interactive session so inference stays on
  the subscription quota rather than the Agent SDK credit pool that
  gates `claude -p` from 2026-06-15.
version: 0.1.0
argument-hint: "(none) - fetches the NFR target from dominator MCP, runs k6 via mcp-k6, records the verdict back to dominator"
allowed-tools:
  - Read
  - Edit
  - Write
  - Bash
  - Grep
  - Glob
  - Agent
  - mcp__dominator__dominator_ping
  - mcp__dominator__dominator_get_nfr
  - mcp__dominator__dominator_record_result
  - mcp__k6__run_script
  - mcp__k6__validate_script
---

# /nfr-judge — dominator NFR judge

Human-initiated entry point. Drives the dominator MCP server's tools
plus the external mcp-k6 server's load-test tools without ever
invoking `claude -p`, so all inference happens inside this
interactive claude code session's subscription quota.

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

1. **Verify MCP wiring**. Call `mcp__dominator__dominator_ping`. The
   tool must return `pong`. If it errors, the dominator MCP server
   is not attached — abort and ask the human to relaunch claude
   with `--mcp-config`.

2. **Fetch the NFR target**. Call `mcp__dominator__dominator_get_nfr`
   with `{"target_id": "<id>"}` (= human supplies the target). The
   tool reads `.pass/config.yaml` and returns the NFR thresholds:

   ```json
   {
     "initialized": true,
     "target_id": "<id>",
     "nfr": {"performance": {...}, "reliability": {...}, "scalability": {...}},
     "target": {"url": "..."},
     "instruction": "Run k6 via mcp-k6, compare results against these thresholds, then call dominator.record_result with verdict='pass' or 'fail'."
   }
   ```

   If `initialized` is `false` (no `.pass/config.yaml` or load error),
   surface the `reason` and abort — the human must run `dominator mcp`
   from a project root with a valid NFR config.

3. **Validate the k6 script**. Call `mcp__k6__validate_script` with the
   script path for the target.

4. **Run the k6 load test**. Call `mcp__k6__run_script` with the
   validated script. Capture stdout / metrics.

5. **Judge pass/fail against thresholds**. The session's reasoning
   compares the captured metrics against the `nfr` thresholds returned
   in step 2. The judgement happens inside the human-initiated claude
   code session — no `claude -p` invocation is allowed.

6. **Record the verdict**. Call
   `mcp__dominator__dominator_record_result` with
   `{"target_id": ..., "verdict": "pass"|"fail", "summary": ...}`.
   The tool persists an `EventJudgmentRecorded` to the event store and
   returns `{"persisted": true, "recorded": true, "persistence": "event-store"}`.
   (If the MCP server was started without a writable `.pass/` dir it
   falls back to `persistence: "preview-only"` with `recorded: false`.)

## What this skill must NOT do

- Invoke `claude -p`, `claude --print`, the Anthropic Agent SDK, or
  any shell wrapper that does so (= refs/issues/0027 §5 billing
  boundary). The repo-wide semgrep gate
  (`.semgrep/jun15-no-headless-llm.yaml`) blocks these patterns in
  production code.
- Auto-trigger inference from a SessionStart hook or any other
  non-human-initiated path. The slash command typed by a human is
  the only valid entry to this workflow.
- Persist a result by writing to `.pass/events/` directly. The
  `dominator.record_result` MCP tool is the canonical path; it
  encapsulates the transactional event-source append (ADR 0005).
- Run k6 directly via `bash` / `Bash` — use `mcp__k6__run_script`
  exclusively so the audit trail (= OTel `messaging.*` attrs) stays
  consistent.

## Done criteria

A `/nfr-judge` run is complete when, in a real claude code session with
both MCP servers attached:

1. `ping` returns `pong` (handshake + tool dispatch verified).
2. `get_nfr` returns `initialized: true` with the NFR thresholds.
3. k6 validate + run succeed via mcp-k6.
4. `record_result` returns `persisted: true` / `persistence: "event-store"`,
   and the verdict shows up in `dominator status` (EventJudgmentRecorded
   read model).

## Related

- Canonical plan: `refs/HTMLification/docs/archive/0027-jun15-mcp-pivot.html`
- Pattern reference:
  - dominator ADR 0003 (`~/tap/dominator/docs/adr/0003-mcp-pivot.md`) — MCP pivot
  - dominator ADR 0004 (`~/tap/dominator/docs/adr/0004-mcp-pivot-k6-adapter-stub.md`) — K6MCPAdapter stub
  - dominator ADR 0005 (`~/tap/dominator/docs/adr/0005-record-result-event-store-wiring.md`) — record_result event store wiring
- Billing boundary table: refs 0027 §5
- Mechanical gate (semgrep rules): refs 0027 §6 + `.semgrep/jun15-no-headless-llm.yaml`
- D-Mail 9-field schema: refs 0027 §8 + `internal/domain/dmail_envelope.go`
