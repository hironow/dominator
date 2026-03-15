//go:build contract

package contract_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/session"
)

// TestDMailFormat_YAMLFrontmatter verifies that D-Mail output files follow
// the YAML frontmatter + Markdown format.
func TestDMailFormat_YAMLFrontmatter(t *testing.T) {
	// given: a DMailEmitter configured with a temp directory
	dir := t.TempDir()
	stateDir := filepath.Join(dir, ".pass")
	if err := session.InitPassDir(stateDir, &domain.NopLogger{}); err != nil {
		t.Fatalf("InitPassDir: %v", err)
	}

	emitter := &session.DMailEmitter{
		StateDir: stateDir,
		Logger:   &domain.NopLogger{},
	}

	// when: emit a violation D-Mail
	result := domain.JudgedData{
		PlanID:     "test-plan-001",
		ScriptPath: "load-test.js",
		Duration:   "30s",
		VUs:        10,
		Results: domain.K6Results{
			P95LatencyMs:     800,
			ErrorRatePercent: 2.0,
			SuccessRate:      95.0,
			ActualRPS:        80,
		},
		Verdict: domain.VerdictViolation,
		Deviations: []domain.NfrDeviation{
			{
				Metric:    "p95_latency_ms",
				Threshold: 500,
				Actual:    800,
				Deviation: 60.0,
				Severity:  domain.SeverityHigh,
			},
		},
	}

	if err := emitter.EmitViolation(result); err != nil {
		t.Fatalf("EmitViolation: %v", err)
	}

	// then: verify D-Mail files exist in outbox
	outboxDir := filepath.Join(stateDir, "outbox")
	entries, err := os.ReadDir(outboxDir)
	if err != nil {
		t.Fatalf("ReadDir outbox: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("expected D-Mail files in outbox, got none")
	}

	// Verify each D-Mail file has YAML frontmatter
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(outboxDir, entry.Name()))
		if err != nil {
			t.Fatalf("read D-Mail %s: %v", entry.Name(), err)
		}

		text := string(content)

		// Must start with YAML frontmatter delimiter
		if !strings.HasPrefix(text, "---\n") {
			t.Errorf("D-Mail %s must start with '---' YAML frontmatter delimiter", entry.Name())
		}

		// Must have closing frontmatter delimiter
		parts := strings.SplitN(text, "---\n", 3)
		if len(parts) < 3 {
			t.Errorf("D-Mail %s must have opening and closing '---' frontmatter delimiters", entry.Name())
			continue
		}

		// Frontmatter must contain required fields
		frontmatter := parts[1]
		requiredFields := []string{"name:", "kind:", "description:", "severity:"}
		for _, field := range requiredFields {
			if !strings.Contains(frontmatter, field) {
				t.Errorf("D-Mail %s frontmatter missing required field %q", entry.Name(), field)
			}
		}

		// Body must contain markdown content
		body := parts[2]
		if !strings.Contains(body, "#") {
			t.Errorf("D-Mail %s body should contain markdown heading", entry.Name())
		}
	}
}

// TestDMailFormat_PassReport verifies pass D-Mail format.
func TestDMailFormat_PassReport(t *testing.T) {
	// given
	dir := t.TempDir()
	stateDir := filepath.Join(dir, ".pass")
	if err := session.InitPassDir(stateDir, &domain.NopLogger{}); err != nil {
		t.Fatalf("InitPassDir: %v", err)
	}

	emitter := &session.DMailEmitter{
		StateDir: stateDir,
		Logger:   &domain.NopLogger{},
	}

	result := domain.JudgedData{
		PlanID:     "test-plan-pass",
		ScriptPath: "load-test.js",
		Duration:   "30s",
		VUs:        10,
		Verdict:    domain.VerdictPass,
	}

	// when
	if err := emitter.EmitPass(result); err != nil {
		t.Fatalf("EmitPass: %v", err)
	}

	// then
	outboxDir := filepath.Join(stateDir, "outbox")
	entries, err := os.ReadDir(outboxDir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}

	found := false
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "nfr-pass-") {
			found = true
			content, err := os.ReadFile(filepath.Join(outboxDir, entry.Name()))
			if err != nil {
				t.Fatalf("read: %v", err)
			}
			text := string(content)
			if !strings.Contains(text, "kind: report") {
				t.Error("pass D-Mail should have kind: report")
			}
			if !strings.Contains(text, "Verdict: pass") {
				t.Error("pass D-Mail body should contain 'Verdict: pass'")
			}
		}
	}
	if !found {
		t.Error("expected nfr-pass-*.md in outbox")
	}
}

// TestPlanJSON_Schema verifies that plan JSON has all required fields.
func TestPlanJSON_Schema(t *testing.T) {
	// given: create a plan
	cfg := domain.DefaultConfig()
	plan := domain.NewPlan("test.js", cfg.Target, cfg.Load, cfg.Nfr)

	// when: marshal to JSON
	data, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		t.Fatalf("marshal plan: %v", err)
	}

	// then: unmarshal into a generic map to verify fields
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}

	requiredFields := []string{
		"plan_id",
		"script",
		"target",
		"load",
		"nfr",
		"approved",
		"created_at",
	}

	for _, field := range requiredFields {
		if _, ok := m[field]; !ok {
			t.Errorf("plan JSON missing required field %q", field)
		}
	}

	// Verify plan_id is a non-empty string
	planID, ok := m["plan_id"].(string)
	if !ok || planID == "" {
		t.Error("plan_id must be a non-empty string")
	}

	// Verify approved is false initially
	approved, ok := m["approved"].(bool)
	if !ok || approved {
		t.Error("plan should not be approved initially")
	}

	// Verify nested structures
	// Note: TargetConfig/LoadConfig/NfrConfig use yaml tags but no json tags,
	// so JSON marshaling uses Go field names (capitalized).
	target, ok := m["target"].(map[string]interface{})
	if !ok {
		t.Error("target must be an object")
	} else {
		for _, field := range []string{"URL", "Protocol", "Spec", "Docs"} {
			if _, ok := target[field]; !ok {
				t.Errorf("target missing field %q", field)
			}
		}
	}

	load, ok := m["load"].(map[string]interface{})
	if !ok {
		t.Error("load must be an object")
	} else {
		for _, field := range []string{"VUs", "Duration", "RampUp"} {
			if _, ok := load[field]; !ok {
				t.Errorf("load missing field %q", field)
			}
		}
	}

	nfr, ok := m["nfr"].(map[string]interface{})
	if !ok {
		t.Error("nfr must be an object")
	} else {
		for _, field := range []string{"Performance", "Reliability", "Scalability"} {
			if _, ok := nfr[field]; !ok {
				t.Errorf("nfr missing field %q", field)
			}
		}
	}
}

// TestPlanJSON_ApprovedPlanHasTimestamp verifies approved_at is set after approval.
func TestPlanJSON_ApprovedPlanHasTimestamp(t *testing.T) {
	// given
	cfg := domain.DefaultConfig()
	plan := domain.NewPlan("test.js", cfg.Target, cfg.Load, cfg.Nfr)

	// when
	plan.Approve()

	// then
	data, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	approvedAt, ok := m["approved_at"].(string)
	if !ok || approvedAt == "" {
		t.Error("approved plan must have approved_at timestamp")
	}

	// Verify timestamp is parseable
	if _, err := time.Parse(time.RFC3339Nano, approvedAt); err != nil {
		t.Errorf("approved_at is not a valid RFC3339 timestamp: %v", err)
	}
}

// TestK6SummaryFormat verifies k6 JSON summary parsing expectations.
func TestK6SummaryFormat(t *testing.T) {
	// This test validates that K6Results can be marshaled/unmarshaled correctly.
	testCases := []struct {
		name    string
		results domain.K6Results
	}{
		{
			name: "typical_results",
			results: domain.K6Results{
				P95LatencyMs:     250.5,
				ErrorRatePercent: 0.5,
				SuccessRate:      99.5,
				ActualRPS:        150.3,
			},
		},
		{
			name: "zero_values",
			results: domain.K6Results{
				P95LatencyMs:     0,
				ErrorRatePercent: 0,
				SuccessRate:      100.0,
				ActualRPS:        0,
			},
		},
		{
			name: "high_latency",
			results: domain.K6Results{
				P95LatencyMs:     5000.0,
				ErrorRatePercent: 10.0,
				SuccessRate:      50.0,
				ActualRPS:        25.0,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// when: marshal
			data, err := json.Marshal(tc.results)
			if err != nil {
				t.Fatalf("marshal K6Results: %v", err)
			}

			// then: unmarshal back
			var loaded domain.K6Results
			if err := json.Unmarshal(data, &loaded); err != nil {
				t.Fatalf("unmarshal K6Results: %v", err)
			}

			// Verify round-trip
			if loaded.P95LatencyMs != tc.results.P95LatencyMs {
				t.Errorf("P95LatencyMs: want %f, got %f", tc.results.P95LatencyMs, loaded.P95LatencyMs)
			}
			if loaded.ErrorRatePercent != tc.results.ErrorRatePercent {
				t.Errorf("ErrorRatePercent: want %f, got %f", tc.results.ErrorRatePercent, loaded.ErrorRatePercent)
			}
			if loaded.SuccessRate != tc.results.SuccessRate {
				t.Errorf("SuccessRate: want %f, got %f", tc.results.SuccessRate, loaded.SuccessRate)
			}
			if loaded.ActualRPS != tc.results.ActualRPS {
				t.Errorf("ActualRPS: want %f, got %f", tc.results.ActualRPS, loaded.ActualRPS)
			}

			// Verify JSON keys exist
			var m map[string]interface{}
			json.Unmarshal(data, &m)
			for _, key := range []string{"p95_latency_ms", "error_rate_percent", "success_rate", "actual_rps"} {
				if _, ok := m[key]; !ok {
					t.Errorf("K6Results JSON missing key %q", key)
				}
			}
		})
	}
}

// TestEventJSON_Schema verifies event JSON envelope format.
func TestEventJSON_Schema(t *testing.T) {
	// given
	ev, err := domain.NewEvent(domain.EventPlanCreated, map[string]string{
		"plan_id": "test-001",
	}, time.Now().UTC())
	if err != nil {
		t.Fatalf("NewEvent: %v", err)
	}

	// when
	data, err := json.Marshal(ev)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}

	// then
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	requiredFields := []string{"id", "type", "timestamp", "tool", "data"}
	for _, field := range requiredFields {
		if _, ok := m[field]; !ok {
			t.Errorf("event JSON missing required field %q", field)
		}
	}

	// Verify tool is "dominator"
	if tool, _ := m["tool"].(string); tool != "dominator" {
		t.Errorf("expected tool 'dominator', got %q", tool)
	}

	// Verify type
	if typ, _ := m["type"].(string); typ != string(domain.EventPlanCreated) {
		t.Errorf("expected type %q, got %q", domain.EventPlanCreated, typ)
	}

	// Verify timestamp is parseable RFC3339
	if ts, _ := m["timestamp"].(string); ts != "" {
		if _, err := time.Parse(time.RFC3339Nano, ts); err != nil {
			t.Errorf("timestamp is not valid RFC3339: %v", err)
		}
	}
}
