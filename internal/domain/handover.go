package domain

import "time"

// HandoverState captures in-progress work state when an operation is
// interrupted by a signal. The struct is pure data — no context, no I/O.
type HandoverState struct { // nosemgrep: first-class-collection.raw-slice-field-domain-go -- transient signal-handler DTO; Completed/Remaining are ordered task sequences with no domain logic, FCC would add ceremony without value [permanent]
	Tool         string // "dominator"
	Operation    string // e.g. "generate", "judge"
	Timestamp    time.Time
	InProgress   string            // Current task description
	Completed    []string          // What was done
	Remaining    []string          // What's left
	PartialState map[string]string // Tool-specific state (key=label, value=detail)
}
