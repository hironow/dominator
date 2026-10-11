# Handover

**Last updated:** 2026-10-12 (JST)

## Current state

The dependency rollup includes OpenTelemetry 1.47.0 and SQLite 1.60.1 from
Dependabot PRs #110 and #112. The application and lint-tool modules require
Go 1.27.2. mise and the E2E Docker image use the same version. CI and release
continue to select Go from the application module.

golangci-lint 2.14.0 reads Go 1.27 export data. The existing lint configuration
and all security checks are unchanged. `just update-lint-go` updates only
that tool and its required dependencies.

## Validation

The original Go 1.26.0 failure was reproduced locally. Go 1.27.2 with the
old lint tool also reproduced its unsupported-export-data failure.
With the compatible tool, `just check` passed: formatting, vet, lint,
Semgrep, root checks, annotation audit, normal tests, and docs checks.
Docker is unavailable locally; the required PR E2E job verifies that coverage.

## Next actions

Merge the rollup only after its applicable CI checks, including Docker E2E,
succeed. Close #110 and #112 once their updates are merged in the rollup.
Future dependency updates must preserve the application/toolchain alignment
and pass `just check` plus PR CI.

## Context

Flatt and Takumi setup steps remain absent. Existing human-owned decisions
in `docs/decision-queue.md` are unchanged.
