package cmd_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/hironow/dominator/internal/cmd"
)

func TestStatusCmd_SubcommandExists(t *testing.T) {
	// given
	root := cmd.NewRootCommand()

	// then
	found := false
	for _, sub := range root.Commands() {
		if sub.Name() == "status" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("status subcommand not found")
	}
}

func TestStatusCmd_JSONOutput(t *testing.T) {
	// given: initialize .pass/ via init command
	dir := t.TempDir()
	t.Chdir(dir)

	initRoot := cmd.NewRootCommand()
	initBuf := new(bytes.Buffer)
	initRoot.SetOut(initBuf)
	initRoot.SetErr(initBuf)
	initRoot.SetArgs([]string{"init"})
	if err := initRoot.Execute(); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// when: run status --output json
	root := cmd.NewRootCommand()
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"status", "-o", "json", dir})

	execErr := root.Execute()

	// then
	if execErr != nil {
		t.Fatalf("status -o json failed: %v", execErr)
	}
	if !json.Valid(stdout.Bytes()) {
		t.Errorf("stdout is not valid JSON: %s", stdout.String())
	}
}

func TestStatusCmd_FailsWithoutInit(t *testing.T) {
	// given: a temp dir without .pass/
	dir := t.TempDir()
	root := cmd.NewRootCommand()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"status", dir})

	// when
	err := root.Execute()

	// then
	if err == nil {
		t.Fatal("expected error for uninitialized state")
	}
	if !strings.Contains(err.Error(), "init") {
		t.Errorf("expected error to mention 'init', got: %s", err.Error())
	}
}

func TestStatusCmd_OutputFlag(t *testing.T) {
	rootCmd := cmd.NewRootCommand()
	f := rootCmd.PersistentFlags().Lookup("output")
	if f == nil {
		t.Fatal("--output flag not found")
	}
	if f.DefValue != "text" {
		t.Errorf("--output default = %q, want %q", f.DefValue, "text")
	}
}

func TestStatusCmd_TextOutput(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	initRoot := cmd.NewRootCommand()
	initBuf := new(bytes.Buffer)
	initRoot.SetOut(initBuf)
	initRoot.SetErr(initBuf)
	initRoot.SetArgs([]string{"init"})
	initRoot.Execute()

	root := cmd.NewRootCommand()
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"status", dir})

	err := root.Execute()
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	// Text mode should produce some output on stderr
	// (stdout is reserved for machine-readable data)
}
