package session

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/hironow/dominator/internal/domain"
)

// CheckStateDir verifies that the .pass/ state directory exists.
// If repair is true and the directory is missing, it attempts to create it.
func CheckStateDir(repoRoot string, repair bool) domain.DoctorCheck {
	passRoot := filepath.Join(repoRoot, domain.StateDir)
	info, err := os.Stat(passRoot)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if repair {
				if mkErr := os.MkdirAll(passRoot, 0o755); mkErr != nil {
					return domain.DoctorCheck{
						Name:    "state-dir",
						Status:  domain.CheckFail,
						Message: fmt.Sprintf("repair failed: %v", mkErr),
					}
				}
				return domain.DoctorCheck{
					Name:    "state-dir",
					Status:  domain.CheckFixed,
					Message: fmt.Sprintf("created %s", passRoot),
				}
			}
			return domain.DoctorCheck{
				Name:    "state-dir",
				Status:  domain.CheckFail,
				Message: fmt.Sprintf("%s not found", passRoot),
				Hint:    "run 'dominator init' to create the state directory",
			}
		}
		return domain.DoctorCheck{
			Name:    "state-dir",
			Status:  domain.CheckFail,
			Message: fmt.Sprintf("stat error: %v", err),
		}
	}
	if !info.IsDir() {
		return domain.DoctorCheck{
			Name:    "state-dir",
			Status:  domain.CheckFail,
			Message: fmt.Sprintf("%s exists but is not a directory", passRoot),
		}
	}
	return domain.DoctorCheck{
		Name:    "state-dir",
		Status:  domain.CheckOK,
		Message: fmt.Sprintf("%s exists", passRoot),
	}
}

// CheckConfig verifies that config.yaml is parseable and valid.
func CheckConfig(configPath string) domain.DoctorCheck {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return domain.DoctorCheck{
			Name:    "config-valid",
			Status:  domain.CheckFail,
			Message: fmt.Sprintf("config parse error: %v", err),
			Hint:    "fix config.yaml syntax errors",
		}
	}

	// Check if file exists
	if _, statErr := os.Stat(configPath); errors.Is(statErr, fs.ErrNotExist) {
		return domain.DoctorCheck{
			Name:    "config-valid",
			Status:  domain.CheckWarn,
			Message: "config.yaml not found (using defaults)",
			Hint:    "run 'dominator init' to generate config.yaml",
		}
	}

	errs := domain.ValidateConfig(cfg)
	if len(errs) > 0 {
		return domain.DoctorCheck{
			Name:    "config-valid",
			Status:  domain.CheckFail,
			Message: fmt.Sprintf("config validation errors: %s", errs[0]),
			Hint:    "fix config.yaml values",
		}
	}

	return domain.DoctorCheck{
		Name:    "config-valid",
		Status:  domain.CheckOK,
		Message: "config.yaml is valid",
	}
}
