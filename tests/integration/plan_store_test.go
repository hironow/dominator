//go:build integration

package integration_test

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/session"
)

// TestPlanStore_FullLifecycle tests create, load, approve, and list operations.
func TestPlanStore_FullLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// given: temp state directory with required structure
	dir := t.TempDir()
	stateDir := filepath.Join(dir, ".pass")
	if err := session.InitPassDir(stateDir, &domain.NopLogger{}); err != nil {
		t.Fatalf("InitPassDir: %v", err)
	}

	store := &session.PlanStore{StateDir: stateDir}
	cfg := domain.DefaultConfig()

	// Write a fake k6 script
	scriptPath := filepath.Join(stateDir, "k6-scripts", "test-api.js")
	if err := os.WriteFile(scriptPath, []byte("// k6 script"), 0o644); err != nil {
		t.Fatalf("write script: %v", err)
	}

	// when: create and save a plan
	plan := domain.NewPlan("test-api.js", cfg.Target, cfg.Load, cfg.Nfr)
	if err := store.SavePlan(plan); err != nil {
		t.Fatalf("SavePlan: %v", err)
	}

	// then: load by ID
	loaded, err := store.LoadPlan(plan.ID)
	if err != nil {
		t.Fatalf("LoadPlan: %v", err)
	}
	if loaded.ID != plan.ID {
		t.Errorf("expected plan ID %s, got %s", plan.ID, loaded.ID)
	}
	if loaded.Script != "test-api.js" {
		t.Errorf("expected script 'test-api.js', got %q", loaded.Script)
	}
	if loaded.Approved {
		t.Error("expected plan not to be approved initially")
	}

	// then: load latest
	latest, err := store.LoadLatestPlan()
	if err != nil {
		t.Fatalf("LoadLatestPlan: %v", err)
	}
	if latest.ID != plan.ID {
		t.Errorf("LoadLatestPlan: expected ID %s, got %s", plan.ID, latest.ID)
	}

	// when: approve
	approved, err := store.ApprovePlan(plan.ID)
	if err != nil {
		t.Fatalf("ApprovePlan: %v", err)
	}
	if !approved.Approved {
		t.Error("expected plan to be approved after ApprovePlan")
	}
	if approved.ApprovedAt.IsZero() {
		t.Error("expected ApprovedAt to be set")
	}

	// then: reload and verify approved state persisted
	reloaded, err := store.LoadPlan(plan.ID)
	if err != nil {
		t.Fatalf("reload after approve: %v", err)
	}
	if !reloaded.Approved {
		t.Error("expected approved state to persist after reload")
	}

	// then: list scripts
	scripts, err := store.ListScripts()
	if err != nil {
		t.Fatalf("ListScripts: %v", err)
	}
	if len(scripts) != 1 {
		t.Errorf("expected 1 script, got %d", len(scripts))
	}
	if len(scripts) > 0 && scripts[0] != "test-api.js" {
		t.Errorf("expected script 'test-api.js', got %q", scripts[0])
	}
}

// TestPlanStore_ConcurrentAccess tests that multiple goroutines can
// create and load plans without data corruption.
func TestPlanStore_ConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// given
	dir := t.TempDir()
	stateDir := filepath.Join(dir, ".pass")
	if err := session.InitPassDir(stateDir, &domain.NopLogger{}); err != nil {
		t.Fatalf("InitPassDir: %v", err)
	}
	store := &session.PlanStore{StateDir: stateDir}
	cfg := domain.DefaultConfig()

	const concurrency = 10
	var wg sync.WaitGroup
	errs := make(chan error, concurrency)
	planIDs := make(chan domain.PlanID, concurrency)

	// when: create plans concurrently
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			plan := domain.NewPlan("concurrent-test.js", cfg.Target, cfg.Load, cfg.Nfr)
			if err := store.SavePlan(plan); err != nil {
				errs <- err
				return
			}
			planIDs <- plan.ID
		}()
	}
	wg.Wait()
	close(errs)
	close(planIDs)

	// then: no errors
	for err := range errs {
		t.Errorf("concurrent SavePlan error: %v", err)
	}

	// then: all plans are loadable
	for id := range planIDs {
		loaded, err := store.LoadPlan(id)
		if err != nil {
			t.Errorf("LoadPlan(%s) after concurrent save: %v", id, err)
			continue
		}
		if loaded.ID != id {
			t.Errorf("expected ID %s, got %s", id, loaded.ID)
		}
	}
}

// TestPlanStore_LoadNonExistent verifies that loading a non-existent plan
// returns a meaningful error.
func TestPlanStore_LoadNonExistent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// given
	dir := t.TempDir()
	stateDir := filepath.Join(dir, ".pass")
	if err := session.InitPassDir(stateDir, &domain.NopLogger{}); err != nil {
		t.Fatalf("InitPassDir: %v", err)
	}
	store := &session.PlanStore{StateDir: stateDir}

	// when
	_, err := store.LoadPlan("nonexistent-plan-id")

	// then
	if err == nil {
		t.Fatal("expected error when loading non-existent plan")
	}
}

// TestPlanStore_MultiplePlansLatest verifies LoadLatestPlan returns the
// most recently created plan when multiple plans exist.
func TestPlanStore_MultiplePlansLatest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// given
	dir := t.TempDir()
	stateDir := filepath.Join(dir, ".pass")
	if err := session.InitPassDir(stateDir, &domain.NopLogger{}); err != nil {
		t.Fatalf("InitPassDir: %v", err)
	}
	store := &session.PlanStore{StateDir: stateDir}
	cfg := domain.DefaultConfig()

	// when: create two plans
	plan1 := domain.NewPlan("first.js", cfg.Target, cfg.Load, cfg.Nfr)
	if err := store.SavePlan(plan1); err != nil {
		t.Fatalf("save plan1: %v", err)
	}

	plan2 := domain.NewPlan("second.js", cfg.Target, cfg.Load, cfg.Nfr)
	if err := store.SavePlan(plan2); err != nil {
		t.Fatalf("save plan2: %v", err)
	}

	// then: latest should be plan2 (most recently modified)
	latest, err := store.LoadLatestPlan()
	if err != nil {
		t.Fatalf("LoadLatestPlan: %v", err)
	}
	if latest.Script != "second.js" {
		t.Errorf("expected latest plan script to be 'second.js', got %q", latest.Script)
	}
}
