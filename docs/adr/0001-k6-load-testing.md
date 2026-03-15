# 0001. k6 Load Testing Engine

**Date:** 2026-03-15
**Status:** Accepted

## Context

Dominator は対象サービスの NFR (非機能要件) を検証するために負荷テストエンジンが必要。

候補:
- **k6** (Grafana): Go エンジン、JS/TS スクリプティング、30k+ GitHub stars
- **Locust**: Python ベース、分散負荷生成に強い
- **Gatling**: Scala/Java ベース、エンタープライズ向け
- **Artillery**: Node.js ベース、YAML 設定駆動

## Decision

**k6 を採用する。**

理由:
1. **Go エンジン**: dominator 自体が Go なので技術スタックの親和性が高い
2. **JS/TS スクリプティング**: v1.0+ で TypeScript ネイティブ対応。Claude Code が生成するスクリプトの型安全性が向上
3. **exec.Command 統合**: プロセス境界で分離するため AGPL-3.0 ライセンスの感染リスクなし
4. **openapi-to-k6**: OpenAPI spec → k6 スクリプト変換の既存ツールが参考になる
5. **Grafana エコシステム**: k6 実行結果を Grafana/SigNoz に送信可能 (下流 viewer)
6. **UNIX stdio 準拠**: `--summary-export` で JSON を stdout に出力可能

## Consequences

### Positive
- Claude Code が生成する k6 スクリプトは JS/TS で IDE サポートが充実
- k6 の `check()` / `threshold` 機能で NFR 判定ロジックをスクリプト内に埋め込み可能
- WebSocket 対応 (`k6/ws` モジュール) で ws-json-rpc プロトコルもカバー

### Negative
- k6 バイナリのインストールが必要 (`doctor` チェックで検出)
- AGPL-3.0 のため、k6 をライブラリとして組み込むことは不可 (exec.Command で回避済み)

### Neutral
- JSON-RPC プロトコルの spec パーサーは k6 に存在しないが、Claude Code がspec を読んで JS を生成するので不要
