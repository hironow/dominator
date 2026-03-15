# Testing Strategy

## Test Layers

| Layer | Directory | Build Tag | Dependencies | CI |
|-------|-----------|-----------|-------------|-----|
| Unit | `internal/*/` | none | none | always |
| Integration | `tests/integration/` | none | SQLite | always |
| Contract | `tests/contract/` | `contract` | D-Mail schema validation | always |
| E2E | `tests/e2e/` | `e2e` | Docker, real k6, real Claude Code | manual / nightly |
| Scenario | `tests/scenario/` | `scenario` | fake-claude, fake-gh, all 4+ tool binaries | CI default (L1+L2) |

## Unit Tests

- Located in `internal/*/` alongside production code
- No build tags required
- Minimize mock usage; prefer real code
- Run: `just test`

## Integration Tests

- Located in `tests/integration/`
- Test component interactions with real SQLite, real filesystem
- Covers: event store, NFR evaluator, config loading, plan store, lifecycle
- Run: `just test-integration`

## Contract Tests

- Located in `tests/contract/`
- Build tag: `//go:build contract`
- Validate D-Mail schema compliance and cross-tool protocol contracts
- Run: `just test-contract`

## E2E Tests (Future)

- Located in `tests/e2e/`
- Build tag: `//go:build e2e`
- Docker compose based (`tests/e2e/compose-e2e.yaml`)
- All dependencies must be real — mocks are strictly prohibited
- Requires real k6 binary and real Claude Code CLI
- Run: `just test-e2e` (requires Docker)

## Scenario Tests (Future)

- Located in `tests/scenario/`
- Build tag: `//go:build scenario`
- Requires all sibling tool repos at the same parent directory
- TestMain builds all tool binaries + fake-claude + fake-gh

### Test Levels

| Level | Focus | Timeout |
|-------|-------|---------|
| L1 | Single closed loop | 120s |
| L2 | Multi-issue scenarios | 180s |
| L3 | Concurrent operations | 300s |
| L4 | Fault injection, recovery | 600s |

Run: `just test-scenario` (L1+L2) or `just test-scenario-all`

## Public API Test Policy

Unit tests prefer **external test packages** (`package xxx_test`) over white-box packages (`package xxx`). External tests exercise only the public API surface, which:

- Validates the API contract that external consumers depend on
- Catches accidental API breakage through compilation
- Permits internal refactoring without test changes
- Reduces coupling between tests and implementation details

White-box tests (`package xxx`) are reserved for cases that require access to unexported symbols. Every same-package test file must include a `// white-box-reason:` comment explaining why public API testing is insufficient.

## Quality Command Contract

### Local Commands

| Command | Purpose | Dependencies |
|---------|---------|-------------|
| `just lint` | Full lint pass | vet, semgrep, root-guard, nosemgrep-audit, lint-md |
| `just check` | Pre-commit gate | fmt, vet, semgrep, root-guard, nosemgrep-audit, test, docs-check |
| `just semgrep` | Semgrep ERROR rules | semgrep |
| `just nosemgrep-audit` | Validate nosemgrep tags | grep/awk |
| `just semgrep-test` | Test semgrep rules against fixtures | semgrep |

## Running Tests

```bash
# Unit + integration (default CI)
just test

# Integration only
just test-integration

# Contract tests
just test-contract

# E2E (requires Docker)
just test-e2e

# All semgrep rules
just semgrep
just semgrep-test
just semgrep-warnings
```
