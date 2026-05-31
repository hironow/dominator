//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
)

// createPlan writes a plan JSON file to .pass/.run/plans/{id}.json inside the container.
func createPlan(t *testing.T, ctx context.Context, c testcontainers.Container, dir, planID, script string) {
	t.Helper()
	plansDir := fmt.Sprintf("%s/.pass/.run/plans", dir)
	execInContainer(t, ctx, c, []string{"mkdir", "-p", plansDir})

	plan := fmt.Sprintf(`{
  "plan_id": "%s",
  "script": "%s",
  "target": {
    "url": "http://localhost:3000",
    "protocol": "http"
  },
  "load": {
    "vus": 10,
    "duration": "30s",
    "ramp_up": "5s"
  },
  "nfr": {
    "performance": {
      "p95_latency_ms": 500,
      "error_rate_percent": 1.0
    },
    "reliability": {
      "success_rate_percent": 99.0
    },
    "scalability": {
      "target_rps": 100
    }
  },
  "approved": false,
  "created_at": "%s"
}`, planID, script, time.Now().UTC().Format(time.RFC3339))

	heredocWrite(t, ctx, c, fmt.Sprintf("%s/%s.json", plansDir, planID), plan)
}

func TestE2E_Pipeline_ApprovePlan(t *testing.T) {
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_pipe_approve"

	initTestRepo(t, ctx, c, dir)
	heredocWrite(t, ctx, c, dir+"/.pass/config.yaml", defaultTestConfigYAML())

	planID := "test-plan-001"
	createPlan(t, ctx, c, dir, planID, "load-test.js")

	// Approve the plan
	stdout, _, err := runCmd(t, ctx, c, dir, "approve", "--plan-id", planID)
	if err != nil {
		t.Fatalf("approve failed: %v", err)
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
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_pipe_reapprove"

	initTestRepo(t, ctx, c, dir)
	heredocWrite(t, ctx, c, dir+"/.pass/config.yaml", defaultTestConfigYAML())

	planID := "test-plan-002"
	createPlan(t, ctx, c, dir, planID, "load-test.js")

	// Approve once
	_, _, err := runCmd(t, ctx, c, dir, "approve", "--plan-id", planID)
	if err != nil {
		t.Fatalf("first approve: %v", err)
	}

	// Approve again — should succeed (idempotent)
	stdout, _, err := runCmd(t, ctx, c, dir, "approve", "--plan-id", planID)
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
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_pipe_status"

	initTestRepo(t, ctx, c, dir)
	heredocWrite(t, ctx, c, dir+"/.pass/config.yaml", defaultTestConfigYAML())

	planID := "test-plan-005"
	createPlan(t, ctx, c, dir, planID, "load-test.js")

	// Approve
	_, _, err := runCmd(t, ctx, c, dir, "approve", "--plan-id", planID)
	if err != nil {
		t.Fatalf("approve: %v", err)
	}

	// Status should reflect the approved plan
	stdout, _, err := runCmd(t, ctx, c, dir, "status", "--output", "json")
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	var result map[string]any
	parseJSONOutput(t, stdout, &result)
	if fmt.Sprintf("%v", result["plan_count"]) == "0" {
		t.Error("expected plan_count > 0 in status output")
	}
}
