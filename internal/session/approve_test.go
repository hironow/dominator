package session_test

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"

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

// --- AutoApprover tests ---

func TestAutoApprover_AlwaysApproves(t *testing.T) {
	// given
	a := &port.AutoApprover{}

	// when
	approved, err := a.RequestApproval(context.Background(), "msg")

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !approved {
		t.Error("expected AutoApprover to always approve")
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

func TestStdinApprover_EOFTerminatedNo(t *testing.T) {
	// given: piped "n" without trailing newline — should deny (not error)
	input := strings.NewReader("n")
	a := session.NewStdinApprover(input, new(bytes.Buffer))

	// when
	approved, err := a.RequestApproval(context.Background(), "test?")

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if approved {
		t.Error("expected denial for EOF-terminated 'n' input")
	}
}

func TestStdinApprover_SharedReader(t *testing.T) {
	// given: a shared reader with approval line + subsequent data
	in := strings.NewReader("y\nnext-line\n")
	a := session.NewStdinApprover(in, new(bytes.Buffer))

	// when
	approved, err := a.RequestApproval(context.Background(), "test?")

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !approved {
		t.Fatal("expected approval")
	}

	// then: remaining data is still available from the shared reader
	remaining := make([]byte, 64)
	n, _ := in.Read(remaining)
	got := string(remaining[:n])
	if got != "next-line\n" {
		t.Errorf("shared reader lost data: got %q, want %q", got, "next-line\n")
	}
}

func TestStdinApprover_ShowsMessage(t *testing.T) {
	// given
	input := strings.NewReader("y\n")
	out := new(bytes.Buffer)
	a := session.NewStdinApprover(input, out)

	// when
	a.RequestApproval(context.Background(), "Continue check?")

	// then
	if !strings.Contains(out.String(), "Continue? [y/N]") {
		t.Errorf("prompt not shown, got: %q", out.String())
	}
}

func TestStdinApprover_ContextCancel(t *testing.T) {
	// given
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	input := new(blockingReader)
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

func TestStdinApprover_Timeout(t *testing.T) {
	// given
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	input := new(blockingReader)
	a := session.NewStdinApprover(input, new(bytes.Buffer))

	// when
	approved, err := a.RequestApproval(ctx, "test?")

	// then
	if approved {
		t.Error("expected denial on timeout")
	}
	if err == nil {
		t.Error("expected error on timeout")
	}
}

// --- CmdApprover tests ---

func TestCmdApprover_EmptyTemplate(t *testing.T) {
	// given
	a := session.NewCmdApprover("")

	// when
	approved, err := a.RequestApproval(context.Background(), "msg")

	// then
	if err == nil {
		t.Error("expected error for empty template")
	}
	if approved {
		t.Error("expected denial for empty template")
	}
}

func TestCmdApprover_FactoryDI(t *testing.T) {
	// given: inject a factory that records the expanded command
	var capturedArgs []string
	a := session.NewCmdApproverForTest("echo {message}",
		func(ctx context.Context, name string, args ...string) *exec.Cmd {
			capturedArgs = args
			return exec.Command("true")
		},
	)

	// when
	approved, err := a.RequestApproval(context.Background(), "hello world")

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !approved {
		t.Error("expected approval for exit code 0")
	}
	if len(capturedArgs) == 0 {
		t.Fatal("expected args to be captured by factory")
	}
	joined := strings.Join(capturedArgs, " ")
	if !strings.Contains(joined, "'hello world'") {
		t.Errorf("expected quoted message in command, got: %s", joined)
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

// blockingReader never returns data, simulating a blocking stdin.
type blockingReader struct{}

func (r *blockingReader) Read(p []byte) (int, error) {
	select {} //nolint:staticcheck // intentional blocking for test
}
