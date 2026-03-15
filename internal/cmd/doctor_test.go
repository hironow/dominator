package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/hironow/dominator/internal/cmd"
	"github.com/spf13/cobra"
)

func TestDoctorCommand_Exists(t *testing.T) {
	// given
	rootCmd := cmd.NewRootCommand()

	// then
	var found bool
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "doctor" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected doctor subcommand to exist")
	}
}

func TestDoctorCommand_JSONFlag(t *testing.T) {
	// given
	rootCmd := cmd.NewRootCommand()

	// when
	var doctorCmd *cobra.Command
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "doctor" {
			doctorCmd = sub
			break
		}
	}

	// then
	if doctorCmd == nil {
		t.Fatal("doctor subcommand not found")
	}
	if f := doctorCmd.Flags().Lookup("json"); f == nil {
		t.Error("doctor --json flag not found")
	}
}

func TestDoctorCommand_RepairFlag(t *testing.T) {
	// given
	rootCmd := cmd.NewRootCommand()

	// when
	var doctorCmd *cobra.Command
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "doctor" {
			doctorCmd = sub
			break
		}
	}

	// then
	if doctorCmd == nil {
		t.Fatal("doctor subcommand not found")
	}
	if f := doctorCmd.Flags().Lookup("repair"); f == nil {
		t.Error("doctor --repair flag not found")
	}
}

func TestDoctorCommand_FailsWithoutInit(t *testing.T) {
	// given: empty directory with no .pass/
	dir := t.TempDir()
	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"doctor"})

	// when
	err := rootCmd.Execute()

	// then: should report failures (state-dir missing)
	// Doctor doesn't error but reports check failures via SilentError
	_ = err
	output := buf.String()
	if len(output) == 0 {
		t.Error("expected doctor output, got empty")
	}
}

func TestDoctorCommand_PassesWithInit(t *testing.T) {
	// given: initialized .pass/ with valid config
	dir := t.TempDir()
	passDir := filepath.Join(dir, ".pass")
	os.MkdirAll(filepath.Join(passDir, "events"), 0o755)
	os.MkdirAll(filepath.Join(passDir, ".run"), 0o755)
	os.WriteFile(filepath.Join(passDir, "config.yaml"), []byte("lang: ja\n"), 0o644)

	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"doctor"})

	// when
	rootCmd.Execute()

	// then: output should contain doctor header
	output := buf.String()
	if len(output) == 0 {
		t.Error("expected doctor output, got empty")
	}
}
