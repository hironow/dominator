package cmd_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hironow/dominator/internal/cmd"
)

// check is a retirement stub since the jun15 MCP pivot follow-up (refs
// issue 0034): plans had no consumer once `run` was retired.

func TestCheckCmd_SubcommandExists(t *testing.T) {
	// given
	root := cmd.NewRootCommand()
	root.SetArgs([]string{"check", "--help"})
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(&bytes.Buffer{})

	// when
	if err := root.Execute(); err != nil {
		t.Fatalf("check --help: %v", err)
	}

	// then
	if !strings.Contains(out.String(), "Retired") {
		t.Errorf("help should state the command is retired, got: %s", out.String())
	}
}

func TestCheckCmd_ReturnsRetirementError(t *testing.T) {
	// given
	root := cmd.NewRootCommand()
	root.SetArgs([]string{"check", t.TempDir()})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})

	// when
	err := root.Execute()

	// then
	if err == nil {
		t.Fatal("expected retirement error")
	}
	if !strings.Contains(err.Error(), "retired") || !strings.Contains(err.Error(), "/nfr-judge") {
		t.Errorf("error should redirect to the /nfr-judge flow, got: %v", err)
	}
}
