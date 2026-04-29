package session_test

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/session"
	"gopkg.in/yaml.v3"
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

func TestPlanStore_ManyPlans(t *testing.T) {
	// given
	dir := t.TempDir()
	store := &session.PlanStore{StateDir: dir}
	var plans []domain.Plan
	for i := 0; i < 10; i++ {
		plan := domain.NewPlan(
			fmt.Sprintf("script-%d.js", i),
			domain.TargetConfig{URL: "http://localhost"},
			domain.LoadConfig{VUs: i + 1, Duration: "30s"},
			domain.NfrConfig{},
		)
		if err := store.SavePlan(plan); err != nil {
			t.Fatalf("SavePlan %d: %v", i, err)
		}
		plans = append(plans, plan)
	}

	// when: load each plan
	for _, p := range plans {
		loaded, err := store.LoadPlan(p.ID)
		if err != nil {
			t.Fatalf("LoadPlan %s: %v", p.ID, err)
		}
		if loaded.Script != p.Script {
			t.Errorf("Script = %q, want %q", loaded.Script, p.Script)
		}
	}
}

func TestPlanStore_ApproveNonExistent(t *testing.T) {
	// given
	dir := t.TempDir()
	store := &session.PlanStore{StateDir: dir}

	// when
	_, err := store.ApprovePlan("nonexistent-plan-id")

	// then
	if err == nil {
		t.Fatal("expected error for nonexistent plan")
	}
}

func TestPlanStore_JSONRoundTrip(t *testing.T) {
	// given
	dir := t.TempDir()
	store := &session.PlanStore{StateDir: dir}
	plan := domain.NewPlan(
		"full-test.js",
		domain.TargetConfig{URL: "http://localhost:8080", Protocol: "openapi", Spec: "https://example.com/spec", Docs: "https://example.com/docs"},
		domain.LoadConfig{VUs: 20, Duration: "60s", RampUp: "10s"},
		domain.NfrConfig{
			Performance: domain.PerformanceNfr{P95LatencyMs: 500, ErrorRatePercent: 1.0},
			Reliability: domain.ReliabilityNfr{SuccessRatePercent: 99.0},
			Scalability: domain.ScalabilityNfr{TargetRPS: 200},
		},
	)
	plan.Approve()

	// when
	if err := store.SavePlan(plan); err != nil {
		t.Fatalf("SavePlan: %v", err)
	}
	loaded, err := store.LoadPlan(plan.ID)
	if err != nil {
		t.Fatalf("LoadPlan: %v", err)
	}

	// then: all fields survive
	if loaded.ID != plan.ID {
		t.Errorf("ID = %s, want %s", loaded.ID, plan.ID)
	}
	if loaded.Script != plan.Script {
		t.Errorf("Script = %q, want %q", loaded.Script, plan.Script)
	}
	if loaded.Target.URL != plan.Target.URL {
		t.Errorf("Target.URL = %q, want %q", loaded.Target.URL, plan.Target.URL)
	}
	if loaded.Target.Protocol != plan.Target.Protocol {
		t.Errorf("Target.Protocol = %q, want %q", loaded.Target.Protocol, plan.Target.Protocol)
	}
	if loaded.Load.VUs != plan.Load.VUs {
		t.Errorf("Load.VUs = %d, want %d", loaded.Load.VUs, plan.Load.VUs)
	}
	if loaded.Load.Duration != plan.Load.Duration {
		t.Errorf("Load.Duration = %q, want %q", loaded.Load.Duration, plan.Load.Duration)
	}
	if loaded.Load.RampUp != plan.Load.RampUp {
		t.Errorf("Load.RampUp = %q, want %q", loaded.Load.RampUp, plan.Load.RampUp)
	}
	if loaded.Nfr.Performance.P95LatencyMs != plan.Nfr.Performance.P95LatencyMs {
		t.Errorf("P95LatencyMs = %d, want %d", loaded.Nfr.Performance.P95LatencyMs, plan.Nfr.Performance.P95LatencyMs)
	}
	if loaded.Nfr.Scalability.TargetRPS != plan.Nfr.Scalability.TargetRPS {
		t.Errorf("TargetRPS = %d, want %d", loaded.Nfr.Scalability.TargetRPS, plan.Nfr.Scalability.TargetRPS)
	}
	if !loaded.Approved {
		t.Error("expected plan to be approved")
	}
	if loaded.ApprovedAt.IsZero() {
		t.Error("expected ApprovedAt to be set")
	}
}

func TestPlanStore_ConcurrentApprove(t *testing.T) {
	// given
	dir := t.TempDir()
	store := &session.PlanStore{StateDir: dir}
	var plans []domain.Plan
	for i := 0; i < 5; i++ {
		plan := domain.NewPlan(fmt.Sprintf("script-%d.js", i), domain.TargetConfig{}, domain.LoadConfig{}, domain.NfrConfig{})
		if err := store.SavePlan(plan); err != nil {
			t.Fatalf("SavePlan %d: %v", i, err)
		}
		plans = append(plans, plan)
	}

	// when: approve all plans concurrently
	var wg sync.WaitGroup
	errs := make([]error, len(plans))
	for i, p := range plans {
		wg.Add(1)
		go func(idx int, id domain.PlanID) {
			defer wg.Done()
			_, err := store.ApprovePlan(id)
			errs[idx] = err
		}(i, p.ID)
	}
	wg.Wait()

	// then: all should succeed
	for i, err := range errs {
		if err != nil {
			t.Errorf("approve %d failed: %v", i, err)
		}
	}

	// Verify all approved
	for _, p := range plans {
		loaded, err := store.LoadPlan(p.ID)
		if err != nil {
			t.Fatalf("LoadPlan %s: %v", p.ID, err)
		}
		if !loaded.Approved {
			t.Errorf("plan %s should be approved", p.ID)
		}
	}
}

func TestPlanStore_LoadLatestPlan_NoPlanDir(t *testing.T) {
	dir := t.TempDir()
	store := &session.PlanStore{StateDir: dir}
	_, err := store.LoadLatestPlan()
	if err == nil {
		t.Fatal("expected error when no plans dir exists")
	}
}

func TestPlanStore_ListScripts_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	k6Dir := filepath.Join(dir, "k6-scripts")
	if err := os.MkdirAll(k6Dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	store := &session.PlanStore{StateDir: dir}
	scripts, err := store.ListScripts()
	if err != nil {
		t.Fatalf("ListScripts: %v", err)
	}
	if len(scripts) != 0 {
		t.Errorf("expected 0 scripts, got %d", len(scripts))
	}
}

func TestPlanStore_SavePlan_CreatesDir(t *testing.T) {
	dir := t.TempDir()
	store := &session.PlanStore{StateDir: dir}
	plan := domain.NewPlan("test.js", domain.TargetConfig{}, domain.LoadConfig{}, domain.NfrConfig{})

	// .run/plans/ should not exist yet
	plansDir := filepath.Join(dir, ".run", "plans")
	if _, err := os.Stat(plansDir); err == nil {
		t.Fatal("plans dir should not exist yet")
	}

	err := store.SavePlan(plan)
	if err != nil {
		t.Fatalf("SavePlan: %v", err)
	}

	if _, err := os.Stat(plansDir); err != nil {
		t.Errorf("plans dir should exist after SavePlan: %v", err)
	}
}

func TestPlanStore_ApproveTwice_Idempotent(t *testing.T) {
	dir := t.TempDir()
	store := &session.PlanStore{StateDir: dir}
	plan := domain.NewPlan("test.js", domain.TargetConfig{}, domain.LoadConfig{}, domain.NfrConfig{})
	store.SavePlan(plan)

	// First approve
	approved1, err := store.ApprovePlan(plan.ID)
	if err != nil {
		t.Fatalf("first approve: %v", err)
	}
	if !approved1.Approved {
		t.Error("expected approved after first approve")
	}

	// Second approve (idempotent)
	approved2, err := store.ApprovePlan(plan.ID)
	if err != nil {
		t.Fatalf("second approve: %v", err)
	}
	if !approved2.Approved {
		t.Error("expected approved after second approve")
	}
}

func TestPlanStore_LoadPlan_PreservesAllFields(t *testing.T) {
	dir := t.TempDir()
	store := &session.PlanStore{StateDir: dir}

	target := domain.TargetConfig{URL: "http://api.example.com", Protocol: "openapi", Spec: "https://example.com/spec"}
	load := domain.LoadConfig{VUs: 25, Duration: "45s", RampUp: "8s"}
	nfr := domain.NfrConfig{
		Performance: domain.PerformanceNfr{P95LatencyMs: 200, ErrorRatePercent: 0.5},
		Reliability: domain.ReliabilityNfr{SuccessRatePercent: 99.5},
		Scalability: domain.ScalabilityNfr{TargetRPS: 300},
	}
	plan := domain.NewPlan("comprehensive.js", target, load, nfr)
	store.SavePlan(plan)

	loaded, err := store.LoadPlan(plan.ID)
	if err != nil {
		t.Fatalf("LoadPlan: %v", err)
	}

	if loaded.Target.Spec != "https://example.com/spec" {
		t.Errorf("Target.Spec = %q, want %q", loaded.Target.Spec, "https://example.com/spec")
	}
	if loaded.Load.RampUp != "8s" {
		t.Errorf("Load.RampUp = %q, want %q", loaded.Load.RampUp, "8s")
	}
	if loaded.Nfr.Performance.ErrorRatePercent != 0.5 {
		t.Errorf("ErrorRatePercent = %f, want 0.5", loaded.Nfr.Performance.ErrorRatePercent)
	}
}

func TestPlanStore_ListScripts_OnlyJS(t *testing.T) {
	dir := t.TempDir()
	k6Dir := filepath.Join(dir, "k6-scripts")
	os.MkdirAll(k6Dir, 0o755)

	files := map[string]bool{
		"load.js":     true,
		"stress.js":   true,
		"spike.js":    true,
		"README.md":   false,
		"config.yaml": false,
		"helper.ts":   false,
	}
	for name := range files {
		os.WriteFile(filepath.Join(k6Dir, name), []byte("//"), 0o644)
	}

	store := &session.PlanStore{StateDir: dir}
	scripts, err := store.ListScripts()
	if err != nil {
		t.Fatalf("ListScripts: %v", err)
	}
	if len(scripts) != 3 {
		t.Errorf("expected 3 .js scripts, got %d: %v", len(scripts), scripts)
	}
}

func TestUpdateConfig_AllKeys(t *testing.T) {
	tests := []struct {
		key   string
		value string
	}{
		{key: "target.url", value: "http://localhost:9090"},
		{key: "target.protocol", value: "json-rpc"},
		{key: "target.spec", value: "https://example.com/spec"},
		{key: "target.docs", value: "https://example.com/docs"},
		{key: "nfr.performance.p95_latency_ms", value: "300"},
		{key: "nfr.performance.error_rate_percent", value: "0.5"},
		{key: "nfr.reliability.success_rate_percent", value: "99.5"},
		{key: "nfr.scalability.target_rps", value: "200"},
		{key: "load.vus", value: "20"},
		{key: "load.duration", value: "60s"},
		{key: "load.ramp_up", value: "10s"},
		{key: "approval.required", value: "false"},
		{key: "lang", value: "en"},
		{key: "claude_cmd", value: "claude-dev"},
		{key: "model", value: "sonnet"},
		{key: "timeout_sec", value: "300"},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			dir := t.TempDir()
			// Write default config
			cfg := domain.DefaultConfig()
			cfgPath := filepath.Join(dir, domain.ConfigFile)
			data, _ := yaml.Marshal(cfg)
			os.WriteFile(cfgPath, data, 0o644)

			err := session.UpdateConfig(dir, tt.key, tt.value)
			if err != nil {
				t.Fatalf("UpdateConfig(%q, %q): %v", tt.key, tt.value, err)
			}
		})
	}
}

func TestUpdateConfig_InvalidKey(t *testing.T) {
	dir := t.TempDir()
	cfg := domain.DefaultConfig()
	cfgPath := filepath.Join(dir, domain.ConfigFile)
	data, _ := yaml.Marshal(cfg)
	os.WriteFile(cfgPath, data, 0o644)

	err := session.UpdateConfig(dir, "unknown.key", "value")
	if err == nil {
		t.Fatal("expected error for unknown key")
	}
}

func TestUpdateConfig_InvalidValue(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{name: "non_integer_latency", key: "nfr.performance.p95_latency_ms", value: "abc"},
		{name: "non_float_error_rate", key: "nfr.performance.error_rate_percent", value: "xyz"},
		{name: "non_integer_vus", key: "load.vus", value: "not-a-number"},
		{name: "non_boolean_approval", key: "approval.required", value: "maybe"},
		{name: "non_integer_timeout", key: "timeout_sec", value: "slow"},
		{name: "non_integer_rps", key: "nfr.scalability.target_rps", value: "fast"},
		{name: "non_float_success", key: "nfr.reliability.success_rate_percent", value: "high"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			cfg := domain.DefaultConfig()
			cfgPath := filepath.Join(dir, domain.ConfigFile)
			data, _ := yaml.Marshal(cfg)
			os.WriteFile(cfgPath, data, 0o644)

			err := session.UpdateConfig(dir, tt.key, tt.value)
			if err == nil {
				t.Fatalf("expected error for invalid value %q on key %q", tt.value, tt.key)
			}
		})
	}
}
