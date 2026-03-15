package domain_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/hironow/dominator/internal/domain"
)

func TestNewEvent_HasUUID(t *testing.T) {
	t.Parallel()

	data := domain.ScriptGeneratedData{
		SpecURL:    "https://example.com/spec.json",
		Protocol:   "openapi",
		ScriptPath: "scripts/load.js",
		Endpoints:  5,
	}
	ev, err := domain.NewEvent(domain.EventScriptGenerated, data, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.ID == "" {
		t.Error("expected non-empty ID")
	}
	// UUID v4 format: 8-4-4-4-12
	if len(ev.ID) != 36 {
		t.Errorf("expected UUID format (36 chars), got %d chars: %q", len(ev.ID), ev.ID)
	}
}

func TestNewEvent_HasTimestamp(t *testing.T) {
	t.Parallel()

	now := time.Now()
	data := domain.ScriptGeneratedData{
		SpecURL:  "https://example.com/spec.json",
		Protocol: "openapi",
	}
	ev, err := domain.NewEvent(domain.EventScriptGenerated, data, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
	if !ev.Timestamp.Equal(now) {
		t.Errorf("timestamp = %v, want %v", ev.Timestamp, now)
	}
}

func TestNewEvent_ToolIsDominator(t *testing.T) {
	t.Parallel()

	data := domain.ScriptGeneratedData{
		SpecURL:  "https://example.com/spec.json",
		Protocol: "openapi",
	}
	ev, err := domain.NewEvent(domain.EventScriptGenerated, data, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Tool != "dominator" {
		t.Errorf("Tool = %q, want %q", ev.Tool, "dominator")
	}
}

func TestNewEvent_DataIsMarshaled(t *testing.T) {
	t.Parallel()

	data := domain.GenerationFailedData{
		SpecURL: "https://example.com/spec.json",
		Reason:  "parse error",
	}
	ev, err := domain.NewEvent(domain.EventGenerationFailed, data, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ev.Data) == 0 {
		t.Error("expected non-empty Data")
	}
}

func TestNewEvent_UniqueIDs(t *testing.T) {
	t.Parallel()

	data := domain.ScriptGeneratedData{SpecURL: "https://example.com"}
	ev1, err := domain.NewEvent(domain.EventScriptGenerated, data, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ev2, err := domain.NewEvent(domain.EventScriptGenerated, data, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev1.ID == ev2.ID {
		t.Error("expected unique IDs for different events")
	}
}

func TestValidateEvent_Valid(t *testing.T) {
	t.Parallel()

	data := domain.ScriptGeneratedData{SpecURL: "https://example.com"}
	ev, err := domain.NewEvent(domain.EventScriptGenerated, data, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := domain.ValidateEvent(ev); err != nil {
		t.Errorf("expected valid event, got %v", err)
	}
}

func TestValidateEvent_MissingFields(t *testing.T) {
	t.Parallel()

	ev := domain.Event{}
	err := domain.ValidateEvent(ev)
	if err == nil {
		t.Error("expected error for empty event")
	}
}

func TestNewEvent_AllEventTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		eventType domain.EventType
		data      any
	}{
		{
			name:      "script_generated",
			eventType: domain.EventScriptGenerated,
			data:      domain.ScriptGeneratedData{SpecURL: "https://example.com", Protocol: "openapi", ScriptPath: "load.js", Endpoints: 5},
		},
		{
			name:      "generation_failed",
			eventType: domain.EventGenerationFailed,
			data:      domain.GenerationFailedData{SpecURL: "https://example.com", Reason: "parse error"},
		},
		{
			name:      "plan_created",
			eventType: domain.EventPlanCreated,
			data:      map[string]string{"plan_id": "test-plan"},
		},
		{
			name:      "plan_approved",
			eventType: domain.EventPlanApproved,
			data:      map[string]string{"plan_id": "test-plan"},
		},
		{
			name:      "judged",
			eventType: domain.EventJudged,
			data:      map[string]string{"verdict": "pass"},
		},
		{
			name:      "violation_detected",
			eventType: domain.EventViolationDetected,
			data:      map[string]string{"metric": "p95_latency_ms"},
		},
		{
			name:      "pass_confirmed",
			eventType: domain.EventPassConfirmed,
			data:      map[string]string{"plan_id": "test-plan"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev, err := domain.NewEvent(tt.eventType, tt.data, time.Now())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ev.Type != tt.eventType {
				t.Errorf("Type = %q, want %q", ev.Type, tt.eventType)
			}
			if ev.ID == "" {
				t.Error("expected non-empty ID")
			}
			if ev.Tool != "dominator" {
				t.Errorf("Tool = %q, want %q", ev.Tool, "dominator")
			}
			if len(ev.Data) == 0 {
				t.Error("expected non-empty Data")
			}
		})
	}
}

func TestEvent_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	// given
	now := time.Now().UTC()
	data := domain.ScriptGeneratedData{
		SpecURL:    "https://example.com/spec.json",
		Protocol:   "openapi",
		ScriptPath: "scripts/load.js",
		Endpoints:  5,
	}
	ev, err := domain.NewEvent(domain.EventScriptGenerated, data, now)
	if err != nil {
		t.Fatalf("NewEvent: %v", err)
	}

	// when: marshal then unmarshal
	jsonData, err := json.Marshal(ev)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var restored domain.Event
	if err := json.Unmarshal(jsonData, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// then
	if restored.ID != ev.ID {
		t.Errorf("ID = %q, want %q", restored.ID, ev.ID)
	}
	if restored.Type != ev.Type {
		t.Errorf("Type = %q, want %q", restored.Type, ev.Type)
	}
	if restored.Tool != ev.Tool {
		t.Errorf("Tool = %q, want %q", restored.Tool, ev.Tool)
	}
	if !restored.Timestamp.Equal(ev.Timestamp) {
		t.Errorf("Timestamp = %v, want %v", restored.Timestamp, ev.Timestamp)
	}
	if len(restored.Data) == 0 {
		t.Error("expected non-empty Data after round-trip")
	}

	// Verify data payload survived
	var restoredData domain.ScriptGeneratedData
	if err := json.Unmarshal(restored.Data, &restoredData); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if restoredData.SpecURL != data.SpecURL {
		t.Errorf("SpecURL = %q, want %q", restoredData.SpecURL, data.SpecURL)
	}
	if restoredData.Endpoints != data.Endpoints {
		t.Errorf("Endpoints = %d, want %d", restoredData.Endpoints, data.Endpoints)
	}
}

func TestNewEvent_WithNilData(t *testing.T) {
	t.Parallel()

	// nil data should marshal as "null" which is valid JSON
	ev, err := domain.NewEvent(domain.EventJudged, nil, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.ID == "" {
		t.Error("expected non-empty ID")
	}
	// Data should be "null" (valid JSON)
	if string(ev.Data) != "null" {
		t.Errorf("Data = %q, want %q", string(ev.Data), "null")
	}
}

func TestValidateEvent_IndividualFields(t *testing.T) {
	t.Parallel()

	now := time.Now()
	data := domain.ScriptGeneratedData{SpecURL: "https://example.com"}
	validEvent, _ := domain.NewEvent(domain.EventScriptGenerated, data, now)

	tests := []struct {
		name    string
		mutate  func(*domain.Event)
		wantErr string
	}{
		{
			name:    "missing_id",
			mutate:  func(e *domain.Event) { e.ID = "" },
			wantErr: "ID is required",
		},
		{
			name:    "missing_type",
			mutate:  func(e *domain.Event) { e.Type = "" },
			wantErr: "Type is required",
		},
		{
			name:    "zero_timestamp",
			mutate:  func(e *domain.Event) { e.Timestamp = time.Time{} },
			wantErr: "Timestamp must not be zero",
		},
		{
			name:    "missing_tool",
			mutate:  func(e *domain.Event) { e.Tool = "" },
			wantErr: "Tool is required",
		},
		{
			name:    "empty_data",
			mutate:  func(e *domain.Event) { e.Data = nil },
			wantErr: "Data must not be empty",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := validEvent // copy
			tt.mutate(&ev)
			err := domain.ValidateEvent(ev)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !containsStr(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func containsStr(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestNewEvent_ErrorOnUnmarshalableData(t *testing.T) {
	t.Parallel()

	// channels cannot be JSON-marshaled
	ch := make(chan int)
	_, err := domain.NewEvent(domain.EventJudged, ch, time.Now())
	if err == nil {
		t.Fatal("expected error for unmarshalable data, got nil")
	}
}

func TestEventTypeConstants(t *testing.T) {
	t.Parallel()

	expected := map[domain.EventType]string{
		domain.EventScriptGenerated:  "script.generated",
		domain.EventGenerationFailed: "generation.failed",
		domain.EventPlanCreated:      "plan.created",
		domain.EventPlanApproved:     "plan.approved",
		domain.EventJudged:           "judged",
		domain.EventViolationDetected: "violation.detected",
		domain.EventPassConfirmed:    "pass.confirmed",
	}
	for et, want := range expected {
		if string(et) != want {
			t.Errorf("EventType %v = %q, want %q", et, string(et), want)
		}
	}
}
