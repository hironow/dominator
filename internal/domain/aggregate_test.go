package domain_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/hironow/dominator/internal/domain"
)

func TestJudgeAggregate_RecordJudgment(t *testing.T) {
	tests := []struct {
		name        string
		targetID    string
		verdict     string
		summary     string
		wantErr     bool
		wantVerdict domain.Verdict
	}{
		{name: "pass verdict produces VerdictPass", targetID: "api-v1", verdict: "pass", summary: "all NFR met", wantErr: false, wantVerdict: domain.VerdictPass},
		{name: "fail verdict produces VerdictViolation", targetID: "api-v1", verdict: "fail", summary: "p95 exceeded", wantErr: false, wantVerdict: domain.VerdictViolation},
		{name: "invalid verdict is rejected", targetID: "api-v1", verdict: "maybe", summary: "", wantErr: true},
		{name: "empty target_id is rejected", targetID: "", verdict: "pass", summary: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			agg := domain.NewJudgeAggregate(domain.DefaultConfig())
			now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)

			// when
			ev, err := agg.RecordJudgment(tt.targetID, tt.verdict, tt.summary, now)

			// then
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ev.Type != domain.EventJudgmentRecorded {
				t.Errorf("event type = %q, want %q", ev.Type, domain.EventJudgmentRecorded)
			}
			if _, parseErr := domain.ParseEvent(ev); parseErr != nil {
				t.Errorf("ParseEvent: %v", parseErr)
			}
			var data domain.JudgmentRecordedData
			if uErr := json.Unmarshal(ev.Data, &data); uErr != nil {
				t.Fatalf("unmarshal data: %v", uErr)
			}
			if data.Verdict != tt.wantVerdict {
				t.Errorf("verdict = %q, want %q", data.Verdict, tt.wantVerdict)
			}
			if data.TargetID != tt.targetID {
				t.Errorf("target_id = %q, want %q", data.TargetID, tt.targetID)
			}
			if data.Summary != tt.summary {
				t.Errorf("summary = %q, want %q", data.Summary, tt.summary)
			}
		})
	}
}
