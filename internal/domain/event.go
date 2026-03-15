package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// EventType identifies the kind of domain event.
type EventType string

const (
	EventScriptGenerated  EventType = "script.generated"
	EventGenerationFailed EventType = "generation.failed"

	// Phase 2: check/approve/run events.
	EventPlanCreated       EventType = "plan.created"
	EventPlanApproved      EventType = "plan.approved"
	EventJudged            EventType = "judged"
	EventViolationDetected EventType = "violation.detected"
	EventPassConfirmed     EventType = "pass.confirmed"
)

// Event is the envelope for all domain events in the event store.
type Event struct {
	ID        string          `json:"id"`
	Type      EventType       `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Tool      string          `json:"tool"`
	Data      json.RawMessage `json:"data,omitempty"`
}

// NewEvent creates a new Event with a UUID, the given timestamp, and marshaled data payload.
func NewEvent(eventType EventType, data any, timestamp time.Time) (Event, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return Event{}, fmt.Errorf("marshal event data: %w", err)
	}
	return Event{
		ID:        uuid.NewString(),
		Type:      eventType,
		Timestamp: timestamp,
		Tool:      "dominator",
		Data:      raw,
	}, nil
}

// ValidateEvent checks that an Event has all required fields populated.
// Returns an error describing all validation failures.
func ValidateEvent(e Event) error {
	var errs []string
	if e.ID == "" {
		errs = append(errs, "ID is required")
	}
	if e.Type == "" {
		errs = append(errs, "Type is required")
	}
	if e.Timestamp.IsZero() {
		errs = append(errs, "Timestamp must not be zero")
	}
	if e.Tool == "" {
		errs = append(errs, "Tool is required")
	}
	if len(e.Data) == 0 {
		errs = append(errs, "Data must not be empty")
	}
	if len(errs) > 0 {
		return errors.New("invalid event: " + strings.Join(errs, "; "))
	}
	return nil
}

// ScriptGeneratedData is the payload for EventScriptGenerated.
type ScriptGeneratedData struct {
	SpecURL    string `json:"spec_url"`
	Protocol   string `json:"protocol"`
	ScriptPath string `json:"script_path"`
	Endpoints  int    `json:"endpoints"`
}

// GenerationFailedData is the payload for EventGenerationFailed.
type GenerationFailedData struct {
	SpecURL string `json:"spec_url"`
	Reason  string `json:"reason"`
}

// AppendResult captures metrics from an event store Append operation.
type AppendResult struct {
	BytesWritten int // total bytes written to event files
}

// LoadResult captures metrics from an event store Load operation.
type LoadResult struct {
	FileCount        int // number of .jsonl files scanned
	CorruptLineCount int // number of lines skipped due to parse errors
}
