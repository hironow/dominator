//go:build integration

package integration_test

import (
	"bytes"
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
)

// TestLifecycle_InitGenerateCheckApproveRun tests the full CLI lifecycle
// without real Claude Code or k6 (those require E2E).
func TestLifecycle_InitGenerateCheckApproveRun(t *testing.T) {
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
	requiredDirs := []string{
		".run",
		"events",
		"outbox",
		"inbox",
		"archive",
		"k6-scripts",
		"insights",
	}
	for _, d := range requiredDirs {
		path := filepath.Join(passDir, d)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("expected directory %s to exist: %v", d, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("expected %s to be a directory", d)
		}
	}

	// Verify config.yaml was created
	configPath := filepath.Join(passDir, "config.yaml")
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("expected config.yaml to exist: %v", err)
	}

	// 3. Generate would need Claude — skip in integration, test the wiring
	// Instead, create a fake k6 script in .pass/k6-scripts/
	fakeScript := `import http from 'k6/http';
export default function() { http.get('https://example.com'); }
`
	scriptPath := filepath.Join(passDir, "k6-scripts", "load-test.js")
	if err := os.WriteFile(scriptPath, []byte(fakeScript), 0o644); err != nil {
		t.Fatalf("write fake script: %v", err)
	}

	// 4. Check -> verify plan JSON output
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	checkCmd := cmd.NewRootCommand()
	checkCmd.SetArgs([]string{"check", dir})
	checkCmd.SetOut(stdout)
	checkCmd.SetErr(stderr)
	if err := checkCmd.Execute(); err != nil {
		t.Fatalf("check failed: %v", err)
	}

	// Parse plan from stdout
	var plan domain.Plan
	if err := json.Unmarshal(stdout.Bytes(), &plan); err != nil {
		t.Fatalf("unmarshal plan from stdout: %v (raw: %s)", err, stdout.String())
	}
	if plan.ID == "" {
		t.Error("expected plan ID to be non-empty")
	}
	if plan.Script != "load-test.js" {
		t.Errorf("expected plan script to be 'load-test.js', got %q", plan.Script)
	}
	if plan.Approved {
		t.Error("plan should not be approved before explicit approval")
	}

	// 5. Approve -> verify plan approved
	stdout.Reset()
	stderr.Reset()
	approveCmd := cmd.NewRootCommand()
	approveCmd.SetArgs([]string{"approve", "--plan-id", string(plan.ID), dir})
	approveCmd.SetOut(stdout)
	approveCmd.SetErr(stderr)
	if err := approveCmd.Execute(); err != nil {
		t.Fatalf("approve failed: %v", err)
	}

	var approvedPlan domain.Plan
	if err := json.Unmarshal(stdout.Bytes(), &approvedPlan); err != nil {
		t.Fatalf("unmarshal approved plan: %v", err)
	}
	if !approvedPlan.Approved {
		t.Error("expected plan to be approved after approval")
	}
	if approvedPlan.ApprovedAt.IsZero() {
		t.Error("expected approved_at to be set")
	}

	// 6. Verify event store has events recorded
	eventStore := eventsource.NewFileEventStore(
		filepath.Join(passDir, "events"),
		&domain.NopLogger{},
	)
	events, _, err := eventStore.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll events: %v", err)
	}

	hasCreated := false
	hasApproved := false
	for _, ev := range events {
		if ev.Type == domain.EventPlanCreated {
			hasCreated = true
		}
		if ev.Type == domain.EventPlanApproved {
			hasApproved = true
		}
	}
	if !hasCreated {
		t.Error("expected plan.created event in event store")
	}
	if !hasApproved {
		t.Error("expected plan.approved event in event store")
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

// TestLifecycle_CheckWithoutInit verifies check fails gracefully when .pass/ doesn't exist.
func TestLifecycle_CheckWithoutInit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// given: temp directory without init
	dir := t.TempDir()

	// when: run check
	root := cmd.NewRootCommand()
	root.SetArgs([]string{"check", dir})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	err := root.Execute()

	// then: error
	if err == nil {
		t.Fatal("expected check to fail without init")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

// TestLifecycle_EventTimestampOrdering verifies that events from
// init -> check -> approve maintain chronological order.
func TestLifecycle_EventTimestampOrdering(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// given
	dir := t.TempDir()

	// init
	initCmd := cmd.NewRootCommand()
	initCmd.SetArgs([]string{"init", dir})
	initCmd.SetOut(&bytes.Buffer{})
	initCmd.SetErr(&bytes.Buffer{})
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("init: %v", err)
	}

	// write script
	passDir := filepath.Join(dir, ".pass")
	scriptPath := filepath.Join(passDir, "k6-scripts", "test.js")
	if err := os.WriteFile(scriptPath, []byte("// k6 script"), 0o644); err != nil {
		t.Fatalf("write script: %v", err)
	}

	time.Sleep(10 * time.Millisecond) // ensure timestamp separation

	// check
	stdout := &bytes.Buffer{}
	checkCmd := cmd.NewRootCommand()
	checkCmd.SetArgs([]string{"check", dir})
	checkCmd.SetOut(stdout)
	checkCmd.SetErr(&bytes.Buffer{})
	if err := checkCmd.Execute(); err != nil {
		t.Fatalf("check: %v", err)
	}

	var plan domain.Plan
	if err := json.Unmarshal(stdout.Bytes(), &plan); err != nil {
		t.Fatalf("unmarshal plan: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	// approve
	approveRoot := cmd.NewRootCommand()
	approveRoot.SetArgs([]string{"approve", "--plan-id", string(plan.ID), dir})
	approveRoot.SetOut(&bytes.Buffer{})
	approveRoot.SetErr(&bytes.Buffer{})
	if err := approveRoot.Execute(); err != nil {
		t.Fatalf("approve: %v", err)
	}

	// then: verify chronological order
	eventStore := eventsource.NewFileEventStore(
		filepath.Join(passDir, "events"),
		&domain.NopLogger{},
	)
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

// TestLifecycle_RunUnapprovedPlan verifies that run rejects an unapproved plan.
func TestLifecycle_RunUnapprovedPlan(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// given: initialized directory with a plan (not approved)
	dir := t.TempDir()
	passDir := filepath.Join(dir, ".pass")
	if err := session.InitPassDir(passDir, &domain.NopLogger{}); err != nil {
		t.Fatalf("InitPassDir: %v", err)
	}

	// write script and create plan
	scriptPath := filepath.Join(passDir, "k6-scripts", "test.js")
	if err := os.WriteFile(scriptPath, []byte("// k6"), 0o644); err != nil {
		t.Fatalf("write script: %v", err)
	}

	stdout := &bytes.Buffer{}
	checkRoot := cmd.NewRootCommand()
	checkRoot.SetArgs([]string{"check", dir})
	checkRoot.SetOut(stdout)
	checkRoot.SetErr(&bytes.Buffer{})
	if err := checkRoot.Execute(); err != nil {
		t.Fatalf("check: %v", err)
	}

	var plan domain.Plan
	if err := json.Unmarshal(stdout.Bytes(), &plan); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// when: try to run without approving
	runCmd := cmd.NewRootCommand()
	runCmd.SetArgs([]string{"run", "--plan-id", string(plan.ID), dir})
	runCmd.SetOut(&bytes.Buffer{})
	runCmd.SetErr(&bytes.Buffer{})
	err := runCmd.Execute()

	// then: error about unapproved plan
	if err == nil {
		t.Fatal("expected run to fail for unapproved plan")
	}
	if !strings.Contains(err.Error(), "not approved") {
		t.Errorf("expected 'not approved' error, got: %v", err)
	}
}
