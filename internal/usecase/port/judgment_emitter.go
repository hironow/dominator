package port

import (
	"time"
)

// JudgmentEventEmitter wraps the JudgeAggregate event production + event
// store persistence for session-initiated judgment records (= the
// dominator.record_result MCP tool). Emit chain:
// agg.RecordJudgment() → store.Append().
type JudgmentEventEmitter interface {
	EmitJudgmentRecorded(targetID, verdict, summary string, now time.Time) error // nosemgrep: domain-primitives.multiple-string-params-go -- targetID/verdict/summary are semantically distinct fields of the record_result MCP payload [permanent]
}

// NopJudgmentEventEmitter is a no-op emitter for tests and for the
// preview-only path when the event store is unavailable. It satisfies
// JudgmentEventEmitter without persisting anything.
type NopJudgmentEventEmitter struct{}

// EmitJudgmentRecorded discards the call and returns nil.
func (NopJudgmentEventEmitter) EmitJudgmentRecorded(_, _, _ string, _ time.Time) error { // nosemgrep: domain-primitives.multiple-string-params-go -- Nop implementation of JudgmentEventEmitter interface [permanent]
	return nil
}
