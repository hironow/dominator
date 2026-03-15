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

func TestInitCommand_AlreadyInitialized(t *testing.T) {
	// given: .pass/ directory already exists
	dir := t.TempDir()
	passDir := filepath.Join(dir, ".pass")
	if err := os.MkdirAll(passDir, 0755); err != nil {
		t.Fatalf("create pass dir: %v", err)
	}

	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"init"})

	// when
	execErr := rootCmd.Execute()

	// then: should fail with "already exists"
	if execErr == nil {
		t.Fatal("expected error for already initialized, got nil")
	}
	if got := execErr.Error(); !strings.Contains(got, "already exists") {
		t.Errorf("expected 'already exists' in error, got: %s", got)
	}
}

func TestInitCommand_AlreadyExists_SuggestsForce(t *testing.T) {
	// given: .pass/ directory already exists
	dir := t.TempDir()
	passDir := filepath.Join(dir, ".pass")
	if err := os.MkdirAll(passDir, 0755); err != nil {
		t.Fatalf("create pass dir: %v", err)
	}

	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"init"})

	// when
	execErr := rootCmd.Execute()

	// then
	if execErr == nil {
		t.Fatal("expected error when .pass already exists")
	}
	if !strings.Contains(execErr.Error(), "--force") {
		t.Errorf("expected '--force' hint in error, got: %v", execErr)
	}
}

func TestInitCommand_Force_OverwritesExisting(t *testing.T) {
	// given: .pass/ directory already exists
	dir := t.TempDir()
	passDir := filepath.Join(dir, ".pass")
	if err := os.MkdirAll(passDir, 0755); err != nil {
		t.Fatalf("create pass dir: %v", err)
	}

	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"init", "--force"})

	// when
	execErr := rootCmd.Execute()

	// then
	if execErr != nil {
		t.Fatalf("init --force failed: %v", execErr)
	}
}

func TestInitCmd_OtelBackend_CreatesOtelEnv(t *testing.T) {
	// given
	dir := t.TempDir()
	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"init", "--otel-backend", "jaeger"})

	// when
	execErr := rootCmd.Execute()

	// then
	if execErr != nil {
		t.Fatalf("init --otel-backend jaeger failed: %v", execErr)
	}
	otelPath := filepath.Join(dir, ".pass", ".otel.env")
	data, readErr := os.ReadFile(otelPath)
	if readErr != nil {
		t.Fatalf(".otel.env not created: %v", readErr)
	}
	if !strings.Contains(string(data), "OTEL_EXPORTER_OTLP_ENDPOINT") {
		t.Errorf("expected OTEL_EXPORTER_OTLP_ENDPOINT in .otel.env, got:\n%s", data)
	}
}

func TestInitCommand_OtelFlags_Exist(t *testing.T) {
	// given
	rootCmd := cmd.NewRootCommand()

	// when — find init subcommand
	var initCmd *cobra.Command
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "init" {
			initCmd = sub
			break
		}
	}
	if initCmd == nil {
		t.Fatal("init subcommand not found")
	}

	// then — otel flags exist
	for _, flag := range []string{"otel-backend", "otel-entity", "otel-project"} {
		if initCmd.Flags().Lookup(flag) == nil {
			t.Errorf("init flag --%s not found", flag)
		}
	}
}

func TestInitCommand_CreatesK6ScriptsDir(t *testing.T) {
	// given
	dir := t.TempDir()
	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"init"})

	// when
	execErr := rootCmd.Execute()

	// then
	if execErr != nil {
		t.Fatalf("init failed: %v", execErr)
	}
	k6Dir := filepath.Join(dir, ".pass", "k6-scripts")
	info, err := os.Stat(k6Dir)
	if err != nil {
		t.Fatalf("k6-scripts dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("k6-scripts should be a directory")
	}
}

func TestInitCommand_CreatesInsightsDir(t *testing.T) {
	// given
	dir := t.TempDir()
	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"init"})

	// when
	execErr := rootCmd.Execute()

	// then
	if execErr != nil {
		t.Fatalf("init failed: %v", execErr)
	}
	insightsDir := filepath.Join(dir, ".pass", "insights")
	info, err := os.Stat(insightsDir)
	if err != nil {
		t.Fatalf("insights dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("insights should be a directory")
	}
}
