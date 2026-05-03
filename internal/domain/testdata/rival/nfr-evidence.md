# Contract: NFR evidence focus fixture

## Intent
- Provide a fixture exercising every supported `nfr.*` evidence key.
- Mix in prose bullets and unsupported keys to verify they are ignored.

## Domain
- Command: parse evidence into deterministic items.
- Event: evidence parsed.
- Read model: NfrConfig populated from evidence.

## Decisions
- Parser MUST drop unknown `nfr.*` keys silently.
- Parser MUST drop free-form prose bullets silently.

## Steps
1. Cover all four supported NFR keys.
   - Target: `internal/domain/rival_contract.go`
   - Acceptance: all four populate the NfrConfig output.
2. Confirm prose bullets do not appear in the parsed items list.
   - Target: `internal/domain/rival_contract_test.go`
   - Acceptance: parsed items length equals four supported entries.

## Boundaries
- Do not include non-NFR evidence kinds (check/test/lint/semgrep) in this fixture.
- Do not exercise contract metadata parsing here.

## Evidence
- nfr.p95_latency_ms: <= 300
- nfr.error_rate_percent: <= 0.5
- nfr.success_rate_percent: >= 99.9
- nfr.target_rps: >= 250
- nfr.unknown_metric: <= 10
- Keep prose bullets out of the structured NFR mapping.
