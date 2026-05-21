# Handover

**Last updated:** 2026-05-22 (asia/tokyo, jun15 MCP pivot 0027 archive 入り)
**Updated by:** Claude Opus 4.7 session

## Current State

jun15 MCP pivot (refs/issues/0027) **全 phase 完了 + archive 入り**。
dominator は Phase 2c で MCP server-first architecture を確立し、
Phase 4 follow-up #1-#4 完了後の 2026-05-22 に 0027 が archive (=
`tap/refs/HTMLification/docs/archive/0027-jun15-mcp-pivot.html`)。

dominator 固有の jun15 landmark:

- ADR 0003 (= `docs/adr/0003-mcp-pivot.md`) で architectural pin 固定
- 2 MCP tool 全 real impl (= get_nfr / record_result)
- dual MCP attach pattern: `dominator mcp` (= 自前 server) + `mcp-k6`
  (= k6 公式 server) を claude code session に同時 attach
- `/nfr-judge` skill が claude code session 経由の唯一の judge-driving 経路
- `.semgrep/jun15-no-headless-llm.yaml` 5 rule で headless LLM 経路を
  permanent block
- Phase 4 では dominator は対象外 (= record_result は preview-only marker
  維持。 EventJudgmentRecorded 自動 emit 拡張は将来作業として明示)

## In Progress

なし。 jun15 MCP pivot に関する作業は完了し refs 0027 は archive。

## Next Actions

なし (= Phase 4 #1-#4 全完了)。 後続作業候補は別 issue で fork:

1. **dominator EventJudgmentRecorded 拡張**: paintress Phase 4 #4 と
   同 pattern (= cmd composition root で aggregate + emitter 構築、
   session に port 注入) で record_result を preview-only → event
   store emit へ昇格。 scope やや中規模
2. Phase 3 cost (c) Anthropic dashboard credit 0 verify (= 2026-06-15
   launch 以降の operational evidence)

## Known Risks / Blockers

- `dominator run` 等の LLM-using subcommand は `ErrMCPPivotDeprecated`
  返却に倒れているため、 scheduler / CI で wrap していた job は
  `/nfr-judge` skill 経由に書き換え必要
- dual MCP attach (= dominator + mcp-k6) は claude code session の
  `--mcp-config` で両方を宣言する必要があり、 plugin README で例示済

## Context the Next Actor Needs

- **canonical plan archive**: `tap/refs/HTMLification/docs/archive/0027-jun15-mcp-pivot.html`
- **post-mortem**: `tap/refs/HTMLification/lessons/0027-jun15-mcp-pivot-post-mortem.html`
  (= Pattern 08 で Phase 4 persistence promotion を catalog)
- **billing boundary 原則**: LLM 発火は常に human-initiated、 daemon は route まで
- **semgrep gate**: `.semgrep/jun15-no-headless-llm.yaml` 5 rule、 production
  path に `permanent` nosemgrep 例外禁止
- **dual MCP attach**: dominator は k6 公式 MCP server (= mcp-k6) と
  同時 attach する pattern なので、 claude code session 起動時に `--mcp-config`
  で両 server を宣言

## Relevant Files and Commands

- `docs/adr/0003-mcp-pivot.md` - architectural pin
- `.semgrep/jun15-no-headless-llm.yaml` - billing-boundary gate (5 rule)
- `internal/session/mcp_server.go` - dominator MCP server (2 tool real impl)
- `internal/cmd/mcp.go` - `dominator mcp` cobra subcommand
- `plugins/dominator/skills/nfr-judge/SKILL.md` - human-driven entry point (dual MCP attach 例示)
- `just lint` - semgrep + root-guard + markdownlint (0 issues 維持)
- `just semgrep` - semgrep gate (0 findings 維持)
- `go test ./...` - dominator test suite
