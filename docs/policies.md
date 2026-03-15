# Policy Engine

PolicyEngine dispatches domain events to registered handlers (best-effort, fire-and-forget).

## Location

- Engine: `internal/domain/policy.go`
- Policy definitions: `internal/domain/policy.go` (`Policies` variable)

## Event -> Handler Mapping

| Policy Name | WHEN [EVENT] | THEN [COMMAND] | Description |
|---|---|---|---|
| ScriptGeneratedNotify | script.generated | NotifyGeneration | Notify that a k6 script was generated |
| GenerationFailedAlert | generation.failed | AlertFailure | Alert that script generation failed |

## Domain Events

| Event Type | Payload Type | Description |
|---|---|---|
| `script.generated` | `ScriptGeneratedData` | k6 script created from API spec |
| `generation.failed` | `GenerationFailedData` | Script generation error |
| `plan.created` | `Plan` | Execution plan created |
| `plan.approved` | `Plan` | Plan approved for execution |
| `judged` | `JudgedData` | Judgment result (pass or violation) |
| `violation.detected` | `JudgedData` | NFR threshold exceeded |
| `pass.confirmed` | `JudgedData` | All NFRs within thresholds |

## Dispatch Guarantee

Best-effort (at-most-once). Handler failures are silently logged.
No retry, no dead-letter queue, no error propagation to callers.

## Judgment Flow (COMMAND -> Aggregate -> EVENT)

```
COMMAND: RunJudge(planID)
    |
    v
Aggregate: JudgeAggregate
    |
    +-- Load approved Plan from PlanStore
    +-- Execute k6 via K6Runner
    +-- EvaluateNfr(results, thresholds)
    |
    v
EVENT: judged
    +-- IF violation: EVENT violation.detected
    +-- IF pass:      EVENT pass.confirmed
    |
    v
POLICY: (emit D-Mail, record insights)
```
