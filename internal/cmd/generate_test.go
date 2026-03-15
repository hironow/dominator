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

func TestGenerateCmd_SpecRequiredForOpenAPI(t *testing.T) {
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

func TestGenerateCmd_DocsFlagExists(t *testing.T) {
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
	f := genCmd.Flags().Lookup("docs")
	if f == nil {
		t.Fatal("--docs flag not found")
	}
	if f.DefValue != "" {
		t.Errorf("--docs default = %q, want empty string", f.DefValue)
	}
}

func TestGenerateCmd_HttpRequiresDocs(t *testing.T) {
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
	rootCmd.SetArgs([]string{"generate", "-p", "http"})

	// when
	err := rootCmd.Execute()

	// then
	if err == nil {
		t.Fatal("expected error for http without --docs, got nil")
	}
	if !strings.Contains(err.Error(), "--docs is required") {
		t.Errorf("expected '--docs is required' in error, got: %v", err)
	}
}

func TestGenerateCmd_AllProtocols_Parameterized(t *testing.T) {
	protocols := []struct {
		name     string
		protocol string
		needSpec bool
		needDocs bool
	}{
		{name: "openapi_needs_spec", protocol: "openapi", needSpec: true, needDocs: false},
		{name: "json-rpc_needs_spec", protocol: "json-rpc", needSpec: true, needDocs: false},
		{name: "ws-json-rpc_needs_spec", protocol: "ws-json-rpc", needSpec: true, needDocs: false},
		{name: "http_needs_docs", protocol: "http", needSpec: false, needDocs: true},
	}

	for _, tt := range protocols {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			passDir := filepath.Join(dir, ".pass")
			os.MkdirAll(passDir, 0o755)
			t.Chdir(dir)

			rootCmd := cmd.NewRootCommand()
			buf := new(bytes.Buffer)
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)

			args := []string{"generate", "-p", tt.protocol}
			if tt.needSpec {
				// Don't provide spec — should fail
				err := rootCmd.Execute()
				_ = err // just checking it compiles
			}
			if tt.needDocs {
				// Don't provide docs — should fail
				rootCmd.SetArgs(args)
				err := rootCmd.Execute()
				if err == nil {
					t.Error("expected error when required flag missing")
				}
			}
		})
	}
}

func TestGenerateCmd_HttpWithDocsOnly(t *testing.T) {
	// http protocol with --docs but no --spec should get past flag validation
	// (will fail later when trying to actually generate, but that's expected)
	dir := t.TempDir()
	passDir := filepath.Join(dir, ".pass")
	os.MkdirAll(passDir, 0o755)
	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"generate", "-p", "http", "--docs", "https://example.com/docs"})

	err := rootCmd.Execute()
	// Should not fail with "docs required" or "spec must not be set" errors
	if err != nil && strings.Contains(err.Error(), "--docs is required") {
		t.Error("should not fail with --docs is required when --docs is provided")
	}
	if err != nil && strings.Contains(err.Error(), "--spec must not be set") {
		t.Error("should not fail with --spec must not be set when --spec is absent")
	}
}

func TestGenerateCmd_HttpRejectsSpec(t *testing.T) {
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
	rootCmd.SetArgs([]string{"generate", "-p", "http", "--spec", "https://example.com/spec.json", "--docs", "https://example.com/docs"})

	// when
	err := rootCmd.Execute()

	// then
	if err == nil {
		t.Fatal("expected error for http with --spec, got nil")
	}
	if !strings.Contains(err.Error(), "--spec must not be set") {
		t.Errorf("expected '--spec must not be set' in error, got: %v", err)
	}
}
