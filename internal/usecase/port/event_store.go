package port

import "github.com/hironow/dominator/internal/domain"

// EventStore persists domain events.
type EventStore interface {
	Append(events ...domain.Event) (domain.AppendResult, error)
}
