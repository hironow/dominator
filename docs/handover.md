# Handover

**Last updated:** 2026-05-21 (asia/tokyo, Phase 2c kickoff)
**Updated by:** Claude Opus 4.7 session

## Current State

`feat/jun15-mcp-pivot` long-lived branch を切り、 refs/issues/0027
(jun15 MCP pivot v4) の Phase 2c (= dominator horizontal expansion)
を着手。 paintress Phase 1 (PR #213, ADR 0017) + sightjack Phase 2a
(PR #213, ADR 0018) + amadeus Phase 2b (PR #214, ADR 0026) で確立
した 10-11 commit pattern を dominator 用に copy する。

本 commit (= scaffold) で配置済:

- `.semgrep/jun15-no-headless-llm.yaml`: 5 rule + transitional
  exclude on `internal/session/claude_adapter.go` および
  `internal/session/doctor.go` (= 現状 `claude --print` exec を
  保持しているため、 sub-B で MCP 移行と一緒に削除予定)

## In Progress

- branch: `feat/jun15-mcp-pivot` (= long-lived feature branch、 main
  merge は Phase 2c 全完了後)
- linked issue: `refs/HTMLification/docs/issues/0027-jun15-mcp-pivot.html`
- canonical pattern: paintress ADR 0017 + sightjack ADR 0018 +
  amadeus ADR 0026 (= LLM owner inversion、 Go CLI を MCP server
  data plane に縮約)
- Phase 2c MVP scope (= refs 0027 §4 dominator row):
    - [x] feat/jun15-mcp-pivot branch 作成 + scaffold commit (= 本 commit)
    - [ ] MCP server endpoint (= `internal/session/mcp_server.go`) skeleton + `dominator mcp` cobra subcommand
    - [ ] dominator.get_nfr / record_result / ping 等の MCP tool **interface fixed + stub**
    - [ ] `/nfr-judge` slash command の skill definition (= `plugins/dominator/skills/nfr-judge/SKILL.md`)
    - [ ] D-Mail envelope schema 参照 (= paintress canonical を symmetric copy)
    - [ ] **sub-A**: `internal/session/claude_adapter.go` + `internal/session/doctor.go::CheckMCPK6` の `claude --print` invocation を deprecate stub に置換
    - [ ] **sub-B**: semgrep transitional exclude 削除 + skipped test 完全削除
    - [ ] **sub-C**: `docs/adr/0003-mcp-pivot.md` 起票 (= dominator 内 ADR 連番継続) + handover finalize
    - [ ] **sub-D** (post-merge): docs/cli regen + e2e t.Skip if needed

## Next Actions

次 commit で MCP server skeleton 着手:

1. `internal/session/mcp_server.go` を新規実装 (= amadeus 0ee032e を copy + dominator 用 adapt)
2. `internal/cmd/mcp.go` cobra subcommand
3. root.go に `newMCPCommand()` register
4. test 配置

## Known Risks / Blockers

- dominator は **NFR judge** で k6 load test + pass/fail 判定という
  別 domain だが、 LLM 利用箇所 (= claude_adapter.go + doctor.go
  CheckMCPK6) は同一構造なので 9-commit pattern は同じ
- dominator は mcp-k6 (= external MCP server) との連携あり、 これは
  pivot で deprecate するのではなく、 claude code session が
  同 MCP を直接 attach する形に整理する
- ADR 番号: dominator は ADR 0001-0002 のみ、 0003 が次

## Context the Next Actor Needs

- **canonical plan**: `refs/HTMLification/docs/issues/0027-jun15-mcp-pivot.html`
- **paintress ADR 0017**: `~/tap/paintress/docs/adr/0017-mcp-pivot.md`
- **sightjack ADR 0018**: `~/tap/sightjack/docs/adr/0018-mcp-pivot.md`
- **amadeus ADR 0026**: `~/tap/amadeus/docs/adr/0026-mcp-pivot.md`
- **billing boundary 原則**: LLM 発火は常に human-initiated、 daemon は
  route まで、 consume 側は明示 slash command で trigger
- **semgrep gate**: `.semgrep/jun15-no-headless-llm.yaml` 5 rule、
  production path に `permanent` nosemgrep 例外禁止

## Relevant Files and Commands

- `.semgrep/jun15-no-headless-llm.yaml` - billing-boundary gate (5
  rule、 transitional exclude on claude_adapter.go + doctor.go)
- `internal/session/claude_adapter.go` - 現状の LLM invocation entry
  point (= sub-A で deprecate 予定)
- `internal/session/doctor.go` - CheckMCPK6 で `claude --print` 利用
  (= sub-B で MCP ping に置換)
- `just lint` - full lint (vet + semgrep + root-guard + markdown)
- `just semgrep` - semgrep gate (= 0 finding 維持目標)
- `just test` - dominator test suite
