package cmd_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hironow/dominator/internal/cmd"
)

func TestInboxCmd_SubcommandExists(t *testing.T) {
	// given
	rootCmd := cmd.NewRootCommand()

	// when
	var found bool
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "inbox" {
			found = true
			break
		}
	}

	// then
	if !found {
		t.Fatal("inbox subcommand not found")
	}
}

func TestInboxCmd_FailsWithoutInit(t *testing.T) {
	// given: directory without .pass/
	dir := t.TempDir()
	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"inbox"})

	// when
	err := rootCmd.Execute()

	// then
	if err == nil {
		t.Fatal("expected error when .pass/ not initialized, got nil")
	}
	if !strings.Contains(err.Error(), "dominator init") {
		t.Errorf("expected error to mention 'dominator init', got: %v", err)
	}
}

func TestInboxCmd_EmptyInbox(t *testing.T) {
	// given: directory with .pass/ but no inbox/ dir
	dir := t.TempDir()
	passDir := filepath.Join(dir, ".pass")
	if err := os.MkdirAll(passDir, 0o755); err != nil {
		t.Fatalf("create pass dir: %v", err)
	}
	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetArgs([]string{"inbox"})

	// when
	err := rootCmd.Execute()

	// then: should fail with "inbox empty"
	if err == nil {
		t.Fatal("expected error for empty inbox, got nil")
	}
	if !strings.Contains(err.Error(), "inbox empty") {
		t.Errorf("expected 'inbox empty' error, got: %v", err)
	}
	// stdout should contain empty array
	if !strings.Contains(stdout.String(), "[]") {
		t.Errorf("expected empty array on stdout, got: %s", stdout.String())
	}
}

func TestInboxCmd_ProcessesDMail(t *testing.T) {
	// given: directory with .pass/inbox/ containing a D-Mail
	dir := t.TempDir()
	inboxDir := filepath.Join(dir, ".pass", "inbox")
	if err := os.MkdirAll(inboxDir, 0o755); err != nil {
		t.Fatalf("create inbox dir: %v", err)
	}

	dmailContent := `---
name: impl-feedback-20260101
kind: implementation-feedback
severity: medium
---

# Implementation Feedback

Some feedback content here.
`
	if err := os.WriteFile(filepath.Join(inboxDir, "impl-feedback-20260101.md"), []byte(dmailContent), 0o644); err != nil {
		t.Fatalf("write D-Mail: %v", err)
	}
	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetArgs([]string{"inbox"})

	// when
	err := rootCmd.Execute()

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var entries []struct {
		Name       string `json:"name"`
		Kind       string `json:"kind"`
		Severity   string `json:"severity"`
		Suggestion string `json:"suggestion"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &entries); err != nil {
		t.Fatalf("invalid JSON output: %v\nraw: %s", err, stdout.String())
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Kind != "implementation-feedback" {
		t.Errorf("expected kind 'implementation-feedback', got %q", entries[0].Kind)
	}
	if entries[0].Suggestion != "dominator generate --force" {
		t.Errorf("expected suggestion 'dominator generate --force', got %q", entries[0].Suggestion)
	}

	// Verify file was archived
	archivePath := filepath.Join(dir, ".pass", "archive", "inbox", "impl-feedback-20260101.md")
	if _, err := os.Stat(archivePath); err != nil {
		t.Errorf("expected D-Mail to be archived at %s: %v", archivePath, err)
	}
	// Verify original file was removed
	origPath := filepath.Join(inboxDir, "impl-feedback-20260101.md")
	if _, err := os.Stat(origPath); !os.IsNotExist(err) {
		t.Errorf("expected original D-Mail to be removed from inbox")
	}
}
