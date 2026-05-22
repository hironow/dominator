package session_test

import (
	"context"
	"testing"
	"time"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/session"
)

func TestStatus_CountsEventJudgmentRecorded(t *testing.T) {
	// given: a .pass dir with an EventJudgmentRecorded in the event store
	passDir := t.TempDir()
	store := session.NewEventStore(passDir, &domain.NopLogger{})
	agg := domain.NewJudgeAggregate(domain.DefaultConfig())
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)

	ev, err := agg.RecordJudgment("api-v1", "fail", "p95 exceeded", now)
	if err != nil {
		t.Fatalf("RecordJudgment: %v", err)
	}
	if _, err := store.Append(ev); err != nil {
		t.Fatalf("Append: %v", err)
	}

	// when
	report := session.Status(context.Background(), passDir, &domain.NopLogger{})

	// then
	if report.JudgeCount != 1 {
		t.Errorf("JudgeCount = %d, want 1", report.JudgeCount)
	}
	if report.Verdict != domain.VerdictViolation {
		t.Errorf("Verdict = %q, want %q", report.Verdict, domain.VerdictViolation)
	}
	if report.LastJudgment.IsZero() {
		t.Errorf("LastJudgment should be set, got zero")
	}
}
