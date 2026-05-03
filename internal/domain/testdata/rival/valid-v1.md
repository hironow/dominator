# Contract: Enforce checkout p95 latency budget

## Intent
- Keep the `/checkout` endpoint within the agreed latency budget under 100 VU load.
- Success means dominator emits a Pass verdict with no NFR deviations.

## Domain
- Command: judge load test results against NFR thresholds.
- Event: NFR judgment evaluated.
- Read model: NfrConfig + K6Results -> Verdict + deviations list.

## Decisions
- Use existing `domain.EvaluateNfr` for verdict computation.
- Map Rival Contract `nfr.*` evidence keys onto `domain.NfrConfig` fields.

## Steps
1. Map `nfr.p95_latency_ms` evidence onto `Performance.P95LatencyMs`.
   - Target: `internal/domain/rival_contract.go`
   - Acceptance: `EvidenceItemsToNfrConfig` populates the matching field.
2. Map `nfr.target_rps` onto `Scalability.TargetRPS`.
   - Target: `internal/domain/rival_contract.go`
   - Acceptance: int parse succeeds for valid values.

## Boundaries
- Do not change `domain.NfrConfig` shape in this PR.
- Do not introduce runtime behavior changes (Phase 0 is pure code only).
- Do not consume the contract from the judge usecase yet (Phase 4).

## Evidence
- check: just check
- test: just test
- nfr.p95_latency_ms: <= 250
- nfr.error_rate_percent: <= 1
- nfr.success_rate_percent: >= 99
- nfr.target_rps: >= 100
- Maintain a regression test for Pass verdicts under 100 VUs.
