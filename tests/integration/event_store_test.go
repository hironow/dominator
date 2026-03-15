//go:build integration

package integration_test

import (
	"testing"
	"time"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/eventsource"
)

// TestEventStore_AppendAndReplay verifies that appended events can be loaded back.
func TestEventStore_AppendAndReplay(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// given
	dir := t.TempDir()
	store := eventsource.NewFileEventStore(dir, &domain.NopLogger{})

	now := time.Now().UTC()
	ev1, err := domain.NewEvent(domain.EventPlanCreated, map[string]string{
		"plan_id": "test-plan-001",
		"script":  "load-test.js",
	}, now)
	if err != nil {
		t.Fatalf("NewEvent: %v", err)
	}

	ev2, err := domain.NewEvent(domain.EventPlanApproved, map[string]string{
		"plan_id": "test-plan-001",
	}, now.Add(time.Second))
	if err != nil {
		t.Fatalf("NewEvent: %v", err)
	}

	// when: append
	result, appendErr := store.Append(ev1, ev2)
	if appendErr != nil {
		t.Fatalf("Append: %v", appendErr)
	}
	if result.BytesWritten == 0 {
		t.Error("expected BytesWritten > 0")
	}

	// then: load all
	events, loadResult, loadErr := store.LoadAll()
	if loadErr != nil {
		t.Fatalf("LoadAll: %v", loadErr)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if loadResult.FileCount != 1 {
		t.Errorf("expected 1 file, got %d", loadResult.FileCount)
	}
	if loadResult.CorruptLineCount != 0 {
		t.Errorf("expected 0 corrupt lines, got %d", loadResult.CorruptLineCount)
	}

	// then: verify event types and ordering
	if events[0].Type != domain.EventPlanCreated {
		t.Errorf("expected first event type %s, got %s", domain.EventPlanCreated, events[0].Type)
	}
	if events[1].Type != domain.EventPlanApproved {
		t.Errorf("expected second event type %s, got %s", domain.EventPlanApproved, events[1].Type)
	}
	if events[0].Tool != "dominator" {
		t.Errorf("expected tool 'dominator', got %q", events[0].Tool)
	}
}

// TestEventStore_MultipleFiles verifies that events spanning multiple days
// are stored in separate files and loaded in chronological order.
func TestEventStore_MultipleFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// given
	dir := t.TempDir()
	store := eventsource.NewFileEventStore(dir, &domain.NopLogger{})

	// Create events on different dates
	day1 := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	day2 := time.Date(2025, 1, 2, 10, 0, 0, 0, time.UTC)
	day3 := time.Date(2025, 1, 3, 10, 0, 0, 0, time.UTC)

	ev1, err := domain.NewEvent(domain.EventPlanCreated, map[string]string{"day": "1"}, day1)
	if err != nil {
		t.Fatalf("NewEvent day1: %v", err)
	}
	ev2, err := domain.NewEvent(domain.EventPlanApproved, map[string]string{"day": "2"}, day2)
	if err != nil {
		t.Fatalf("NewEvent day2: %v", err)
	}
	ev3, err := domain.NewEvent(domain.EventJudged, map[string]string{"day": "3"}, day3)
	if err != nil {
		t.Fatalf("NewEvent day3: %v", err)
	}

	// when: append all events (they should go to different files)
	if _, err := store.Append(ev1, ev2, ev3); err != nil {
		t.Fatalf("Append: %v", err)
	}

	// then: LoadAll returns all events in order
	events, loadResult, err := store.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
	if loadResult.FileCount != 3 {
		t.Errorf("expected 3 files, got %d", loadResult.FileCount)
	}

	// Verify chronological order
	for i := 1; i < len(events); i++ {
		if !events[i].Timestamp.After(events[i-1].Timestamp) {
			t.Errorf("event[%d] timestamp should be after event[%d]", i, i-1)
		}
	}
}

// TestEventStore_LoadSince verifies that LoadSince filters events correctly.
func TestEventStore_LoadSince(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// given
	dir := t.TempDir()
	store := eventsource.NewFileEventStore(dir, &domain.NopLogger{})

	now := time.Now().UTC()
	ev1, _ := domain.NewEvent(domain.EventPlanCreated, map[string]string{"seq": "1"}, now.Add(-2*time.Hour))
	ev2, _ := domain.NewEvent(domain.EventPlanApproved, map[string]string{"seq": "2"}, now.Add(-1*time.Hour))
	ev3, _ := domain.NewEvent(domain.EventJudged, map[string]string{"seq": "3"}, now)

	if _, err := store.Append(ev1, ev2, ev3); err != nil {
		t.Fatalf("Append: %v", err)
	}

	// when: load events after the first event's timestamp
	events, _, err := store.LoadSince(now.Add(-90 * time.Minute))
	if err != nil {
		t.Fatalf("LoadSince: %v", err)
	}

	// then: should return 2 events (ev2 and ev3)
	if len(events) != 2 {
		t.Fatalf("expected 2 events after cutoff, got %d", len(events))
	}
	if events[0].Type != domain.EventPlanApproved {
		t.Errorf("first event after cutoff should be plan.approved, got %s", events[0].Type)
	}
}

// TestEventStore_EmptyStore verifies LoadAll on an empty/non-existent directory.
func TestEventStore_EmptyStore(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// given: non-existent directory (LoadAll should handle gracefully)
	dir := t.TempDir() + "/nonexistent"
	store := eventsource.NewFileEventStore(dir, &domain.NopLogger{})

	// when
	events, _, err := store.LoadAll()

	// then: no error, empty slice
	if err != nil {
		t.Fatalf("LoadAll on empty store: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}

// TestEventStore_InvalidEventRejected verifies that invalid events are rejected.
func TestEventStore_InvalidEventRejected(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// given
	dir := t.TempDir()
	store := eventsource.NewFileEventStore(dir, &domain.NopLogger{})

	// when: append an invalid event (empty type)
	invalidEvent := domain.Event{
		ID:        "test-id",
		Type:      "", // invalid: empty type
		Timestamp: time.Now(),
		Tool:      "dominator",
		Data:      []byte(`{}`),
	}

	_, err := store.Append(invalidEvent)

	// then: error
	if err == nil {
		t.Fatal("expected error for invalid event")
	}
}
