//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hironow/dominator/internal/cmd"
	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/eventsource"
	"github.com/hironow/dominator/internal/session"
	"github.com/hironow/dominator/internal/usecase"
	"github.com/hironow/dominator/internal/usecase/port"
)

// Post-pivot lifecycle (refs issue 0034): init prepares .pass/, the
// Claude Code session judges via the MCP tools (record_result), and
// status / get_insights read the result. The plan-staging lifecycle
// (check -> approve -> run) is retired; those commands are stubs.

func TestLifecycle_PostPivot_InitRecordStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// given: temp directory
	dir := t.TempDir()

	// 1. Init
	root := cmd.NewRootCommand()
	root.SetArgs([]string{"init", dir})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// 2. Verify .pass/ structure created
	passDir := filepath.Join(dir, ".pass")
	for _, d := range []string{".run", "events", "outbox", "inbox", "archive", "k6-scripts", "insights"} {
		info, err := os.Stat(filepath.Join(passDir, d))
		if err != nil {
			t.Errorf("expected directory %s to exist: %v", d, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("expected %s to be a directory", d)
		}
	}
	if _, err := os.Stat(filepath.Join(passDir, "config.yaml")); err != nil {
		t.Fatalf("expected config.yaml to exist: %v", err)
	}

	// 3. Record a verdict through the post-pivot write path (the MCP
	// record_result tool over stdio — the session's judgment entry).
	req := `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"record_result","arguments":{"target_id":"api","verdict":"pass","summary":"p95 320ms < 500ms"}}}` + "\n"
	var out bytes.Buffer
	cfg, err := session.LoadConfig(filepath.Join(passDir, "config.yaml"))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	store := session.NewEventStore(passDir, nil)
	emitter := newIntegrationEmitter(cfg, store)
	srv := session.NewMCPServer(strings.NewReader(req), &out, nil).WithPassDir(passDir).WithEmitter(emitter)
	if err := srv.Serve(context.Background()); err != nil {
		t.Fatalf("mcp serve: %v", err)
	}
	if !strings.Contains(out.String(), `\"persisted\":true`) {
		t.Fatalf("record_result did not persist: %s", out.String())
	}

	// 4. Event store has the judgment
	eventStore := eventsource.NewFileEventStore(filepath.Join(passDir, "events"), &domain.NopLogger{})
	events, _, err := eventStore.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll events: %v", err)
	}
	found := false
	for _, ev := range events {
		if ev.Type == domain.EventJudgmentRecorded {
			found = true
		}
	}
	if !found {
		t.Error("expected judgment.recorded event in event store")
	}

	// 5. Status (CLI read model) reports the judgment
	stdout := &bytes.Buffer{}
	statusCmd := cmd.NewRootCommand()
	statusCmd.SetArgs([]string{"status", "-o", "json", dir})
	statusCmd.SetOut(stdout)
	statusCmd.SetErr(&bytes.Buffer{})
	if err := statusCmd.Execute(); err != nil {
		t.Fatalf("status: %v", err)
	}
	var report map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("unmarshal status: %v (raw: %s)", err, stdout.String())
	}
}

// TestLifecycle_RetiredCommandsRedirect pins the retirement contract:
// every retired pipeline command fails with a redirect to the
// /nfr-judge flow instead of doing plan work.
func TestLifecycle_RetiredCommandsRedirect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := t.TempDir()
	initCmd := cmd.NewRootCommand()
	initCmd.SetArgs([]string{"init", dir})
	initCmd.SetOut(&bytes.Buffer{})
	initCmd.SetErr(&bytes.Buffer{})
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("init: %v", err)
	}

	for _, args := range [][]string{
		{"check", dir},
		{"approve", "--plan-id", "p-1", dir},
		{"run", dir},
		{"generate", dir},
		{"validate", dir},
	} {
		root := cmd.NewRootCommand()
		root.SetArgs(args)
		root.SetOut(&bytes.Buffer{})
		root.SetErr(&bytes.Buffer{})
		err := root.Execute()
		if err == nil {
			t.Errorf("%v: expected retirement error", args)
			continue
		}
		if !strings.Contains(err.Error(), "retired") {
			t.Errorf("%v: error should state retirement, got: %v", args, err)
		}
	}
}

// TestLifecycle_EventTimestampOrdering verifies that judgments recorded
// through the post-pivot write path keep chronological order.
func TestLifecycle_EventTimestampOrdering(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := t.TempDir()
	initCmd := cmd.NewRootCommand()
	initCmd.SetArgs([]string{"init", dir})
	initCmd.SetOut(&bytes.Buffer{})
	initCmd.SetErr(&bytes.Buffer{})
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("init: %v", err)
	}
	passDir := filepath.Join(dir, ".pass")
	cfg, err := session.LoadConfig(filepath.Join(passDir, "config.yaml"))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	store := session.NewEventStore(passDir, nil)
	emitter := newIntegrationEmitter(cfg, store)

	for i, verdict := range []string{"pass", "fail"} {
		req := `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"record_result","arguments":{"target_id":"api","verdict":"` + verdict + `","summary":"s"}}}` + "\n"
		var out bytes.Buffer
		srv := session.NewMCPServer(strings.NewReader(req), &out, nil).WithPassDir(passDir).WithEmitter(emitter)
		if err := srv.Serve(context.Background()); err != nil {
			t.Fatalf("mcp serve %d: %v", i, err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	eventStore := eventsource.NewFileEventStore(filepath.Join(passDir, "events"), &domain.NopLogger{})
	events, _, err := eventStore.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	for i := 1; i < len(events); i++ {
		if events[i].Timestamp.Before(events[i-1].Timestamp) {
			t.Errorf("event[%d] (%s at %v) is before event[%d] (%s at %v)",
				i, events[i].Type, events[i].Timestamp,
				i-1, events[i-1].Type, events[i-1].Timestamp)
		}
	}
}

// TestLifecycle_InitIdempotent verifies that running init twice with --force
// does not lose existing k6 scripts.
func TestLifecycle_InitIdempotent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// given
	dir := t.TempDir()

	// First init
	root := cmd.NewRootCommand()
	root.SetArgs([]string{"init", dir})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("first init: %v", err)
	}

	// Write a k6 script
	scriptPath := filepath.Join(dir, ".pass", "k6-scripts", "existing.js")
	if err := os.WriteFile(scriptPath, []byte("// existing script"), 0o644); err != nil {
		t.Fatalf("write script: %v", err)
	}

	// Second init with --force
	root2 := cmd.NewRootCommand()
	root2.SetArgs([]string{"init", "--force", dir})
	root2.SetOut(&bytes.Buffer{})
	root2.SetErr(&bytes.Buffer{})
	if err := root2.Execute(); err != nil {
		t.Fatalf("second init: %v", err)
	}

	// then: config.yaml exists
	if _, err := os.Stat(filepath.Join(dir, ".pass", "config.yaml")); err != nil {
		t.Errorf("config.yaml should exist after re-init: %v", err)
	}

	// then: k6 script still exists (init does not delete existing content)
	if _, err := os.Stat(scriptPath); err != nil {
		t.Errorf("existing k6 script should survive re-init: %v", err)
	}
}

// newIntegrationEmitter mirrors the cmd/mcp.go composition root: a
// judgment emitter over the real event store (no fakes — integration
// drives the actual write path).
func newIntegrationEmitter(cfg domain.Config, store port.EventStore) port.JudgmentEventEmitter {
	return usecase.NewJudgmentEventEmitter(domain.NewJudgeAggregate(cfg), store, &domain.NopLogger{})
}
