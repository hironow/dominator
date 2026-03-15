package session

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/hironow/dominator/internal/domain"
)

// ValidateStateDir checks that the .pass/ state directory exists and contains
// the required subdirectories.
func ValidateStateDir(repoRoot string) error {
	passRoot := filepath.Join(repoRoot, domain.StateDir)
	info, err := os.Stat(passRoot)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%s not found — run 'dominator init'", passRoot)
		}
		return fmt.Errorf("stat %s: %w", passRoot, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s exists but is not a directory", passRoot)
	}
	return nil
}
