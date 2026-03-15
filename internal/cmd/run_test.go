package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hironow/dominator/internal/cmd"
	"github.com/spf13/cobra"
)

func TestRunCmd_SubcommandExists(t *testing.T) {
	// given
	rootCmd := cmd.NewRootCommand()

	// when
	var found bool
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "run" {
			found = true
			break
		}
	}

	// then
	if !found {
		t.Fatal("run subcommand not found")
	}
}

func TestRunCmd_PlanIdRequired(t *testing.T) {
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
	rootCmd.SetArgs([]string{"run"})

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

func TestRunCmd_FailsWithoutInit(t *testing.T) {
	// given: directory without .pass/
	dir := t.TempDir()
	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"run", "--plan-id", "test-id"})

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

func TestRunCmd_PlanIdFlag(t *testing.T) {
	rootCmd := cmd.NewRootCommand()
	var runCmd *cobra.Command
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "run" {
			runCmd = sub
			break
		}
	}
	if runCmd == nil {
		t.Fatal("run subcommand not found")
	}

	f := runCmd.Flags().Lookup("plan-id")
	if f == nil {
		t.Fatal("--plan-id flag not found")
	}
}

func TestRunCmd_NonExistentPlan(t *testing.T) {
	dir := t.TempDir()
	passDir := filepath.Join(dir, ".pass")
	for _, sub := range []string{".run", "events"} {
		os.MkdirAll(filepath.Join(passDir, sub), 0o755)
	}
	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"run", "--plan-id", "does-not-exist"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent plan")
	}
}
