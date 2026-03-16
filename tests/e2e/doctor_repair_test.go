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

	// Doctor should detect the issue
	_, stderr, _ := runCmd(t, dir, "doctor")
	if !strings.Contains(stderr, "FAIL") && !strings.Contains(stderr, "skills") {
		t.Logf("doctor stderr: %s", stderr)
	}

	// Doctor --repair should fix it
	_, stderr, err := runCmd(t, dir, "doctor", "--repair")
	if err != nil {
		t.Logf("doctor --repair stderr: %s", stderr)
	}

	// Skills should be restored
	assertFileExists(t, skillsDir)
}

func TestE2E_DoctorRepair_MissingGitignore(t *testing.T) {
	dir := initTestRepo(t)
	writeConfig(t, dir, defaultTestConfig())

	// Remove .gitignore from state dir
	gitignorePath := filepath.Join(dir, ".pass", ".gitignore")
	os.Remove(gitignorePath)

	// Doctor --repair should restore it
	_, _, _ = runCmd(t, dir, "doctor", "--repair")

	assertFileExists(t, gitignorePath)
}

func TestE2E_DoctorRepair_MissingDirectories(t *testing.T) {
	dir := initTestRepo(t)
	writeConfig(t, dir, defaultTestConfig())

	// Remove inbox directory
	inboxDir := filepath.Join(dir, ".pass", "inbox")
	os.RemoveAll(inboxDir)

	// Doctor --repair should restore it
	_, _, _ = runCmd(t, dir, "doctor", "--repair")

	assertFileExists(t, inboxDir)
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
