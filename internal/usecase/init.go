package usecase

import (
	"path/filepath"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/usecase/port"
)

// RunInit delegates state directory creation to the InitRunner port.
// The InitCommand is already valid by construction (parse-don't-validate).
func RunInit(cmd domain.InitCommand, runner port.InitRunner) error {
	stateDir := filepath.Join(cmd.RepoRoot().String(), domain.StateDir)
	return runner.InitPassDir(stateDir)
}
