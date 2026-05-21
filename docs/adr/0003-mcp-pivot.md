# 0003. MCP pivot: claude code session owns LLM, dominator Go CLI is MCP server data plane

**Date:** 2026-05-21
**Status:** Accepted

## Context

Starting 2026-06-15, Anthropic Claude Code subscription plans (Pro,
Max 5x, Max 20x) bill `claude -p` and Agent SDK usage against a
separate monthly Agent SDK credit pool ($20 / $100 / $200) that is
disjoint from the interactive usage quota. The previous dominator
Go CLI architecture invoked `claude --print` as an `exec.CommandContext`
subprocess on every `generate` step via `internal/session/claude_adapter.go`,
plus a separate `claude --print` invocation in `internal/session/doctor.go::CheckMCPK6`
that prompted Claude to list its MCP tools so dominator could detect
whether `mcp-k6` was attached. Every production run after 2026-06-15
would draw on credit pool capacity that is not sized for autonomous
NFR judgement loops.

We surveyed every technically plausible way to keep the existing
control flow running off the interactive quota:

- PTY automation (`creack/pty`, `expect`), tmux `send-keys` +
  `capture-pane`, Remote Control protocol use, and TTY spoofing via
  `script(1)` all violate the Anthropic Acceptable Use Policy clause
  on bypassing product-imposed restrictions, regardless of intent.
- `--output-format`, `--input-format`, `--fallback-model`,
  `--max-budget-usd`, and the rest of the structured-output flag set
  are documented as `--print`-only, so even a successful TTY-spoof
  would still degrade to TUI scraping for any automation-grade
  output.

We also considered keeping the existing Go CLI architecture and
swapping the auth path to a direct Anthropic API key or a third-party
provider (Bedrock / Vertex / Foundry). This works technically but
abandons subscription billing entirely and shifts dominator to a
per-token cost model with no upper bound, which is the opposite of
the design goal that motivated subscription onboarding.

The refs/issues/0027 plan synthesised these constraints into a single
direction: invert the LLM ownership. paintress Phase 1 (ADR 0017)
established the canonical 9-commit pattern; sightjack Phase 2a (ADR
0018) and amadeus Phase 2b (ADR 0026) confirmed the cross-repo
symmetry; dominator Phase 2c applies the same pattern adapted to
the NFR judge / k6 load test pipeline.

## Decision

dominator relinquishes the LLM owner role. From this commit forward,
the architecture is:

1. **Human-initiated claude code interactive session is the LLM
   owner.** All inference happens inside the session's subscription
   quota. No production code path may invoke `claude --print`, import
   the Anthropic Agent SDK, read `ANTHROPIC_API_KEY`, or otherwise
   call the API outside the active session.
2. **dominator Go CLI exposes an MCP server (`dominator mcp`).** The
   server speaks JSON-RPC 2.0 over stdio and registers tools that
   wrap the existing data plane (NFR config, k6 run history, event
   sourcing, D-Mail outbox/inbox, OTel instrumentation). The session
   loads the server via `--mcp-config`. mcp-k6 (the external load
   test execution MCP server) is attached separately to the same
   session, so the data plane (dominator) and execution plane (mcp-k6)
   stay narrowly scoped and independently versioned.
3. **A claude code skill drives the workflow.** The `/nfr-judge`
   slash command (under `plugins/dominator/skills/nfr-judge/SKILL.md`)
   is the only sanctioned entry point. Hooks may emit human-readable
   notices on stderr but must not auto-trigger LLM calls and must not
   surface inbox payloads on stdout (which the official hooks docs
   feed into the session's context).
4. **D-Mail cross-tool messaging follows the paintress canonical
   schema.** `internal/domain/dmail_envelope.go` is a symmetric
   re-implementation of paintress's canonical 9-field envelope so
   dominator can decode incoming envelopes (e.g., NFR check requests
   from amadeus after PR merge gates) without depending on
   paintress's / sightjack's / amadeus's import graphs. The legacy
   D-Mail format (specification / report / feedback / convergence
   kinds) coexists during the pivot transition; a future ADR
   (post-MCP-pivot) reconciles them.
5. **A semgrep gate enforces the boundary.** The rule set in
   `.semgrep/jun15-no-headless-llm.yaml` blocks every executable
   path that could re-introduce `claude --print`, the Agent SDK,
   `ANTHROPIC_API_KEY`, or shell-wrapped variants thereof. `permanent`
   nosemgrep exemptions on this rule are not allowed in production
   paths; the only legitimate exclusion is `tests/**` for the
   fake-claude binary used in test fixtures.

The Go CLI keeps its event sourcing, transactional outbox, OTel
spans, NFR config layer, k6 run history, and domain language. What
it loses is the right to spawn a subprocess that calls the model.

## Enforcement inventory

### Entry points

- `cmd/dominator` Cobra subcommands that previously drove inference:
  `generate`, `run`, `check` (any that depend on `port.ClaudeRunner`).
- `internal/session/claude_adapter.ClaudeAdapter.Run` — the unified
  provider runner that prior to this ADR built a prompt and exec'd
  `claude --print` with stream-json.
- `internal/session/doctor.CheckMCPK6` — the `claude --print`
  invocation that prompted Claude to enumerate MCP servers to detect
  mcp-k6 attachment.
- Any future code that wants to call the model from outside an
  interactive session (Agent SDK, GitHub Actions, third-party SDK
  apps).

### Persistent data carried into the new path

- `internal/domain/dmail_envelope.go` (9-field `DMailEnvelope`) pins
  the cross-tool message schema (`message_id`, `source_tool`,
  `target_tool`, `kind`, `body_path`, `created_at`, `seen_at`,
  `ack_at`, `idempotency_key`). The on-disk layout is
  `inbox/<message_id>.yaml` + `inbox/<message_id>.body.md`.
- `dominator.get_nfr` / `dominator.record_result` MCP tools will
  expose the existing NFR config lookup and k6 run history append
  as MCP resources. Phase 2c ships these as stubs that surface the
  contract descriptor; real wiring lands in a follow-up commit.
- OTel span attributes (`gen_ai.*`, `messaging.*`) continue to flow
  through the MCP server, so the trace topology that previously
  spanned `claude --print` invocations now spans MCP tool calls.

### Bypass candidates

- Direct `exec.Command("claude", "--print", ...)` from Go code —
  blocked by `jun15-no-claude-print-exec-go`.
- Shell wrappers (`bash -lc "claude --print ..."`, `sh -c`, `just`
  recipes, `scripts/*.sh`) — blocked by
  `jun15-no-claude-print-shell-wrapper` and the literal-scan rule.
- Anthropic SDK imports in Go (`github.com/anthropics/anthropic-sdk-go`)
  — blocked by `jun15-no-anthropic-sdk-import-go`.
- `ANTHROPIC_API_KEY` env reads — blocked by
  `jun15-no-anthropic-api-key-read`.
- `SessionStart` / `PreToolUse` hooks that stream inbox content on
  stdout — blocked by the documentation convention (`stderr only`)
  and the `type: prompt` hook prohibition.
- Future `--bare`-mode invocations of `claude` from outside the
  session — covered by the shell-wrapper rule.
- Direct k6 invocation via `bash` / `Bash` from skills — discouraged
  by the skill SKILL.md `must NOT do` section in favour of the
  external `mcp__k6__*` MCP tools, which preserve the audit trail
  (`messaging.*` OTel attrs).

### Tests proving coverage

- `internal/session/mcp_server_test.go` — five tests prove the
  `dominator mcp` stdio server advertises all three Phase 2c tools
  (`dominator.ping` + `get_nfr` + `record_result`), dispatches each
  correctly, and returns the JSON-RPC `-32601` error for unknown
  tools / methods.
- `internal/session/claude_adapter_test.go::TestClaudeAdapter_RunReturnsErrMCPPivotDeprecated`
  — one canonical test proves `ClaudeAdapter.Run` returns
  `session.ErrMCPPivotDeprecated` via `errors.Is`.
- `internal/domain/dmail_envelope_test.go` — five tests cover the
  YAML schema (amadeus -> dominator direction), required-field
  validation, idempotency-key dedup, ack semantics, and the
  `inbox/<id>.yaml` + `body.md` file pair.
- `just semgrep` — 65 rules, 0 findings, including the five
  `jun15-no-headless-llm` gate rules with no production-path
  exclusions (only `tests/**` remains for the fake-claude binary).

## Consequences

### Positive

- Subscription billing keeps paying for all dominator LLM use after
  2026-06-15. Credit pool consumption from dominator is zero by
  construction.
- The Acceptable Use Policy boundary is honoured: every model call
  is human-initiated inside an interactive session.
- The Go CLI's domain plane (event sourcing, SQLite outbox, OTel,
  NFR config, k6 run history, D-Mail semantics) survives intact and
  is now exposed via a stable MCP contract that other tools can
  adopt.
- The semgrep gate makes the boundary mechanical; future contributors
  cannot silently re-introduce headless LLM calls.
- The 9-field `DMailEnvelope` adopted from paintress unifies the
  cross-tool message format across all four tools (paintress,
  sightjack, amadeus, dominator) and any future consumer.
- Separating dominator (data plane) from mcp-k6 (execution plane)
  keeps each MCP server narrowly scoped — operators can swap mcp-k6
  for an alternative load tester (e.g., k6-cloud, fortio) without
  touching dominator.

### Negative

- `dominator generate` (and any subcommand that depends on
  `ClaudeAdapter.Run`) now returns `ErrMCPPivotDeprecated`.
  Operators must launch a claude code session and invoke
  `/nfr-judge` manually. Schedulers and CI jobs that wrapped these
  subcommands no longer work without that human-in-the-loop step.
- The `mcp-k6` doctor check (`CheckMCPK6`) no longer detects whether
  mcp-k6 is attached — that responsibility now lies with the human
  operator passing `--mcp-config`. A follow-up commit will rewire
  the doctor probe to use the dominator MCP server's `dominator.ping`
  plus an optional mcp-k6 tool discovery via the session's own MCP
  client.
- Multi-tool parallel orchestration loses the easy concurrency story
  that came with independent Go processes. A single interactive
  session is the natural unit of work.

### Neutral

- `ClaudeAdapter` struct fields (`ClaudeCmd`, `Model`, `TimeoutSec`,
  `Logger`) are retained even though `Run` no longer reads them,
  because the composition root (`internal/cmd/generate.go`) still
  constructs the adapter and the Phase 2 MCP server tools are
  expected to reuse parts of the config (model name, timeout).
- The `tests/fixtures/dmail/dmail-2026-06-01T13-00-00Z-jkl012.{yaml,body.md}`
  fixture pair carries the amadeus → dominator direction of the
  symmetric 9-field envelope; paintress's, sightjack's, and amadeus's
  fixtures cover the other directions of the same contract.

## References

- refs/issues/0027 — canonical plan including all four codex review
  rounds, the billing boundary table, the mechanical gate, the MVP
  scope reduction, the hook context-injection warning, and the
  D-Mail schema fixation.
- paintress ADR 0017 — canonical 9-commit pattern that originated
  the LLM owner inversion.
- sightjack ADR 0018 — Phase 2a confirmation of the cross-repo
  symmetry on the scan / wave / discuss / apply pipeline.
- amadeus ADR 0026 — Phase 2b confirmation on the review / sync /
  convergence / auto-merge pipeline.
- Local ADRs 0001 (k6-load-testing), 0002 (mcp-k6-integration) —
  the k6 + mcp-k6 architectural decisions this ADR preserves.
- <https://code.claude.com/docs/en/headless> — 2026-06-15 credit
  pool change announcement and `--bare` mode documentation.
- <https://support.claude.com/en/articles/15036540> — per-plan
  credit allocation table.
