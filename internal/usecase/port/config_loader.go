package port

import "github.com/hironow/dominator/internal/domain"

// ConfigLoader loads dominator configuration from a state directory.
type ConfigLoader interface {
	Load(stateDir string) (domain.Config, error)
}
