package session_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/hironow/dominator/internal/platform"
	"github.com/hironow/dominator/internal/session"
)

func TestInboxReader_ReadAll_WithSampleDMail(t *testing.T) {
	// given
	dir := t.TempDir()
	inboxDir := filepath.Join(dir, "inbox")
	if err := os.MkdirAll(inboxDir, 0o755); err != nil {
		t.Fatalf("create inbox dir: %v", err)
	}

	dmailContent := `---
name: test-feedback
kind: implementation-feedback
severity: high
---

# Feedback

Details here.
`
	if err := os.WriteFile(filepath.Join(inboxDir, "test-feedback.md"), []byte(dmailContent), 0o644); err != nil {
		t.Fatalf("write D-Mail: %v", err)
	}

	logger := platform.NewLogger(os.Stderr, false)
	reader := &session.InboxReader{
		StateDir: dir,
		Logger:   logger,
	}

	// when
	entries, err := reader.ReadAll()

	// then
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Kind != "implementation-feedback" {
		t.Errorf("expected kind 'implementation-feedback', got %q", entries[0].Kind)
	}
	if entries[0].Severity != "high" {
		t.Errorf("expected severity 'high', got %q", entries[0].Severity)
	}
	if entries[0].Suggestion != "dominator generate --force" {
		t.Errorf("expected suggestion 'dominator generate --force', got %q", entries[0].Suggestion)
	}
}

func TestInboxReader_ReadAll_EmptyDir(t *testing.T) {
	// given
	dir := t.TempDir()
	// No inbox dir created

	logger := platform.NewLogger(os.Stderr, false)
	reader := &session.InboxReader{
		StateDir: dir,
		Logger:   logger,
	}

	// when
	entries, err := reader.ReadAll()

	// then
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if entries != nil {
		t.Errorf("expected nil entries for missing dir, got %d entries", len(entries))
	}
}

func TestInboxReader_ReadAll_ConvergenceKind(t *testing.T) {
	// given
	dir := t.TempDir()
	inboxDir := filepath.Join(dir, "inbox")
	if err := os.MkdirAll(inboxDir, 0o755); err != nil {
		t.Fatalf("create inbox dir: %v", err)
	}

	dmailContent := `---
name: convergence-report
kind: convergence
severity: low
---

# Convergence Report
`
	if err := os.WriteFile(filepath.Join(inboxDir, "convergence-report.md"), []byte(dmailContent), 0o644); err != nil {
		t.Fatalf("write D-Mail: %v", err)
	}

	logger := platform.NewLogger(os.Stderr, false)
	reader := &session.InboxReader{
		StateDir: dir,
		Logger:   logger,
	}

	// when
	entries, err := reader.ReadAll()

	// then
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Suggestion != "dominator check" {
		t.Errorf("expected suggestion 'dominator check', got %q", entries[0].Suggestion)
	}
}

func TestInboxReader_Archive_MovesFile(t *testing.T) {
	// given
	dir := t.TempDir()
	inboxDir := filepath.Join(dir, "inbox")
	if err := os.MkdirAll(inboxDir, 0o755); err != nil {
		t.Fatalf("create inbox dir: %v", err)
	}
	filename := "test-mail.md"
	if err := os.WriteFile(filepath.Join(inboxDir, filename), []byte("test"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	logger := platform.NewLogger(os.Stderr, false)
	reader := &session.InboxReader{
		StateDir: dir,
		Logger:   logger,
	}

	// when
	err := reader.Archive(filename)

	// then
	if err != nil {
		t.Fatalf("Archive failed: %v", err)
	}

	// Original should be gone
	if _, err := os.Stat(filepath.Join(inboxDir, filename)); !errors.Is(err, fs.ErrNotExist) {
		t.Error("expected original file to be removed from inbox")
	}

	// Archive should exist
	archivePath := filepath.Join(dir, "archive", "inbox", filename)
	if _, err := os.Stat(archivePath); err != nil {
		t.Errorf("expected archived file at %s: %v", archivePath, err)
	}
}
