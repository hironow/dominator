# 0004. MCP pivot K6 adapter stub: extend enforcement to K6MCPAdapter

**Date:** 2026-05-22
**Status:** Accepted (extends [0003])

## Context

ADR 0003 (`MCP pivot: claude code session owns LLM, dominator Go CLI
is MCP server data plane`) pinned the architectural invariant
"LLM invocation is always human-initiated inside a claude code
session" and declared the semgrep gate
`.semgrep/jun15-no-headless-llm.yaml` as the mechanical enforcement.

The 2026-05-22 audit recorded in
`tap/refs/HTMLification/docs/issues/0028-jun15-pivot-transmigration-residue-cleanup.html`
discovered that the 0003 enforcement inventory listed
`internal/session/claude_adapter.ClaudeAdapter.Run` and
`internal/session/doctor.CheckMCPK6` as the legacy LLM invocation
sites, and stopped there. A third production-reachable headless LLM
path survived the Phase 2c stub-out:

- `internal/session/k6_mcp_adapter.go::K6MCPAdapter.Run` and
  `K6MCPAdapter.Validate` built `[]string{"--print", "--verbose",
  "--output-format", "stream-json", "--model", a.Model, "-p", prompt}`
  and called `exec.CommandContext(ctx, a.ClaudeCmd, args...)` to
  invoke Claude Code as an intermediary that would then call the
  mcp-k6 `run_script` / `validate_script` tools. Both call sites
  carried `// nosemgrep: ... [permanent]` exemptions on the
  `dangerous-exec-command` rule, justifying them as "ClaudeCmd
  comes from trusted configuration (config.yaml), not user input".

These adapters were the live execution path used by `dominator run`
and `dominator validate`. The architecture document (ADR 0003 and
the `/nfr-judge` skill SKILL.md) already prescribed "attach mcp-k6
directly to the human-initiated session"; the Go CLI was meant to
be a data plane that records results and exposes NFR config, not
an LLM-invoking intermediary. But the K6MCPAdapter had been
left live during Phase 2c, with the `permanent` nosemgrep exemption
acting as a tacit acknowledgement that it predated the pivot
decision and had not yet been retired.

The 0003 semgrep gate (rule `jun15-no-claude-print-exec-go`) used
`pattern-either: exec.CommandContext(..., "claude", "--print", ...)`
literal matching. K6MCPAdapter built args as a `[]string` slice
and spread them into `exec.CommandContext(ctx, a.ClaudeCmd, args...)`,
where neither the binary path nor the `--print` flag were literal
at the `exec.CommandContext` call site. The literal-only pattern
missed the dynamic spread; the `permanent` nosemgrep further
silenced the secondary `dangerous-exec-command` rule that might
otherwise have surfaced the call.

The audit found the same shape in paintress (`claude_adapter.go`
- `doctor.go` + `issues.go`) and produced refs/issues/0028 to
track repo-wide cleanup. paintress ADR 0018 captures the symmetric
decision on the paintress side.

## Decision

The Phase 2c enforcement is extended downward to the K6MCPAdapter
helper layer:

1. **`K6MCPAdapter.Run` and `K6MCPAdapter.Validate` are stub-replaced.**
   The 100-line bodies that built args, spawned Claude as an
   intermediary, parsed the stream-json response, and extracted the
   mcp-k6 tool result are removed. The new bodies return
   `domain.K6Results{}, ErrMCPPivotDeprecated` and
   `ErrMCPPivotDeprecated` respectively. The sentinel is reused
   from `internal/session/claude_adapter.go`. Config fields are
   retained on the struct so existing composition roots
   (`internal/cmd/run.go`, `internal/cmd/validate.go`) compile.

2. **The `permanent` nosemgrep exemptions on the `dangerous-exec-command`
   rule are removed** because the underlying `exec.CommandContext`
   call sites are gone. This eliminates the production-code
   permanent-exemption violation noted in
   refs/issues/0027 §5 ("no `permanent` nosemgrep on production
   paths").

3. **`dominator run` / `dominator validate` cobra subcommands
   continue to invoke usecase.RunJudge / usecase.ValidatePlan,
   which now propagate `ErrMCPPivotDeprecated` via the
   `port.K6Runner` interface contract.** Operators must run
   `/nfr-judge` inside a claude code session with mcp-k6 attached
   directly. The skill SKILL.md already documents this entry point;
   no behavioural change for operators who follow the documented
   workflow.

4. **A new semgrep rule catches the dynamic-spread shape.** The
   rule `jun15-no-print-flag-literal-go` flags any `"--print"`
   string literal in production Go files under `internal/**` /
   `cmd/**` / `main.go`, with `_test.go` / `tests/**` / `.semgrep/**`
   excluded. This catches `args := []string{..., "--print", ...}`
   regardless of how `args` is later spread into `exec.Command`
   or `exec.CommandContext`.

The new rule is applied symmetrically across all five tools
(sightjack / paintress / amadeus / phonewave / dominator) so that
future regressions are caught at the same gate regardless of which
repo introduces them.

## Enforcement inventory (refinement of 0003)

### Entry points (all now stub-protected)

- `internal/cmd/run.go` (`dominator run`) - reaches K6MCPAdapter
  via `usecase.RunJudge`, now propagates `ErrMCPPivotDeprecated`.
- `internal/cmd/validate.go` (`dominator validate`) - reaches
  K6MCPAdapter via `usecase.ValidatePlan`, same propagation.
- `internal/session/claude_adapter.ClaudeAdapter.Run` - stubbed
  in ADR 0003 Phase 2c.
- `internal/session/k6_mcp_adapter.K6MCPAdapter.Run` /
  `K6MCPAdapter.Validate` - stubbed in this ADR.

### Bypass candidates (now blocked)

- Dynamic args spread into `exec.CommandContext(ctx, claudeCmd, args...)`:
  blocked by `jun15-no-print-flag-literal-go`.
- Direct `exec.Command("claude", "--print", ...)` literal: still
  blocked by 0003's `jun15-no-claude-print-exec-go`.
- Shell wrappers: still blocked by 0003's
  `jun15-no-claude-print-shell-wrapper`.
- Anthropic SDK / `ANTHROPIC_API_KEY`: still blocked by 0003's
  `jun15-no-anthropic-*` rules.

### Tests proving coverage

- `internal/session/claude_adapter_test.go::TestClaudeAdapter_RunReturnsErrMCPPivotDeprecated`
  asserts the `ErrMCPPivotDeprecated` sentinel is reachable via
  `errors.Is`. This sentinel is now shared by `K6MCPAdapter` so
  the same assertion path covers both adapters.
- `just semgrep` runs 66 rules with 0 findings, including the new
  `jun15-no-print-flag-literal-go` rule.
- `go test ./...` passes; `usecase/judge_test.go` uses a
  `fakeK6Runner` so usecase-layer coverage is unaffected by the
  K6MCPAdapter body change.

## Consequences

### Positive

- `dominator run` and `dominator validate` consume zero credit
  pool capacity post 2026-06-15. The architectural pin from
  ADR 0003 ("LLM invocation is human-initiated") is now mechanically
  enforced down to the helper layer.
- The `permanent` nosemgrep exemption on production code (which
  refs/issues/0027 §5 explicitly forbade) is removed.
- The new rule is repo-symmetric across all five tools, so a
  future maintainer who copies the same pattern into a different
  tool will hit the gate at PR time.
- The architecture now matches the documented operator workflow:
  attach mcp-k6 directly to a claude code session via
  `--mcp-config`, then invoke `/nfr-judge` - no Claude
  intermediary, no credit pool consumption from dominator.

### Negative

- `dominator run` and `dominator validate` are no longer usable
  as standalone CLI commands. CI / scheduler integrations that
  wrapped these subcommands must shift to the human-initiated
  session pattern.
- The previous architecture (Claude as mcp-k6 intermediary)
  produced a stream-json transcript that captured the full
  reasoning trail; the new direct-attach pattern produces only
  the mcp-k6 tool call/response pair without LLM commentary.
  Operators who relied on the transcript for debugging must
  fall back to claude code's session log.

### Neutral

- `internal/session/k6_summary.go::ParseK6Summary` is preserved.
  This pure function parses k6's native summary JSON and is reused
  by future code paths that ingest mcp-k6 results from elsewhere
  (e.g., a session-level recording).
- The `K6MCPAdapter` struct fields (`ClaudeCmd`, `Model`,
  `TimeoutSec`, `Logger`) remain even though `Run` / `Validate`
  no longer read them; the composition roots
  (`internal/cmd/run.go`, `internal/cmd/validate.go`) still
  construct the adapter.

## References

- [ADR 0003](0003-mcp-pivot.md) - the base architectural pin
  this ADR extends.
- `tap/refs/HTMLification/docs/issues/0028-jun15-pivot-transmigration-residue-cleanup.html`
    - 2026-05-22 audit report and the fix plan this ADR implements.
- `tap/refs/HTMLification/docs/archive/0027-jun15-mcp-pivot.html`
    - canonical multi-tool MCP pivot plan.
- `tap/refs/HTMLification/lessons/0027-jun15-mcp-pivot-post-mortem.html`
    - Pattern 09 ("stub the helpers, not just the entry points")
  documents the lesson behind this ADR.
- `.semgrep/jun15-no-headless-llm.yaml` - the gate that mechanically
  enforces this decision.
- paintress ADR 0018 - symmetric helper-level stub on the paintress
  side (ClaudeAdapter / doctor / issues cobra).
