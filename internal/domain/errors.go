package domain

import (
	"errors"
	"fmt"
)

// SilentError wraps an error whose message has already been printed to stderr
// by the command itself. main.go should suppress output for this error
// while still honouring the exit code via ExitCode.
type SilentError struct{ Err error } // nosemgrep: structure.multiple-exported-structs-go -- error sentinel pair (SilentError/JudgmentError); both implement the error interface and ExitCode references both — co-location is required for the switch logic [permanent]

func (e *SilentError) Error() string { return e.Err.Error() }
func (e *SilentError) Unwrap() error { return e.Err }

// ExitCode maps an error to a process exit code.
//
//	nil              -> 0 (success / NFR pass)
//	JudgmentError    -> 1 (NFR violation)
//	other            -> 2 (runtime error)
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var je *JudgmentError
	if errors.As(err, &je) {
		return 1
	}
	return 2
}

// shutdownKey is the context key for the outer (shutdown) context.
type shutdownKey struct{}

// ShutdownKey is used to embed the outer context in workCtx via context.WithValue.
// Commands retrieve it to get a context that survives workCtx cancellation.
var ShutdownKey = shutdownKey{}

// JudgmentError is returned when NFR judgment fails.
type JudgmentError struct { // nosemgrep: structure.multiple-exported-structs-go -- error sentinel pair (SilentError/JudgmentError); both implement the error interface and ExitCode references both — co-location is required for the switch logic [permanent]
	Violations int
}

func (e *JudgmentError) Error() string {
	return fmt.Sprintf("NFR judgment failed: %d violation(s)", e.Violations)
}
