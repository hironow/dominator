package session_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/session"
	"github.com/hironow/dominator/internal/usecase/port"
)

// --- BuildApprover tests ---

func TestBuildApprover_AutoApprove(t *testing.T) {
	// given
	cfg := domain.ApprovalConfig{Required: false} // IsAutoApprove() = true

	// when
	approver := session.BuildApprover(cfg, nil, nil)

	// then
	if _, ok := approver.(*port.AutoApprover); !ok {
		t.Errorf("expected AutoApprover, got %T", approver)
	}
}

func TestBuildApprover_CmdApprover(t *testing.T) {
	// given
	cfg := domain.ApprovalConfig{Required: true, ApproveCmd: "echo approve"}

	// when
	approver := session.BuildApprover(cfg, nil, nil)

	// then
	if approver == nil {
		t.Fatal("expected non-nil approver")
	}
	if _, ok := approver.(*port.AutoApprover); ok {
		t.Error("expected CmdApprover, got AutoApprover")
	}
}

func TestBuildApprover_StdinApprover(t *testing.T) {
	// given
	cfg := domain.ApprovalConfig{Required: true}
	input := strings.NewReader("")
	out := new(bytes.Buffer)

	// when
	approver := session.BuildApprover(cfg, input, out)

	// then
	if approver == nil {
		t.Fatal("expected non-nil approver")
	}
	if _, ok := approver.(*port.AutoApprover); ok {
		t.Error("expected StdinApprover, got AutoApprover")
	}
}

// --- StdinApprover tests ---

func TestStdinApprover_Yes(t *testing.T) {
	// given
	input := strings.NewReader("y\n")
	out := new(bytes.Buffer)
	a := session.NewStdinApprover(input, out)

	// when
	approved, err := a.RequestApproval(context.Background(), "test?")

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !approved {
		t.Error("expected approval")
	}
}

func TestStdinApprover_YesFull(t *testing.T) {
	// given
	input := strings.NewReader("yes\n")
	out := new(bytes.Buffer)
	a := session.NewStdinApprover(input, out)

	// when
	approved, err := a.RequestApproval(context.Background(), "test?")

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !approved {
		t.Error("expected approval")
	}
}

func TestStdinApprover_No(t *testing.T) {
	// given
	input := strings.NewReader("n\n")
	out := new(bytes.Buffer)
	a := session.NewStdinApprover(input, out)

	// when
	approved, err := a.RequestApproval(context.Background(), "test?")

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if approved {
		t.Error("expected denial")
	}
}

func TestStdinApprover_EmptyDefault(t *testing.T) {
	// given: empty enter = deny (safe default)
	input := strings.NewReader("\n")
	out := new(bytes.Buffer)
	a := session.NewStdinApprover(input, out)

	// when
	approved, err := a.RequestApproval(context.Background(), "test?")

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if approved {
		t.Error("expected denial on empty input")
	}
}

func TestStdinApprover_NilInput(t *testing.T) {
	// given: nil input (library/non-interactive usage)
	a := session.NewStdinApprover(nil, new(bytes.Buffer))

	// when
	approved, err := a.RequestApproval(context.Background(), "test?")

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if approved {
		t.Error("expected denial for nil input")
	}
}

func TestStdinApprover_EOFTerminatedYes(t *testing.T) {
	// given: piped input "y" without trailing newline
	input := strings.NewReader("y")
	a := session.NewStdinApprover(input, new(bytes.Buffer))

	// when
	approved, err := a.RequestApproval(context.Background(), "test?")

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !approved {
		t.Error("expected approval on EOF-terminated 'y'")
	}
}

func TestStdinApprover_ContextCancel(t *testing.T) {
	// given
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	input := strings.NewReader("")
	a := session.NewStdinApprover(input, new(bytes.Buffer))

	// when
	approved, err := a.RequestApproval(ctx, "test?")

	// then
	if err == nil {
		t.Fatal("expected context error")
	}
	if approved {
		t.Error("expected denial on cancelled context")
	}
}

// --- ApprovalConfig (ApproverConfig interface) tests ---

func TestApprovalConfig_IsAutoApprove(t *testing.T) {
	// given
	required := domain.ApprovalConfig{Required: true}
	notRequired := domain.ApprovalConfig{Required: false}

	// then
	if required.IsAutoApprove() {
		t.Error("Required=true should not be auto-approve")
	}
	if !notRequired.IsAutoApprove() {
		t.Error("Required=false should be auto-approve")
	}
}

func TestApprovalConfig_ApproveCmdString(t *testing.T) {
	// given
	cfg := domain.ApprovalConfig{Required: true, ApproveCmd: "echo approve"}

	// then
	if cfg.ApproveCmdString() != "echo approve" {
		t.Errorf("ApproveCmdString = %q, want %q", cfg.ApproveCmdString(), "echo approve")
	}
}
