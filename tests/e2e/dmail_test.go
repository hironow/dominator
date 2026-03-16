//go:build e2e

package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestE2E_ArchivePrune_DryRun(t *testing.T) {
	dir := initTestRepo(t)
	writeConfig(t, dir, defaultTestConfig())

	// Place old archive file
	archiveDir := filepath.Join(dir, ".pass", "archive")
	oldTime := time.Now().Add(-365 * 24 * time.Hour)
	content := "---\nname: old-report\nkind: report\ndescription: old\n---\nOld content\n"
	fpath := filepath.Join(archiveDir, "old-report.md")
	if err := os.WriteFile(fpath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	os.Chtimes(fpath, oldTime, oldTime)

	// Dry run should not delete
	stdout, _, err := runCmd(t, dir, "archive-prune", "--days", "30", "--dry-run")
	if err != nil {
		t.Fatalf("archive-prune --dry-run: %v", err)
	}
	_ = stdout

	// File should still exist after dry run
	assertFileExists(t, fpath)
}

func TestE2E_ArchivePrune_PreservesRecent(t *testing.T) {
	dir := initTestRepo(t)
	writeConfig(t, dir, defaultTestConfig())

	// Place recent archive file
	archiveDir := filepath.Join(dir, ".pass", "archive")
	content := "---\nname: recent-report\nkind: report\ndescription: recent\n---\nRecent content\n"
	fpath := filepath.Join(archiveDir, "recent-report.md")
	if err := os.WriteFile(fpath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	// Prune with default days should preserve recent files
	_, _, err := runCmd(t, dir, "archive-prune", "--days", "30", "--yes")
	if err != nil {
		t.Fatalf("archive-prune: %v", err)
	}

	// Recent file should still exist
	assertFileExists(t, fpath)
}

func TestE2E_DMailOutbox_GeneratedByViolation(t *testing.T) {
	// This test verifies the outbox directory structure exists
	// Full violation flow requires k6/Claude — tested in scenario tests
	dir := initTestRepo(t)
	writeConfig(t, dir, defaultTestConfig())

	outboxDir := filepath.Join(dir, ".pass", "outbox")
	entries, err := os.ReadDir(outboxDir)
	if err != nil {
		t.Fatalf("read outbox: %v", err)
	}
	// Initially empty — no violations have been generated
	if len(entries) != 0 {
		t.Errorf("expected empty outbox, got %d entries", len(entries))
	}
}

func TestE2E_DMailArchive_EmptyOnInit(t *testing.T) {
	dir := initTestRepo(t)

	archiveDir := filepath.Join(dir, ".pass", "archive")
	entries, err := os.ReadDir(archiveDir)
	if err != nil {
		t.Fatalf("read archive: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty archive on init, got %d entries", len(entries))
	}
}

func TestE2E_K6Scripts_DirExists(t *testing.T) {
	dir := initTestRepo(t)

	k6Dir := filepath.Join(dir, ".pass", "k6-scripts")
	assertFileExists(t, k6Dir)
}

func TestE2E_Config_Show(t *testing.T) {
	dir := initTestRepo(t)
	writeConfig(t, dir, defaultTestConfig())

	stdout, _, err := runCmd(t, dir, "config", "show")
	if err != nil {
		t.Fatalf("config show: %v", err)
	}
	if !strings.Contains(stdout, "target") {
		t.Errorf("expected 'target' in config output, got: %s", stdout)
	}
}

func TestE2E_Config_Set(t *testing.T) {
	dir := initTestRepo(t)
	writeConfig(t, dir, defaultTestConfig())

	_, _, err := runCmd(t, dir, "config", "set", "lang", "en")
	if err != nil {
		t.Fatalf("config set: %v", err)
	}

	// Verify the change
	stdout, _, err := runCmd(t, dir, "config", "show")
	if err != nil {
		t.Fatalf("config show after set: %v", err)
	}
	if !strings.Contains(stdout, "en") {
		t.Errorf("expected 'en' in config after set, got: %s", stdout)
	}
}
