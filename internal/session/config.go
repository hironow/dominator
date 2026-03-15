package session

import (
	"errors"
	"io/fs"
	"os"

	"github.com/hironow/dominator/internal/domain"
	"gopkg.in/yaml.v3"
)

// LoadConfig reads a YAML configuration file from path.
// If the file does not exist, it returns DefaultConfig with no error.
func LoadConfig(path string) (domain.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return domain.DefaultConfig(), nil
		}
		return domain.Config{}, err
	}
	cfg := domain.DefaultConfig()
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return domain.Config{}, err
	}
	return cfg, nil
}

// SaveConfig writes a Config to a YAML file at path.
func SaveConfig(path string, cfg domain.Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
