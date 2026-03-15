package session

import (
	"path/filepath"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/eventsource"
	"github.com/hironow/dominator/internal/usecase/port"
)

// NewEventStore creates an EventStore backed by daily JSONL files
// in the given stateDir's events/ subdirectory.
// This factory exists so that cmd (composition root) can create an EventStore
// without importing the eventsource package directly (ADR S0008).
func NewEventStore(stateDir string, logger domain.Logger) port.EventStore {
	return eventsource.NewFileEventStore(filepath.Join(stateDir, "events"), logger)
}
