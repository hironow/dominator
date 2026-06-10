# 0008. Retire the plan-staging commands (check / approve)

**Date:** 2026-06-10
**Status:** Accepted

## Context

`check` (create an execution plan from k6 scripts + NFR thresholds) and
`approve` (mark it approved) existed to gate the Go-owned `run` loop.
The jun15 MCP pivot retired `run` to an error stub — leaving plan
staging with **no consumer**: zero references from skills, phonewave
routing, or any active workflow (refs issue 0034). The
integration-lifecycle suite still assumed the old
init→check→approve→run flow, and one of its tests
(`TestLifecycle_RunUnapprovedPlan`) had been silently broken since the
pivot because `just test-integration` is not part of the commit gate.
docs/conformance.md's claim that the pair "remain local data-plane
helpers" no longer matched reality.

## Decision

1. **`check` / `approve` become error stubs** (same form as
   `generate` / `run` / `validate`), redirecting to the /nfr-judge
   flow. The legacy `--plan-id` flag stays registered so old
   invocations fail with the retirement message, not a flag error.
2. **Replay compatibility is preserved**: `domain.Plan`,
   `EventPlanCreated` and `EventPlanApproved` remain registered event
   types; `status` keeps its read-only plan count from `.run/plans/`.
3. **Orphaned machinery is removed**: `usecase/check.go`,
   `session.PlanStore`, `port.PlanStore` (the config helpers that
   shared the file move to `session/config_store.go` unchanged).
4. **The integration lifecycle is redesigned** around the post-pivot
   flow: init → MCP `record_result` (the real write path, no fakes) →
   event-store assertions → `status`; a dedicated test pins the
   retirement contract for all five retired commands — fixing the
   silently-broken test as a side effect.

## Consequences

### Positive

- The command surface tells the truth; the lifecycle suite tests the
  flow that actually exists; old ledgers still replay.

### Negative

- Operators with scripts calling check/approve get errors (intended —
  the redirect message names the replacement).

### Neutral

- `.run/plans/*.json` files from the pre-pivot era remain readable by
  `status` until a future cleanup.
