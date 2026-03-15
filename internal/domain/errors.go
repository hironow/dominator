package domain

import "fmt"

// SilentError wraps an error whose message has already been printed to stderr
// by the command itself. main.go should suppress output for this error
// while still honouring the exit code via ExitCode.
type SilentError struct{ Err error }

func (e *SilentError) Error() string { return e.Err.Error() }
func (e *SilentError) Unwrap() error { return e.Err }

// ExitCode maps an error to a process exit code.
//
//	nil   -> 0 (success)
//	other -> 1 (runtime error)
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	return 1
}

// IndexEntry holds metadata about an archived D-Mail for the archive index.
type IndexEntry struct {
	Name    string `json:"name"`
	Kind    string `json:"kind"`
	Tool    string `json:"tool"`
	ModTime string `json:"mod_time"`
	Summary string `json:"summary,omitempty"`
}

// shutdownKey is the context key for the outer (shutdown) context.
type shutdownKey struct{}

// ShutdownKey is used to embed the outer context in workCtx via context.WithValue.
// Commands retrieve it to get a context that survives workCtx cancellation.
var ShutdownKey = shutdownKey{}

// JudgmentError is returned when NFR judgment fails.
type JudgmentError struct {
	Violations int
}

func (e *JudgmentError) Error() string {
	return fmt.Sprintf("NFR judgment failed: %d violation(s)", e.Violations)
}
