package cmd_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hironow/dominator/internal/cmd"
)

// approve is a retirement stub since the jun15 MCP pivot follow-up
// (refs issue 0034); see check_test.go.

func TestApproveCmd_SubcommandExists(t *testing.T) {
	// given
	root := cmd.NewRootCommand()
	root.SetArgs([]string{"approve", "--help"})
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(&bytes.Buffer{})

	// when
	if err := root.Execute(); err != nil {
		t.Fatalf("approve --help: %v", err)
	}

	// then
	if !strings.Contains(out.String(), "Retired") {
		t.Errorf("help should state the command is retired, got: %s", out.String())
	}
}

func TestApproveCmd_ReturnsRetirementError(t *testing.T) {
	// given: even with the legacy flag, the stub answers with the redirect
	root := cmd.NewRootCommand()
	root.SetArgs([]string{"approve", "--plan-id", "p-1", t.TempDir()})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})

	// when
	err := root.Execute()

	// then
	if err == nil {
		t.Fatal("expected retirement error")
	}
	if !strings.Contains(err.Error(), "retired") {
		t.Errorf("error should state retirement, got: %v", err)
	}
}
