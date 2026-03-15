package session_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/session"
)

func TestPlanStore_SaveAndLoad(t *testing.T) {
	// given
	dir := t.TempDir()
	store := &session.PlanStore{StateDir: dir}
	plan := domain.NewPlan("test.js", domain.TargetConfig{URL: "http://localhost"}, domain.LoadConfig{VUs: 10, Duration: "30s"}, domain.NfrConfig{})

	// when
	if err := store.SavePlan(plan); err != nil {
		t.Fatalf("SavePlan: %v", err)
	}
	loaded, err := store.LoadPlan(plan.ID)

	// then
	if err != nil {
		t.Fatalf("LoadPlan: %v", err)
	}
	if loaded.ID != plan.ID {
		t.Errorf("ID = %s, want %s", loaded.ID, plan.ID)
	}
	if loaded.Script != plan.Script {
		t.Errorf("Script = %s, want %s", loaded.Script, plan.Script)
	}
	if loaded.Approved {
		t.Error("plan should not be approved")
	}
}

func TestPlanStore_LoadPlan_NotFound(t *testing.T) {
	// given
	dir := t.TempDir()
	store := &session.PlanStore{StateDir: dir}

	// when
	_, err := store.LoadPlan("nonexistent")

	// then
	if err == nil {
		t.Fatal("expected error for nonexistent plan")
	}
}

func TestPlanStore_ApprovePlan(t *testing.T) {
	// given
	dir := t.TempDir()
	store := &session.PlanStore{StateDir: dir}
	plan := domain.NewPlan("test.js", domain.TargetConfig{}, domain.LoadConfig{}, domain.NfrConfig{})
	if err := store.SavePlan(plan); err != nil {
		t.Fatalf("SavePlan: %v", err)
	}

	// when
	approved, err := store.ApprovePlan(plan.ID)

	// then
	if err != nil {
		t.Fatalf("ApprovePlan: %v", err)
	}
	if !approved.Approved {
		t.Error("plan should be approved")
	}
	if approved.ApprovedAt.IsZero() {
		t.Error("approved_at should be set")
	}

	// verify persisted
	reloaded, err := store.LoadPlan(plan.ID)
	if err != nil {
		t.Fatalf("LoadPlan after approve: %v", err)
	}
	if !reloaded.Approved {
		t.Error("persisted plan should be approved")
	}
}

func TestPlanStore_ListScripts(t *testing.T) {
	// given
	dir := t.TempDir()
	k6Dir := filepath.Join(dir, "k6-scripts")
	if err := os.MkdirAll(k6Dir, 0o755); err != nil {
		t.Fatalf("create k6-scripts: %v", err)
	}
	for _, name := range []string{"load.js", "stress.js", "README.md"} {
		if err := os.WriteFile(filepath.Join(k6Dir, name), []byte("//test"), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	store := &session.PlanStore{StateDir: dir}

	// when
	scripts, err := store.ListScripts()

	// then
	if err != nil {
		t.Fatalf("ListScripts: %v", err)
	}
	if len(scripts) != 2 {
		t.Errorf("got %d scripts, want 2", len(scripts))
	}
}

func TestPlanStore_ListScripts_NoDir(t *testing.T) {
	// given
	dir := t.TempDir()
	store := &session.PlanStore{StateDir: dir}

	// when
	scripts, err := store.ListScripts()

	// then
	if err != nil {
		t.Fatalf("ListScripts: %v", err)
	}
	if len(scripts) != 0 {
		t.Errorf("got %d scripts, want 0", len(scripts))
	}
}

func TestPlanStore_LoadLatestPlan(t *testing.T) {
	// given
	dir := t.TempDir()
	store := &session.PlanStore{StateDir: dir}
	plan1 := domain.NewPlan("first.js", domain.TargetConfig{}, domain.LoadConfig{}, domain.NfrConfig{})
	plan2 := domain.NewPlan("second.js", domain.TargetConfig{}, domain.LoadConfig{}, domain.NfrConfig{})
	if err := store.SavePlan(plan1); err != nil {
		t.Fatalf("SavePlan 1: %v", err)
	}
	if err := store.SavePlan(plan2); err != nil {
		t.Fatalf("SavePlan 2: %v", err)
	}

	// when
	latest, err := store.LoadLatestPlan()

	// then
	if err != nil {
		t.Fatalf("LoadLatestPlan: %v", err)
	}
	// Latest by mtime should be plan2 (written second)
	if latest.ID != plan2.ID {
		t.Errorf("latest ID = %s, want %s", latest.ID, plan2.ID)
	}
}
