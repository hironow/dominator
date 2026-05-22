package usecase

import (
	"fmt"
	"time"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/usecase/port"
)

// judgmentEventEmitter implements port.JudgmentEventEmitter. It wraps the
// JudgeAggregate event production + event store persistence.
// Emit chain: agg.RecordJudgment() → store.Append().
type judgmentEventEmitter struct {
	agg    *domain.JudgeAggregate
	store  port.EventStore
	logger domain.Logger
}

// NewJudgmentEventEmitter creates a JudgmentEventEmitter that wraps the
// aggregate event chain. Used by the cmd composition root to inject a
// persistence-capable emitter into the session MCP server.
func NewJudgmentEventEmitter(agg *domain.JudgeAggregate, store port.EventStore, logger domain.Logger) port.JudgmentEventEmitter {
	return &judgmentEventEmitter{agg: agg, store: store, logger: logger}
}

// EmitJudgmentRecorded produces an EventJudgmentRecorded via the aggregate
// and appends it to the event store.
func (e *judgmentEventEmitter) EmitJudgmentRecorded(targetID, verdict, summary string, now time.Time) error { // nosemgrep: domain-primitives.multiple-string-params-go -- targetID/verdict/summary are semantically distinct fields of the record_result MCP payload [permanent]
	ev, err := e.agg.RecordJudgment(targetID, verdict, summary, now)
	if err != nil {
		return err
	}
	if _, err := e.store.Append(ev); err != nil {
		return fmt.Errorf("append judgment event: %w", err)
	}
	return nil
}
