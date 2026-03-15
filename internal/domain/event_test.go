package domain_test

import (
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
