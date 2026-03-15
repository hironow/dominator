package session

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hironow/dominator/internal/domain"
)

// K6ScriptWriter implements port.ScriptWriter by writing to .pass/k6-scripts/.
type K6ScriptWriter struct {
	StateDir string
	Logger   domain.Logger
}

// Write saves the script content to the k6-scripts directory under StateDir.
// Returns the full path of the written file.
func (w *K6ScriptWriter) Write(scriptName string, content string) (string, error) {
	dir := filepath.Join(w.StateDir, "k6-scripts")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create k6-scripts dir: %w", err)
	}

	if !strings.HasSuffix(scriptName, ".js") && !strings.HasSuffix(scriptName, ".ts") {
		scriptName += ".js"
	}

	path := filepath.Join(dir, scriptName)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("write k6 script: %w", err)
	}
	w.Logger.Info("Generated k6 script: %s", path)
	return path, nil
}
