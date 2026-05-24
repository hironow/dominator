//go:build e2e

package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// createPlan writes a plan JSON file to .pass/.run/plans/{id}.json
func createPlan(t *testing.T, dir, planID, script string) {
	t.Helper()
	plansDir := filepath.Join(dir, ".pass", ".run", "plans")
	if err := os.MkdirAll(plansDir, 0o755); err != nil {
		t.Fatal(err)
	}

	plan := map[string]any{
		"plan_id": planID,
		"script":  script,
		"target": map[string]any{
			"url":      "http://localhost:3000",
			"protocol": "http",
		},
		"load": map[string]any{
			"vus":      10,
			"duration": "30s",
			"ramp_up":  "5s",
		},
		"nfr": map[string]any{
			"performance": map[string]any{
				"p95_latency_ms":     500,
				"error_rate_percent": 1.0,
			},
			"reliability": map[string]any{
				"success_rate_percent": 99.0,
			},
			"scalability": map[string]any{
				"target_rps": 100,
			},
		},
		"approved":   false,
		"created_at": time.Now().UTC().Format(time.RFC3339),
	}
	data, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(plansDir, planID+".json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestE2E_Pipeline_ApprovePlan(t *testing.T) {
	dir := initTestRepo(t)
	writeConfig(t, dir, defaultTestConfig())

	planID := "test-plan-001"
	createPlan(t, dir, planID, "load-test.js")

	// Approve the plan
	stdout, stderr, err := runCmd(t, dir, "approve", "--plan-id", planID)
	if err != nil {
		t.Fatalf("approve: %v\nstderr: %s", err, stderr)
	}

	// Verify stdout contains approved plan JSON
	var plan map[string]any
	parseJSONOutput(t, stdout, &plan)
	if plan["approved"] != true {
		t.Errorf("expected approved=true, got %v", plan["approved"])
	}
	if plan["plan_id"] != planID {
		t.Errorf("expected plan_id=%s, got %v", planID, plan["plan_id"])
	}
}

func TestE2E_Pipeline_ApproveAlreadyApproved(t *testing.T) {
	dir := initTestRepo(t)
	writeConfig(t, dir, defaultTestConfig())

	planID := "test-plan-002"
	createPlan(t, dir, planID, "load-test.js")

	// Approve once
	_, _, err := runCmd(t, dir, "approve", "--plan-id", planID)
	if err != nil {
		t.Fatalf("first approve: %v", err)
	}

	// Approve again — should succeed (idempotent)
	stdout, _, err := runCmd(t, dir, "approve", "--plan-id", planID)
	if err != nil {
		t.Fatalf("second approve: %v", err)
	}
	var plan map[string]any
	parseJSONOutput(t, stdout, &plan)
	if plan["approved"] != true {
		t.Error("expected approved=true on re-approve")
	}
}

func TestE2E_Pipeline_StatusAfterApprove(t *testing.T) {
	dir := initTestRepo(t)
	writeConfig(t, dir, defaultTestConfig())

	planID := "test-plan-005"
	createPlan(t, dir, planID, "load-test.js")

	// Approve
	_, _, err := runCmd(t, dir, "approve", "--plan-id", planID)
	if err != nil {
		t.Fatalf("approve: %v", err)
	}

	// Status should reflect the approved plan
	stdout, _, err := runCmd(t, dir, "status", "--output", "json")
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	var result map[string]any
	parseJSONOutput(t, stdout, &result)
	approvedCount, _ := result["approved_plan_count"].(float64)
	if approvedCount < 1 {
		t.Errorf("expected at least 1 approved plan, got %v", approvedCount)
	}
}

func TestE2E_Pipeline_EventsRecorded(t *testing.T) {
	dir := initTestRepo(t)
	writeConfig(t, dir, defaultTestConfig())

	planID := "test-plan-006"
	createPlan(t, dir, planID, "load-test.js")

	// Approve — should record event
	_, _, err := runCmd(t, dir, "approve", "--plan-id", planID)
	if err != nil {
		t.Fatalf("approve: %v", err)
	}

	// Check events directory has files
	eventsDir := filepath.Join(dir, ".pass", "events")
	entries, err := os.ReadDir(eventsDir)
	if err != nil {
		t.Fatalf("read events: %v", err)
	}
	if len(entries) == 0 {
		t.Error("expected event files after approve, got empty events dir")
	}
}
