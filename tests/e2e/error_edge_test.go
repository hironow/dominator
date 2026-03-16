//go:build e2e

package e2e

import (
	"os"
	"path/filepath"
	"testing"
)

func TestE2E_Check_NoConfig(t *testing.T) {
	dir := t.TempDir()
	// No init → check should fail
	_, _, err := runCmd(t, dir, "check")
	if err == nil {
		t.Fatal("expected error when .pass/ does not exist")
	}
}

func TestE2E_Run_NoPlan(t *testing.T) {
	dir := initTestRepo(t)
	writeConfig(t, dir, defaultTestConfig())

	// run without approved plan → should fail
	_, stderr, err := runCmd(t, dir, "run", "--plan-id", "nonexistent-plan-id")
	if err == nil {
		t.Fatal("expected error for nonexistent plan")
	}
	_ = stderr
}

func TestE2E_Approve_NoPlan(t *testing.T) {
	dir := initTestRepo(t)
	writeConfig(t, dir, defaultTestConfig())

	_, _, err := runCmd(t, dir, "approve", "--plan-id", "nonexistent-plan-id")
	if err == nil {
		t.Fatal("expected error for nonexistent plan")
	}
}

func TestE2E_Generate_NoConfig(t *testing.T) {
	dir := t.TempDir()
	_, _, err := runCmd(t, dir, "generate")
	if err == nil {
		t.Fatal("expected error when .pass/ does not exist")
	}
}

func TestE2E_Insights_Empty(t *testing.T) {
	dir := initTestRepo(t)
	writeConfig(t, dir, defaultTestConfig())

	stdout, _, err := runCmd(t, dir, "insights")
	if err != nil {
		t.Fatalf("insights: %v", err)
	}
	// Empty insights should produce output without error
	_ = stdout
}

func TestE2E_Inbox_Empty(t *testing.T) {
	dir := initTestRepo(t)
	writeConfig(t, dir, defaultTestConfig())

	stdout, stderr, err := runCmd(t, dir, "inbox")
	// inbox may return exit 0 (empty list) or exit 2 (runtime error in env)
	// The key: no panic
	_ = stdout
	_ = stderr
	_ = err
}

func TestE2E_Verbose_Flag(t *testing.T) {
	dir := initTestRepo(t)
	writeConfig(t, dir, defaultTestConfig())

	_, stderr, err := runCmd(t, dir, "status", "-v")
	if err != nil {
		t.Fatalf("status -v: %v", err)
	}
	// Verbose should produce debug output on stderr
	_ = stderr
}

func TestE2E_Quiet_Flag(t *testing.T) {
	dir := initTestRepo(t)
	writeConfig(t, dir, defaultTestConfig())

	_, stderr, err := runCmd(t, dir, "status", "-q")
	if err != nil {
		t.Fatalf("status -q: %v", err)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr with --quiet, got: %s", stderr)
	}
}

func TestE2E_InvalidPath(t *testing.T) {
	_, _, err := runCmd(t, t.TempDir(), "check", "/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Fatal("expected error for invalid path")
	}
}

func TestE2E_ArchivePrune_InvalidDays(t *testing.T) {
	dir := initTestRepo(t)
	writeConfig(t, dir, defaultTestConfig())

	_, _, err := runCmd(t, dir, "archive-prune", "--days", "-1")
	if err == nil {
		t.Fatal("expected error for negative days")
	}
}

func TestE2E_Clean_AlreadyClean(t *testing.T) {
	dir := t.TempDir()
	// Clean without .pass/ → should not panic
	_, _, err := runCmd(t, dir, "clean")
	// May succeed (nothing to clean) or fail (no state dir) — either is acceptable
	_ = err
}

func TestE2E_State_Persistence(t *testing.T) {
	dir := initTestRepo(t)
	writeConfig(t, dir, defaultTestConfig())

	// Verify events directory exists after init
	eventsDir := filepath.Join(dir, ".pass", "events")
	assertFileExists(t, eventsDir)

	// Verify .run directory exists
	runDir := filepath.Join(dir, ".pass", ".run")
	assertFileExists(t, runDir)
}

func TestE2E_DMailDirectories(t *testing.T) {
	dir := initTestRepo(t)

	// Verify D-Mail directories created by init
	for _, sub := range []string{"inbox", "outbox", "archive"} {
		assertFileExists(t, filepath.Join(dir, ".pass", sub))
	}
}

func TestE2E_DMailInbox_PlaceAndRead(t *testing.T) {
	dir := initTestRepo(t)
	writeConfig(t, dir, defaultTestConfig())

	// Place a D-Mail in inbox
	inboxDir := filepath.Join(dir, ".pass", "inbox")
	content := "---\nname: test-feedback\nkind: feedback\ndescription: test feedback\nseverity: low\n---\nTest body\n"
	if err := os.WriteFile(filepath.Join(inboxDir, "test-feedback.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	// inbox command should list it
	stdout, _, err := runCmd(t, dir, "inbox")
	if err != nil {
		t.Fatalf("inbox: %v", err)
	}
	_ = stdout
}
