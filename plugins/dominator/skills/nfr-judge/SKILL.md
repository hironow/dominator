---
name: nfr-judge
description: >-
  Phase 2c slash command for the dominator jun15 MCP pivot
  (refs/issues/0027). Triggers when the user types "/nfr-judge",
  asks to "run a k6 NFR judgement via dominator", "check NFR
  thresholds with dominator", or "test the dominator MCP server
  end-to-end". Drives the dominator MCP server's stub tools (get_nfr /
  record_result) plus the external mcp-k6 server's load-test tools
  from inside a human-initiated claude code interactive session so
  inference stays on the subscription quota rather than the Agent SDK
  credit pool that gates `claude -p` from 2026-06-15.
version: 0.1.0
argument-hint: "(none) - fetches the next NFR target from dominator MCP, runs k6 via mcp-k6, records the verdict back to dominator"
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

# /nfr-judge — dominator MCP pivot Phase 2c

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

## Workflow

1. **Verify MCP wiring**. Call `mcp__dominator__dominator_ping`. The
   tool must return `pong`. If it errors, the dominator MCP server
   is not attached — abort and ask the human to relaunch claude
   with `--mcp-config`.

2. **Fetch the next NFR target**. Call
   `mcp__dominator__dominator_get_nfr` with
   `{"target_id": "<id>"}` (= human supplies the target). During
   Phase 2c the response is a stub:

   ```json
   {
     "stub": true,
     "target_id": "<id>",
     "nfr": null,
     "reason": "phase-2c-mvp: real implementation lands when ...",
     "contract": {"target_id": "string", "metric": "string", "threshold": "number", "comparator": "string (lt|le|gt|ge|eq)", "unit": "string"}
   }
   ```

   While `stub == true`, **do NOT proceed to k6 execution or
   record_result**. Surface the contract descriptor so the human can
   verify the shape, and stop. Real wiring lands in a subsequent
   commit on the `feat/jun15-mcp-pivot` branch.

3. **(Post-stub) Validate the k6 script**. Call
   `mcp__k6__validate_script` with the script path from the NFR
   specification.

4. **(Post-stub) Run the k6 load test**. Call `mcp__k6__run_script`
   with the validated script. Capture stdout / metrics.

5. **(Post-stub) Judge pass/fail against threshold**. The session's
   reasoning compares the captured metric against
   `nfr.threshold` + `nfr.comparator`. The judgement happens
   inside the human-initiated claude code session — no
   `claude -p` invocation is allowed.

6. **(Post-stub) Record the verdict**. Call
   `mcp__dominator__dominator_record_result` with
   `{"target_id": ..., "verdict": "pass"|"fail", "summary": ...}`.
   Phase 2c stub echoes target_id + verdict + summary_length with
   `recorded: false` to signal no side-effect.

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
  `dominator.record_result` MCP tool is the canonical path; that
  tool will encapsulate the transactional event-source append in a
  later commit.
- Run k6 directly via `bash` / `Bash` — use `mcp__k6__run_script`
  exclusively so the audit trail (= OTel `messaging.*` attrs) stays
  consistent.

## Phase 2c MVP exit criteria

This skill is considered Phase 2c MVP complete when:

1. Calling `/nfr-judge` in a real claude code session with both
   MCP servers attached returns the stub responses from steps 1-2
   without error.
2. The `claude_adapter.go` and `doctor.go::CheckMCPK6`
   `claude --print` invocations are removed and the semgrep
   transitional excludes on those two files are deleted (= the
   final commit on the `feat/jun15-mcp-pivot` branch flips the
   lint gate from advisory to enforced).

## Related

- Canonical plan: `refs/HTMLification/docs/issues/0027-jun15-mcp-pivot.html`
- Pattern reference:
  - paintress ADR 0017 (`~/tap/paintress/docs/adr/0017-mcp-pivot.md`)
  - sightjack ADR 0018 (`~/tap/sightjack/docs/adr/0018-mcp-pivot.md`)
  - amadeus ADR 0026 (`~/tap/amadeus/docs/adr/0026-mcp-pivot.md`)
- Billing boundary table: refs 0027 §5
- Mechanical gate (semgrep rules): refs 0027 §6 + `.semgrep/jun15-no-headless-llm.yaml`
- D-Mail 9-field schema: refs 0027 §8 + `internal/domain/dmail_envelope.go`
