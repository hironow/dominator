package cmd_test

import (
	"bytes"
	"testing"

	"github.com/hironow/dominator/internal/cmd"
	"github.com/spf13/cobra"
)

func TestValidateCommand_Exists(t *testing.T) {
	// given
	rootCmd := cmd.NewRootCommand()

	// then
	var found bool
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "validate" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected validate subcommand to exist")
	}
}

func TestValidateCommand_FailsWithoutInit(t *testing.T) {
	// given: empty directory with no .pass/
	dir := t.TempDir()
	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"validate"})

	// when
	err := rootCmd.Execute()

	// then: should fail because .pass/ doesn't exist
	if err == nil {
		t.Fatal("expected error without .pass/, got nil")
	}
}

func TestValidateCommand_AcceptsPathArg(t *testing.T) {
	// given
	rootCmd := cmd.NewRootCommand()
	var validateCmd *cobra.Command
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "validate" {
			validateCmd = sub
			break
		}
	}

	// then
	if validateCmd == nil {
		t.Fatal("validate subcommand not found")
	}
	// Verify the command has Args configured (MaximumNArgs)
	if validateCmd.Args == nil {
		t.Error("expected Args to be set on validate command")
	}
}

func TestValidateCommand_ShortDescription(t *testing.T) {
	// given
	rootCmd := cmd.NewRootCommand()
	var validateCmd *cobra.Command
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "validate" {
			validateCmd = sub
			break
		}
	}

	// then
	if validateCmd == nil {
		t.Fatal("validate subcommand not found")
	}
	if validateCmd.Short == "" {
		t.Error("validate command has no short description")
	}
}
