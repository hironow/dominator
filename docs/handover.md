# Handover

**Last updated:** 2026-05-22 (asia/tokyo, 0028 K6MCPAdapter stub 化 landed)
**Updated by:** Claude Opus 4.7 session

## Current State

jun15 MCP pivot (refs/issues/0027) **全 phase 完了 + archive 入り**、
**かつ 0028 K6MCPAdapter helper-level residue cleanup も完了**。
dominator は Phase 2c で MCP server-first architecture を確立し、
Phase 4 follow-up 完了後の 2026-05-22 に 0027 が archive (=
`tap/refs/HTMLification/docs/archive/0027-jun15-mcp-pivot.html`)、
同日の audit で発覚した K6MCPAdapter residue は ADR 0004 で stub
化完了。

dominator 固有の jun15 landmark:

- ADR 0003 (= `docs/adr/0003-mcp-pivot.md`) で architectural pin 固定
- ADR 0004 (= `docs/adr/0004-mcp-pivot-k6-adapter-stub.md`) で
  K6MCPAdapter helper の stub 完了 (= 0003 §Enforcement inventory が
  ClaudeAdapter + CheckMCPK6 のみ列挙し K6MCPAdapter を見落としていた
  gap を補強)
- 2 MCP tool 全 real impl (= get_nfr / record_result)
- dual MCP attach pattern: `dominator mcp` (= 自前 server) + `mcp-k6`
  (= k6 公式 server) を claude code session に同時 attach
- `/nfr-judge` skill が claude code session 経由の唯一の judge-driving 経路
- 0028 residue cleanup: `K6MCPAdapter.Run / Validate` 100 行を
  ErrMCPPivotDeprecated 返却 stub に圧縮、 `permanent` nosemgrep
  exemption も同時除去 (= 0027 §5 「production path に permanent
  exemption 禁止」 を厳格遵守)
- `.semgrep/jun15-no-headless-llm.yaml` **6 rule** (= 0003 base 5 +
  0004 追加 `jun15-no-print-flag-literal-go`) で headless LLM 経路 +
  dynamic args spread を permanent block (= 5 ツール symmetric)
- ADR 0005 (= `docs/adr/0005-record-result-event-store-wiring.md`) で
  record_result を preview-only → event store persist へ昇格
  (= Phase 4 follow-up、 paintress Phase 4 #4 の emitter-injection pattern)
- record_result event store wiring: `EventJudgmentRecorded` 新 event type と
  `JudgmentRecordedData{TargetID,Verdict,Summary}` payload。
  `JudgeAggregate.RecordJudgment` → `port.JudgmentEventEmitter` →
  `store.Append`。 cmd composition root で emitter 構築 → session 注入
  (= session-no-direct-new-aggregate 遵守)。 status.go read model に
  `EventJudgmentRecorded` case 追加 (= JudgeCount + verdict 反映)

## In Progress

なし。 jun15 MCP pivot に関する作業は完了し refs 0027 は archive。

## Next Actions

なし (= Phase 4 #1-#4 全完了 + EventJudgmentRecorded 拡張 ADR 0005 完了)。
後続作業候補は別 issue で fork:

1. Phase 3 cost (c) Anthropic dashboard credit 0 verify (= 2026-06-15
   launch 以降の operational evidence)
2. (任意) `EventJudgmentRecorded` に対する policy 連鎖 (= WHEN judgment
   THEN command) が必要になれば PolicyEngine handler を追加

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
- `docs/adr/0004-mcp-pivot-k6-adapter-stub.md` - 0003 §Enforcement
  inventory の K6MCPAdapter gap 補強 ADR (= 0028 residue cleanup の
  rationale + permanent nosemgrep 例外撤去)
- `.semgrep/jun15-no-headless-llm.yaml` - billing-boundary gate
  (6 rule、 0004 で `jun15-no-print-flag-literal-go` 追加)
- `internal/session/k6_mcp_adapter.go` - K6MCPAdapter は
  ErrMCPPivotDeprecated 即返却 stub (= 0004 で 100 行 → 60 行に圧縮、
  permanent nosemgrep exemption 撤去)
- `internal/session/mcp_server.go` - dominator MCP server (2 tool real impl)
- `internal/cmd/mcp.go` - `dominator mcp` cobra subcommand
- `plugins/dominator/skills/nfr-judge/SKILL.md` - human-driven entry point (dual MCP attach 例示)
- `just lint` - semgrep + root-guard + markdownlint (0 issues 維持)
- `just semgrep` - semgrep gate (0 findings 維持)
- `go test ./...` - dominator test suite
