package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hironow/dominator/internal/cmd"
)

func TestApproveCmd_SubcommandExists(t *testing.T) {
	// given
	rootCmd := cmd.NewRootCommand()

	// when
	var found bool
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "approve" {
			found = true
			break
		}
	}

	// then
	if !found {
		t.Fatal("approve subcommand not found")
	}
}

func TestApproveCmd_PlanIdFlagRequired(t *testing.T) {
	// given: initialized .pass/
	dir := t.TempDir()
	passDir := filepath.Join(dir, ".pass")
	if err := os.MkdirAll(passDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"approve"})

	// when
	err := rootCmd.Execute()

	// then
	if err == nil {
		t.Fatal("expected error when --plan-id not provided, got nil")
	}
	if !strings.Contains(err.Error(), "plan-id") {
		t.Errorf("expected error to mention 'plan-id', got: %v", err)
	}
}

func TestApproveCmd_FailsWithoutInit(t *testing.T) {
	// given: directory without .pass/
	dir := t.TempDir()
	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"approve", "--plan-id", "test-id"})

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
