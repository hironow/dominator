# Handover

**Last updated:** 2026-05-21 (asia/tokyo, Phase 2c finalize)
**Updated by:** Claude Opus 4.7 session

## Current State

`feat/jun15-mcp-pivot` long-lived branch 上で refs/issues/0027
(jun15 MCP pivot v4) の Phase 2c (= dominator horizontal expansion)
を 7 commit (scaffold + skeleton + skill + envelope + sub-A + sub-B +
sub-C) 完了。 paintress Phase 1 (PR #213, ADR 0017) + sightjack
Phase 2a (PR #213, ADR 0018) + amadeus Phase 2b (PR #214, ADR 0026)
で確立した 10-11 commit pattern を dominator 用に adapt し、
`claude --print` exec を全 production path から削除、 `dominator mcp`
stdio MCP server + `/nfr-judge` skill + 9-field D-Mail envelope schema
を追加、 dominator ADR 0003 を発行。

Phase 2c 完了内容:

1. **`.semgrep/jun15-no-headless-llm.yaml`** (= 5 rule, scaffold で
   transitional exclude を設定し sub-B で削除済。 残る exclude は
   `tests/**` のみ = fake-claude binary を呼ぶ test fixture 用)。
2. **`dominator mcp` MCP server** (`internal/session/mcp_server.go`)
   = JSON-RPC 2.0 stdio、 4 MiB scanner buffer、 Phase 2c MVP として
   `dominator.ping` / `dominator.get_nfr` / `dominator.record_result`
   を advertise + dispatch。 後 2 つは contract 固定 + stub。 5 test
   pass。
3. **`/nfr-judge` skill** (= `plugins/dominator/skills/nfr-judge/SKILL.md`)
   + plugin README。 `--plugin-dir ./plugins/dominator` で claude code
   session に load、 `mcp__dominator__*` + `mcp__k6__*` (external
   mcp-k6) を allowed-tools に宣言。 dual MCP attach (= dominator
   data plane + mcp-k6 execution plane)。
4. **D-Mail 9-field envelope** (`internal/domain/dmail_envelope.go`)
   = paintress canonical の symmetric copy (= amadeus → dominator
   方向 nfr_check_request 系を含む全方向の parse / validate / ack
   を支える)。
   `tests/fixtures/dmail/dmail-2026-06-01T13-00-00Z-jkl012.{yaml,body.md}`
   で fixture pair を配置 + 5 test pass。 helper は同 package の
   config_test.go::contains との衝突回避で envelopeErrContains 命名。
5. **sub-A** (= claude_adapter.go + doctor.go::CheckMCPK6 の
   `claude --print` invocation を `ErrMCPPivotDeprecated` stub に
   置換)。 ClaudeAdapter.Run body 全削除 + 不要 imports (context /
   fmt / os/exec / strings) 削除、 struct field は composition root
   互換のため保持。 CheckMCPK6 は Skip 結果 (`post jun15 MCP pivot,
   refs/issues/0027`) に置換、 stream-json parser / mcp-k6 detection
   ロジック + bytes / io / time / platform imports 削除。
6. **sub-B** (= semgrep transitional excludes 削除 + canonical
   assertion 追加)。 deprecated 削除対象 test は元から存在しない
   (= dominator は claude_adapter_test.go を持たなかった)、 代わりに
   canonical assertion (`TestClaudeAdapter_RunReturnsErrMCPPivotDeprecated`、
   `errors.Is` で `session.ErrMCPPivotDeprecated` を検証) を 1 件
   新規追加。
7. **sub-C** (= 本 commit、 dominator ADR 0003 起票 + handover finalize)。

## In Progress

+ branch: `feat/jun15-mcp-pivot` (= scaffold + 6 commit、 7 commit 目
  が本 sub-C、 sub-D は必要時 post-merge fixup)
+ main merge は Phase 2c 完了後の PR 作成 + CI green + squash-merge
  待ち (= paintress / sightjack / amadeus PR pattern)
+ 次 phase: なし (= phonewave は LLM 非使用のため対象外、 Phase 2c
  完了で 4 ツール (paintress / sightjack / amadeus / dominator) の
  jun15 MCP pivot 横展開が全完了)

## Next Actions

1. `feat/jun15-mcp-pivot` に PR 作成 (= title: `feat(session):
   Phase 2c dominator jun15 MCP pivot (refs/issues/0027)`)
2. CI を green まで監視 (= sightjack / amadeus PR と同様、
   docs-check と e2e fail が発生する可能性あり、 必要なら sub-D
   fixup)
3. squash-merge 完了後、 refs/issues/0027 を Phase 2 全完了に更新
4. cost monitoring: 全 4 ツール統合 OTel MCP invocation count を
   計測、 Anthropic dashboard で credit 0 維持を手動検証
5. 後続 phase: NFR config + k6 run history wiring (= MVP stubs を
   real implementation に置換)、 Phase 3 finalize (= 公式 docs /
   migration guide / lessons learned 集約)

## Known Risks / Blockers

+ paintress / sightjack / amadeus で post-merge sub-D が必要だった
  patterns: docs-check (CLI doc 未生成) + e2e tests (deprecated CLI
  経由)。 dominator も同様の fixup が CI で必要になる可能性。
+ `dominator generate` 等 LLM-using subcommand が `ErrMCPPivotDeprecated`
  返却に倒れたため、 既存 NFR judgement workflow が claude code
  session 経由になる (= human-in-the-loop 必須化)。
+ `CheckMCPK6` doctor probe が Skip に倒れたため、 operator が
  `--mcp-config` で mcp-k6 を attach し忘れた場合の早期検出が
  失われる。 後続 phase で dominator MCP server の health probe に
  rewire 予定。

## Context the Next Actor Needs

+ **canonical plan**: `refs/HTMLification/docs/issues/0027-jun15-mcp-pivot.html`
+ **paintress ADR 0017**: `~/tap/paintress/docs/adr/0017-mcp-pivot.md`
+ **sightjack ADR 0018**: `~/tap/sightjack/docs/adr/0018-mcp-pivot.md`
+ **amadeus ADR 0026**: `~/tap/amadeus/docs/adr/0026-mcp-pivot.md`
+ **dominator ADR 0003**: `docs/adr/0003-mcp-pivot.md` (= 本 phase の
  architectural pin、 3 ツール ADR の symmetric counterpart)
+ **billing boundary 原則**: LLM 発火は常に human-initiated、 daemon
  は route まで、 consume 側は明示 slash command で trigger
+ **semgrep gate**: `.semgrep/jun15-no-headless-llm.yaml` 5 rule、
  production path に `permanent` nosemgrep 例外禁止、 残る exclude は
  `tests/**` (fake-claude binary) のみ
+ **MCP server tool 命名規約**: `<tool_name>.<verb>` (= dot 区切り、
  paintress の `paintress.next_issue` / sightjack の
  `sightjack.next_wave` / amadeus の `amadeus.next_review` と対称)。
  claude code 側の `mcp__<server>__<tool>` 自動 mapping に対応。
+ **dual MCP attach pattern**: dominator は dominator (data plane)
    + mcp-k6 (execution plane) の 2 server attach が前提。 skill SKILL.md
  の Prerequisites を参照。

## Relevant Files and Commands

+ `docs/adr/0003-mcp-pivot.md` - 本 phase の architectural pin
+ `.semgrep/jun15-no-headless-llm.yaml` - billing-boundary gate (5
  rule、 production scope 完全 enforced、 `tests/**` のみ exclude)
+ `internal/session/mcp_server.go` - dominator MCP server (= Phase 2c
  MVP scope、 3 tool stub)
+ `internal/session/claude_adapter.go` - `ClaudeAdapter.Run` =
  `ErrMCPPivotDeprecated` stub
+ `internal/session/doctor.go` - `CheckMCPK6` = Skip (post jun15
  MCP pivot)
+ `internal/domain/dmail_envelope.go` - 9-field envelope symmetric
  copy
+ `internal/cmd/mcp.go` - `dominator mcp` cobra subcommand
+ `plugins/dominator/skills/nfr-judge/SKILL.md` - human-driven
  entry point
+ `tests/fixtures/dmail/dmail-2026-06-01T13-00-00Z-jkl012.{yaml,body.md}`
    + synthetic D-Mail contract fixture (amadeus → dominator 方向、
  nfr_check_request kind)
+ `just lint` - full lint (vet + semgrep + root-guard + markdown、
  全 pass)
+ `just test` - dominator test suite (= 全 pkg ok)
+ `just semgrep` - semgrep gate (= 0 findings 維持、 65 rules)
