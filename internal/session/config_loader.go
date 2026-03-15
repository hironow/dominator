package session

import (
	"path/filepath"

	"github.com/hironow/dominator/internal/domain"
)

// FileConfigLoader implements port.ConfigLoader by reading from the filesystem.
type FileConfigLoader struct{}

// Load reads config.yaml from the given state directory.
func (l *FileConfigLoader) Load(stateDir string) (domain.Config, error) {
	cfgPath := filepath.Join(stateDir, domain.ConfigFile)
	return LoadConfig(cfgPath)
}
