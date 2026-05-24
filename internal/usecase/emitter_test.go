package usecase_test

import (
	"testing"
	"time"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/usecase"
)

type fakeEventStore struct {
	events []domain.Event
}

func (f *fakeEventStore) Append(events ...domain.Event) (domain.AppendResult, error) {
	f.events = append(f.events, events...)
	return domain.AppendResult{BytesWritten: 1}, nil
}

func (f *fakeEventStore) LoadAll() ([]domain.Event, domain.LoadResult, error) {
	return f.events, domain.LoadResult{}, nil
}

func (f *fakeEventStore) LoadSince(_ time.Time) ([]domain.Event, domain.LoadResult, error) {
	return f.events, domain.LoadResult{}, nil
}

func TestJudgmentEventEmitter_EmitJudgmentRecorded_AppendsEvent(t *testing.T) {
	// given
	store := &fakeEventStore{}
	agg := domain.NewJudgeAggregate(domain.DefaultConfig())
	emitter := usecase.NewJudgmentEventEmitter(agg, store, &domain.NopLogger{})
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)

	// when
	err := emitter.EmitJudgmentRecorded("api-v1", "pass", "all NFR met", now)

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.events) != 1 {
		t.Fatalf("expected 1 appended event, got %d", len(store.events))
	}
	if store.events[0].Type != domain.EventJudgmentRecorded {
		t.Errorf("event type = %q, want %q", store.events[0].Type, domain.EventJudgmentRecorded)
	}
}

func TestJudgmentEventEmitter_EmitJudgmentRecorded_InvalidVerdict_NoAppend(t *testing.T) {
	// given
	store := &fakeEventStore{}
	agg := domain.NewJudgeAggregate(domain.DefaultConfig())
	emitter := usecase.NewJudgmentEventEmitter(agg, store, &domain.NopLogger{})
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)

	// when
	err := emitter.EmitJudgmentRecorded("api-v1", "maybe", "", now)

	// then
	if err == nil {
		t.Fatalf("expected error for invalid verdict, got nil")
	}
	if len(store.events) != 0 {
		t.Errorf("expected no appended events on validation failure, got %d", len(store.events))
	}
}
