package session_test

import (
	"errors"
	"fmt"
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

func TestInboxReader_ReadAll_MalformedFrontmatter(t *testing.T) {
	// given: file without proper YAML frontmatter delimiters
	dir := t.TempDir()
	inboxDir := filepath.Join(dir, "inbox")
	if err := os.MkdirAll(inboxDir, 0o755); err != nil {
		t.Fatalf("create inbox dir: %v", err)
	}

	// No YAML delimiters at all
	if err := os.WriteFile(filepath.Join(inboxDir, "no-delimiters.md"), []byte("Just some text without frontmatter"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	logger := platform.NewLogger(os.Stderr, false)
	reader := &session.InboxReader{StateDir: dir, Logger: logger}

	// when
	entries, err := reader.ReadAll()

	// then
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	// Kind should be empty since no frontmatter
	if entries[0].Kind != "" {
		t.Errorf("expected empty kind for malformed frontmatter, got %q", entries[0].Kind)
	}
}

func TestInboxReader_ReadAll_MissingKind(t *testing.T) {
	// given: frontmatter without kind field
	dir := t.TempDir()
	inboxDir := filepath.Join(dir, "inbox")
	if err := os.MkdirAll(inboxDir, 0o755); err != nil {
		t.Fatalf("create inbox dir: %v", err)
	}

	content := "---\nname: test-no-kind\nseverity: medium\n---\n# No Kind\n"
	if err := os.WriteFile(filepath.Join(inboxDir, "no-kind.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	logger := platform.NewLogger(os.Stderr, false)
	reader := &session.InboxReader{StateDir: dir, Logger: logger}

	// when
	entries, err := reader.ReadAll()

	// then
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Kind != "" {
		t.Errorf("expected empty kind, got %q", entries[0].Kind)
	}
	// Severity should still be parsed
	if entries[0].Severity != "medium" {
		t.Errorf("expected severity 'medium', got %q", entries[0].Severity)
	}
}

func TestInboxReader_ReadAll_AllKinds(t *testing.T) {
	tests := []struct {
		kind       string
		suggestion string
	}{
		{kind: "implementation-feedback", suggestion: "dominator generate --force"},
		{kind: "convergence", suggestion: "dominator check"},
		{kind: "design-feedback", suggestion: "review manually"},
		{kind: "unknown-kind", suggestion: "review manually"},
		{kind: "", suggestion: "review manually"},
	}
	for _, tt := range tests {
		t.Run("kind_"+tt.kind, func(t *testing.T) {
			dir := t.TempDir()
			inboxDir := filepath.Join(dir, "inbox")
			os.MkdirAll(inboxDir, 0o755)

			content := "---\nname: test\nkind: " + tt.kind + "\nseverity: low\n---\n# Test\n"
			os.WriteFile(filepath.Join(inboxDir, "test.md"), []byte(content), 0o644)

			logger := platform.NewLogger(os.Stderr, false)
			reader := &session.InboxReader{StateDir: dir, Logger: logger}

			entries, err := reader.ReadAll()
			if err != nil {
				t.Fatalf("ReadAll: %v", err)
			}
			if len(entries) != 1 {
				t.Fatalf("expected 1 entry, got %d", len(entries))
			}
			if entries[0].Suggestion != tt.suggestion {
				t.Errorf("suggestion = %q, want %q", entries[0].Suggestion, tt.suggestion)
			}
		})
	}
}

func TestInboxReader_ReadAll_SkipsNonMdFiles(t *testing.T) {
	dir := t.TempDir()
	inboxDir := filepath.Join(dir, "inbox")
	os.MkdirAll(inboxDir, 0o755)

	// Write both .md and non-.md files
	os.WriteFile(filepath.Join(inboxDir, "valid.md"), []byte("---\nkind: report\n---\n# Report\n"), 0o644)
	os.WriteFile(filepath.Join(inboxDir, "ignore.txt"), []byte("not a dmail"), 0o644)
	os.WriteFile(filepath.Join(inboxDir, "ignore.json"), []byte("{}"), 0o644)

	logger := platform.NewLogger(os.Stderr, false)
	reader := &session.InboxReader{StateDir: dir, Logger: logger}

	entries, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry (only .md), got %d", len(entries))
	}
}

func TestInboxReader_ReadAll_MultipleDMails(t *testing.T) {
	dir := t.TempDir()
	inboxDir := filepath.Join(dir, "inbox")
	os.MkdirAll(inboxDir, 0o755)

	for i := 0; i < 5; i++ {
		content := fmt.Sprintf("---\nname: mail-%d\nkind: report\nseverity: low\n---\n# Mail %d\n", i, i)
		os.WriteFile(filepath.Join(inboxDir, fmt.Sprintf("mail-%d.md", i)), []byte(content), 0o644)
	}

	logger := platform.NewLogger(os.Stderr, false)
	reader := &session.InboxReader{StateDir: dir, Logger: logger}

	entries, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(entries) != 5 {
		t.Errorf("expected 5 entries, got %d", len(entries))
	}
}

func TestInboxReader_ReadAll_DefaultSeverity(t *testing.T) {
	dir := t.TempDir()
	inboxDir := filepath.Join(dir, "inbox")
	os.MkdirAll(inboxDir, 0o755)

	// No severity field -> should default to "low"
	content := "---\nname: no-severity\nkind: report\n---\n# Report\n"
	os.WriteFile(filepath.Join(inboxDir, "no-sev.md"), []byte(content), 0o644)

	logger := platform.NewLogger(os.Stderr, false)
	reader := &session.InboxReader{StateDir: dir, Logger: logger}

	entries, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Severity != "low" {
		t.Errorf("expected default severity 'low', got %q", entries[0].Severity)
	}
}

func TestInboxReader_Archive_CreatesArchiveDir(t *testing.T) {
	dir := t.TempDir()
	inboxDir := filepath.Join(dir, "inbox")
	os.MkdirAll(inboxDir, 0o755)
	filename := "to-archive.md"
	os.WriteFile(filepath.Join(inboxDir, filename), []byte("test"), 0o644)

	logger := platform.NewLogger(os.Stderr, false)
	reader := &session.InboxReader{StateDir: dir, Logger: logger}

	// Archive dir does not exist yet
	archDir := filepath.Join(dir, "archive", "inbox")
	if _, err := os.Stat(archDir); !errors.Is(err, fs.ErrNotExist) {
		t.Fatal("archive dir should not exist yet")
	}

	err := reader.Archive(filename)
	if err != nil {
		t.Fatalf("Archive: %v", err)
	}

	// Archive dir should now exist
	if _, err := os.Stat(archDir); err != nil {
		t.Errorf("archive dir should exist after Archive: %v", err)
	}
}
