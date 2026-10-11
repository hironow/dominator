# Handover

**Last updated:** 2026-10-11 (JST)
**Updated by:** Codex

## Current State

CI and release no longer use the five Takumi Guard steps.
Go uses its standard module download path.
CI installs Semgrep from official PyPI with `just install-semgrep` and uv.
The application code, Go dependencies, tests, audits, and release attestations are unchanged.
PR jobs skip drafts and run when marked ready for review.

## In Progress

Branch: `ci/remove-flatt`.
`GOTOOLCHAIN=go1.26.1 just check` passed.
Go lint and Semgrep reported zero findings; all normal tests and docs checks passed.
PR CI must pass before squash merge.

## Next Actions

1. Create a ready PR and verify every CI check.
2. Squash merge and remove the two Flatt entries from the Actions allowlist.

## Known Risks / Blockers

The host uses Go 1.27.2.
Use Go 1.26.1 for local verification, matching the repository's Go 1.26 CI toolchain.
CI must verify Docker E2E tests; Docker is unavailable locally.

## Previous handover

**Last updated:** 2026-06-10 (JST)
**Updated by:** claude (AI draft from git history — review before trusting)

### Current State

Dominator is a Go MCP server + data plane for NFR judgment. The "jun15 MCP pivot" is complete and documented: the binary serves NFR thresholds from `.pass/config.yaml` and records pass/fail verdicts as `EventJudgmentRecorded`; k6 execution and LLM work happen in a Claude Code session via mcp-k6 and the `/nfr-judge` skill. Recent history is mostly consolidation: MCP-pivot doc alignment (#58–#61), lint-suppression cleanups (#53–#56), session close hardening (#52), and e2e migration to testcontainers-go (#49). Last commit: `e72dec8` "docs: add decision queue for human-review items (#64)" on 2026-06-10.

### In Progress

- Items in `docs/decision-queue.md` are explicitly awaiting human review (added 2026-06-10)
- No other in-flight work is evident from the shallow git history (only `main` branch visible)

### Next Actions

1. requester による docs/intent.md ドラフトのレビューと確定
2. Review and resolve the human-review items in `docs/decision-queue.md`

### Known Risks / Blockers

- Several lint findings were suppressed as "pre-existing false positives / violations" (#53–#56: gosec, gocyclo, nilerr, govet shadow) rather than fixed — verify the suppressions stay justified

### Context the Next Actor Needs

- `just` is the task entry point (`default: help`); `just check` mirrors the full gate (fmt, vet, lint-go, semgrep, root-guard, nosemgrep-audit, test, docs-check)
- Tool versions are pinned in `mise.toml`; project-specific semgrep rules live in `.semgrep/`
- `docs/` carries substantial reference material (testing.md, policies.md, dmail-protocol-conventions.md, stdio-convention.md, rival-contract-v1.md, adr/, shared-adr/) — read before changing conventions
- The PSYCHO-PASS naming scheme in the README maps directly to components (.pass/ state dir, hue, coefficient, D-Mail severity)

### Relevant Files and Commands

- `justfile` — all dev tasks; `just check` is the pre-commit gate
- `cmd/` and `internal/` — CLI wiring and core layers
- `docs/decision-queue.md` — pending human-review items
- `docs/README.md` — documentation index
- `dominator mcp` — start the MCP data-plane server; `dominator init` — initialize `.pass/`
