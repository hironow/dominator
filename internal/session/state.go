package session

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/hironow/dominator/internal/domain"
	"gopkg.in/yaml.v3"

	"github.com/hironow/dominator/internal/platform"
)

// InitPassDir creates the .pass/ directory structure and writes
// a default config.yaml if one does not already exist.
func InitPassDir(root string, _ domain.Logger) error {
	dirs := []string{
		filepath.Join(root, ".run"),
		filepath.Join(root, "events"),
		filepath.Join(root, "outbox"),
		filepath.Join(root, "inbox"),
		filepath.Join(root, "archive"),
		filepath.Join(root, "k6-scripts"),
		filepath.Join(root, "insights"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}

	// D-Mail routing manifests (refs issue 0031/0035): phonewave derives
	// routes from metadata.produces / metadata.consumes. Canonical source
	// is the embedded template files (template-file parity with the
	// sibling tools); content-compare keeps re-init idempotent and
	// upgrades stale copies.
	for _, name := range []string{"dmail-sendable", "dmail-readable"} {
		destDir := filepath.Join(root, "skills", name)
		if err := os.MkdirAll(destDir, 0o755); err != nil {
			return err
		}
		content, readErr := platform.SkillsFS.ReadFile("templates/skills/" + name + "/SKILL.md")
		if readErr != nil {
			return fmt.Errorf("read embedded manifest %s: %w", name, readErr)
		}
		skillPath := filepath.Join(destDir, "SKILL.md")
		existing, existErr := os.ReadFile(skillPath)
		if existErr != nil || !bytes.Equal(existing, content) {
			if writeErr := os.WriteFile(skillPath, content, 0o644); writeErr != nil {
				return writeErr
			}
		}
	}

	configPath := filepath.Join(root, "config.yaml")
	if err := writeConfigWithDefaults(configPath); err != nil {
		return err
	}

	gitignorePath := filepath.Join(root, ".gitignore")
	requiredEntries := []string{".run/", "outbox/", "inbox/", ".otel.env", "events/"}
	if _, err := os.Stat(gitignorePath); errors.Is(err, fs.ErrNotExist) {
		content := strings.Join(requiredEntries, "\n") + "\n"
		err = os.WriteFile(gitignorePath, []byte(content), 0o644)
		if err != nil {
			return err
		}
	} else if err == nil {
		existing, readErr := os.ReadFile(gitignorePath)
		if readErr == nil {
			var toAdd []string
			for _, entry := range requiredEntries {
				if !strings.Contains(string(existing), entry) {
					toAdd = append(toAdd, entry)
				}
			}
			if len(toAdd) > 0 {
				f, openErr := os.OpenFile(gitignorePath, os.O_APPEND|os.O_WRONLY, 0o644)
				if openErr != nil {
					return openErr
				}
				defer func() { _ = f.Close() }()
				if len(existing) > 0 && !strings.HasSuffix(string(existing), "\n") {
					if _, err := f.Write([]byte("\n")); err != nil {
						return err
					}
				}
				for _, entry := range toAdd {
					if _, err := f.Write([]byte(entry + "\n")); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// writeConfigWithDefaults writes config.yaml with all defaults populated.
// If an existing config.yaml exists, user values are preserved (merged over defaults).
func writeConfigWithDefaults(configPath string) error {
	cfg := domain.DefaultConfig()

	// If config exists, merge user values over defaults
	if data, err := os.ReadFile(configPath); err == nil {
		if unmarshalErr := yaml.Unmarshal(data, &cfg); unmarshalErr != nil {
			return unmarshalErr
		}
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0o644)
}
