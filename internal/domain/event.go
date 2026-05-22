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

	// jun15 MCP pivot Phase 4 follow-up: session-initiated judgment record.
	// Emitted by the dominator.record_result MCP tool when a human-initiated
	// claude code session reports an externally-judged (mcp-k6) verdict.
	// Distinct from EventJudged because it carries no k6 metrics (the
	// session judged externally; only the verdict + summary is recorded).
	EventJudgmentRecorded EventType = "judgment.recorded"
)

// Event is the envelope for all domain events in the event store.
type Event struct { // nosemgrep: structure.multiple-exported-structs-go -- event payload family (Event/ScriptGeneratedData/GenerationFailedData/AppendResult/LoadResult); Event co-locates with its payload DTOs and result metrics as a cohesive event sourcing set [permanent]
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

// ParseEvent validates that an Event has all required fields populated and
// returns the event unchanged when valid. The (Event, error) signature follows
// the Parse-Don't-Validate principle: callers receive a type-level guarantee
// that the returned Event is well-formed.
func ParseEvent(e Event) (Event, error) {
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
		return Event{}, errors.New("invalid event: " + strings.Join(errs, "; "))
	}
	return e, nil
}

// ScriptGeneratedData is the payload for EventScriptGenerated.
type ScriptGeneratedData struct { // nosemgrep: domain-primitives.public-string-field-go, structure.multiple-exported-structs-go -- JSON wire-format event payload; SpecURL is an opaque transport string, not a domain primitive; co-locates with Event/AppendResult/LoadResult as cohesive event sourcing set [permanent]
	SpecURL    string `json:"spec_url"`
	Protocol   string `json:"protocol"`
	ScriptPath string `json:"script_path"`
	Endpoints  int    `json:"endpoints"`
}

// GenerationFailedData is the payload for EventGenerationFailed.
type GenerationFailedData struct { // nosemgrep: domain-primitives.public-string-field-go, structure.multiple-exported-structs-go -- JSON wire-format event payload; SpecURL is an opaque transport string, not a domain primitive; co-locates with Event/AppendResult/LoadResult as cohesive event sourcing set [permanent]
	SpecURL string `json:"spec_url"`
	Reason  string `json:"reason"`
}

// JudgmentRecordedData is the payload for EventJudgmentRecorded. Unlike
// JudgedData (which carries full k6 metrics from a CLI-driven run), this
// payload is lightweight: the session judged externally via mcp-k6 and
// only reports the verdict + a human-readable summary.
type JudgmentRecordedData struct { // nosemgrep: domain-primitives.public-string-field-go, structure.multiple-exported-structs-go -- JSON wire-format event payload; TargetID/Summary are opaque transport strings; co-locates with Event/JudgedData as cohesive event sourcing set [permanent]
	TargetID string  `json:"target_id"`
	Verdict  Verdict `json:"verdict"`
	Summary  string  `json:"summary"`
}

// AppendResult captures metrics from an event store Append operation.
type AppendResult struct { // nosemgrep: structure.multiple-exported-structs-go -- event payload family (Event/ScriptGeneratedData/GenerationFailedData/AppendResult/LoadResult); Event co-locates with its payload DTOs and result metrics as a cohesive event sourcing set [permanent]
	BytesWritten int // total bytes written to event files
}

// LoadResult captures metrics from an event store Load operation.
type LoadResult struct { // nosemgrep: structure.multiple-exported-structs-go -- event payload family (Event/ScriptGeneratedData/GenerationFailedData/AppendResult/LoadResult); Event co-locates with its payload DTOs and result metrics as a cohesive event sourcing set [permanent]
	FileCount        int // number of .jsonl files scanned
	CorruptLineCount int // number of lines skipped due to parse errors
}
