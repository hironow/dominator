# Intent

**Last updated:** 2026-06-10
**Requester:** hironow
**Status:** DRAFT — AI が README / git 履歴から起草。requester 未確認
**Work unit:** dominator — MCP server + data plane for NFR judgment (k6 results)

## Goal

Provide an MCP server + data plane for NFR judgment: serve configured NFR thresholds (`dominator.get_nfr`) and record externally judged k6 results (`dominator.record_result`) in an event store. Following the "jun15 MCP pivot", LLM ownership and k6 execution live in a human-initiated Claude Code session (`/nfr-judge` skill + mcp-k6); the Go CLI is the data plane plus supporting local state commands. (Source: README.md)

## Success Criteria

- Existing quality gates pass: `just check` (fmt, vet, lint-go, semgrep, root-guard, nosemgrep-audit, test, docs-check) per the justfile
- Test suites documented in the justfile pass: `just test`, `just test-integration`, `just test-contract`, `just test-e2e` (e2e migrated to testcontainers-go in #49)
- Broader product-level success criteria are otherwise 未定義 — see Open Questions

## Scope

### In scope

(From README)

- MCP server (`dominator mcp`) exposing `dominator.ping` / `dominator.get_nfr` / `dominator.record_result`
- Local state commands: `init`, `check`, `approve`, `insights`, `inbox`, `config`, `doctor`, `status`, `archive-prune`, `clean`, `update`, `version`
- `.pass/` persistent state directory (config, event store, judgment history)

### Out of scope (Non-goals)

(Explicit in README after the MCP pivot)

- Headless generation / validation / run loop — retired; `generate`, `run`, `validate` are redirect stubs
- k6 execution and LLM ownership — moved to the Claude Code session (mcp-k6 + `/nfr-judge`)

## Constraints

(Evident from repo)

- Go module `github.com/hironow/dominator`, Go 1.26
- Design principle (README): load tests require explicit human approval before execution ("Judgment requires approval")
- Tooling pinned via `mise.toml`; lint via `.golangci.yaml`; project semgrep rules in `.semgrep/`

## Open Questions

- [ ] requester による本ドラフトのレビュー
- [ ] Disposition of pending items in `docs/decision-queue.md` (added 2026-06-10, awaiting human review)
- [ ] Long-term fate of the retired redirect subcommands (`generate` / `run` / `validate`) — keep as redirects or remove?
- [ ] Intended users / distribution beyond the author (goreleaser config exists; release audience unverified)
- [ ] Whether lint findings suppressed as "pre-existing" in #53–#56 (gosec / gocyclo / nilerr / govet shadow) should eventually be fixed rather than suppressed
