package session

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/eventsource"
	"github.com/hironow/dominator/internal/usecase/port"
)

// EnsureRunDir creates the .run/ directory under stateDir if it does not exist.
// Call once before opening stores that write to .run/ (idempotent).
func EnsureRunDir(stateDir string) error {
	runDir := filepath.Join(stateDir, ".run")
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return fmt.Errorf("ensure run dir: %w", err)
	}
	return nil
}

// NewEventStore creates an EventStore backed by daily JSONL files
// in the given stateDir's events/ subdirectory.
// This factory exists so that cmd (composition root) can create an EventStore
// without importing the eventsource package directly (ADR S0008).
func NewEventStore(stateDir string, logger domain.Logger) port.EventStore {
	return eventsource.NewFileEventStore(filepath.Join(stateDir, "events"), logger)
}
