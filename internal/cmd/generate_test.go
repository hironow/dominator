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

func TestGenerateCmd_SubcommandExists(t *testing.T) {
	// given
	rootCmd := cmd.NewRootCommand()

	// when
	var found bool
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "generate" {
			found = true
			break
		}
	}

	// then
	if !found {
		t.Fatal("generate subcommand not found")
	}
}

func TestGenerateCmd_SpecFlagRequired(t *testing.T) {
	// given
	rootCmd := cmd.NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"generate"})

	// when
	err := rootCmd.Execute()

	// then
	if err == nil {
		t.Fatal("expected error when --spec not provided, got nil")
	}
	if !strings.Contains(err.Error(), "spec") {
		t.Errorf("expected error to mention 'spec', got: %v", err)
	}
}

func TestGenerateCmd_ProtocolFlagDefaults(t *testing.T) {
	// given
	rootCmd := cmd.NewRootCommand()

	// when
	var genCmd *cobra.Command
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "generate" {
			genCmd = sub
			break
		}
	}
	if genCmd == nil {
		t.Fatal("generate subcommand not found")
	}

	// then
	f := genCmd.Flags().Lookup("protocol")
	if f == nil {
		t.Fatal("--protocol flag not found")
	}
	if f.DefValue != "openapi" {
		t.Errorf("--protocol default = %q, want %q", f.DefValue, "openapi")
	}
	if f.Shorthand != "p" {
		t.Errorf("--protocol shorthand = %q, want %q", f.Shorthand, "p")
	}
}

func TestGenerateCmd_ForceFlagExists(t *testing.T) {
	// given
	rootCmd := cmd.NewRootCommand()

	// when
	var genCmd *cobra.Command
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "generate" {
			genCmd = sub
			break
		}
	}
	if genCmd == nil {
		t.Fatal("generate subcommand not found")
	}

	// then
	f := genCmd.Flags().Lookup("force")
	if f == nil {
		t.Fatal("--force flag not found")
	}
	if f.DefValue != "false" {
		t.Errorf("--force default = %q, want %q", f.DefValue, "false")
	}
}

func TestGenerateCmd_FailsWithoutInit(t *testing.T) {
	// given: directory without .pass/
	dir := t.TempDir()
	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"generate", "--spec", "https://example.com/spec.json"})

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

func TestGenerateCmd_FailsWithInvalidProtocol(t *testing.T) {
	// given: directory with .pass/
	dir := t.TempDir()
	passDir := filepath.Join(dir, ".pass")
	if err := os.MkdirAll(passDir, 0o755); err != nil {
		t.Fatalf("create pass dir: %v", err)
	}
	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"generate", "--spec", "https://example.com/spec.json", "-p", "invalid-protocol"})

	// when
	err := rootCmd.Execute()

	// then
	if err == nil {
		t.Fatal("expected error for invalid protocol, got nil")
	}
	if !strings.Contains(err.Error(), "invalid protocol") {
		t.Errorf("expected 'invalid protocol' in error, got: %v", err)
	}
}
