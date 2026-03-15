package session

import "github.com/hironow/dominator/internal/domain"

// InitAdapter implements port.InitRunner by delegating to session.InitPassDir.
type InitAdapter struct {
	Logger domain.Logger
}

// InitPassDir creates the state directory structure.
func (a *InitAdapter) InitPassDir(stateDir string) error {
	return InitPassDir(stateDir, a.Logger)
}
