package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hironow/dominator/internal/domain"
)

// WriteLatest writes the judgment result to .pass/.run/latest.json.
// This provides a quick-access snapshot of the most recent run result.
func WriteLatest(stateDir string, judged domain.JudgedData) error {
	dir := filepath.Join(stateDir, ".run")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create .run dir: %w", err)
	}

	data, err := json.MarshalIndent(judged, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal latest: %w", err)
	}

	path := filepath.Join(dir, "latest.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write latest.json: %w", err)
	}
	return nil
}

// ReadLatest reads the most recent judgment result from .pass/.run/latest.json.
func ReadLatest(stateDir string) (*domain.JudgedData, error) {
	path := filepath.Join(stateDir, ".run", "latest.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read latest.json: %w", err)
	}

	var judged domain.JudgedData
	if err := json.Unmarshal(data, &judged); err != nil {
		return nil, fmt.Errorf("parse latest.json: %w", err)
	}
	return &judged, nil
}
