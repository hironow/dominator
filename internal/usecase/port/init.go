package port

// InitRunner handles state directory initialization I/O.
type InitRunner interface {
	InitPassDir(stateDir string) error
}
