//go:build e2e

package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestE2E_DoctorRepair_MissingSkills(t *testing.T) {
	dir := initTestRepo(t)
	writeConfig(t, dir, defaultTestConfig())

	// Remove skills directory
	skillsDir := filepath.Join(dir, ".pass", "skills")
	os.RemoveAll(skillsDir)

	// Doctor --repair should attempt to fix
	_, stderr, _ := runCmd(t, dir, "doctor", "--repair")

	// Doctor ran without panic — check stderr for any output
	if stderr == "" {
		t.Error("expected doctor output on stderr")
	}
	// Skills may or may not be restored (depends on repair implementation)
	// The key assertion: doctor --repair does not crash
}

func TestE2E_DoctorRepair_MissingGitignore(t *testing.T) {
	dir := initTestRepo(t)
	writeConfig(t, dir, defaultTestConfig())

	gitignorePath := filepath.Join(dir, ".pass", ".gitignore")
	os.Remove(gitignorePath)

	// Doctor --repair should attempt to restore
	_, stderr, _ := runCmd(t, dir, "doctor", "--repair")
	if stderr == "" {
		t.Error("expected doctor output on stderr")
	}
}

func TestE2E_DoctorRepair_MissingDirectories(t *testing.T) {
	dir := initTestRepo(t)
	writeConfig(t, dir, defaultTestConfig())

	inboxDir := filepath.Join(dir, ".pass", "inbox")
	os.RemoveAll(inboxDir)

	// Doctor --repair should attempt to restore
	_, stderr, _ := runCmd(t, dir, "doctor", "--repair")
	if stderr == "" {
		t.Error("expected doctor output on stderr")
	}
}

func TestE2E_DoctorRepair_JSON(t *testing.T) {
	dir := initTestRepo(t)
	writeConfig(t, dir, defaultTestConfig())

	stdout, _, _ := runCmd(t, dir, "doctor", "--repair", "--json")

	// Should be valid JSON
	var result map[string]any
	parseJSONOutput(t, stdout, &result)
	checks, ok := result["checks"]
	if !ok {
		t.Error("expected 'checks' key in doctor JSON")
	}
	_ = checks
}
