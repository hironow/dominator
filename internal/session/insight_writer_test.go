package session_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/session"
)

func TestRecordHue_AppendsToFile(t *testing.T) {
	// given
	dir := t.TempDir()
	writer := &session.InsightWriter{StateDir: dir, Logger: &domain.NopLogger{}}
	result := domain.JudgedData{
		PlanID:     "plan-001",
		ScriptPath: "test.js",
		Duration:   "30s",
		VUs:        10,
		Verdict:    domain.VerdictPass,
	}

	// when
	err := writer.RecordHue(result)

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	path := filepath.Join(dir, "insights", "hue.md")
	data, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("failed to read hue.md: %v", readErr)
	}
	content := string(data)
	if !strings.Contains(content, "pass") {
		t.Errorf("expected hue.md to contain 'pass', got: %s", content)
	}
	if !strings.Contains(content, "test.js") {
		t.Errorf("expected hue.md to contain script path, got: %s", content)
	}
	if !strings.Contains(content, "Deviations: 0") {
		t.Errorf("expected hue.md to contain 'Deviations: 0', got: %s", content)
	}
}

func TestRecordHue_AppendsMultipleEntries(t *testing.T) {
	// given
	dir := t.TempDir()
	writer := &session.InsightWriter{StateDir: dir, Logger: &domain.NopLogger{}}
	result := domain.JudgedData{
		PlanID:  "plan-001",
		Verdict: domain.VerdictPass,
	}

	// when: write twice
	if err := writer.RecordHue(result); err != nil {
		t.Fatalf("first write: %v", err)
	}
	if err := writer.RecordHue(result); err != nil {
		t.Fatalf("second write: %v", err)
	}

	// then: both entries should be present
	path := filepath.Join(dir, "insights", "hue.md")
	data, _ := os.ReadFile(path)
	count := strings.Count(string(data), "## ")
	if count != 2 {
		t.Errorf("expected 2 entries (## headers), got %d", count)
	}
}

func TestRecordCoefficient_ViolationOnly(t *testing.T) {
	// given
	dir := t.TempDir()
	writer := &session.InsightWriter{StateDir: dir, Logger: &domain.NopLogger{}}

	passResult := domain.JudgedData{
		PlanID:  "plan-pass",
		Verdict: domain.VerdictPass,
	}
	violationResult := domain.JudgedData{
		PlanID:  "plan-violation",
		Verdict: domain.VerdictViolation,
		Deviations: []domain.NfrDeviation{
			{
				Metric:    "p95_latency_ms",
				Threshold: 500,
				Actual:    750,
				Deviation: 50,
				Severity:  domain.SeverityMedium,
			},
		},
	}

	// when: record pass (should not write)
	if err := writer.RecordCoefficient(passResult); err != nil {
		t.Fatalf("pass result: %v", err)
	}
	path := filepath.Join(dir, "insights", "coefficient.md")
	if _, err := os.Stat(path); err == nil {
		t.Fatal("coefficient.md should not be created for pass verdict")
	}

	// when: record violation (should write)
	if err := writer.RecordCoefficient(violationResult); err != nil {
		t.Fatalf("violation result: %v", err)
	}

	// then
	data, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("failed to read coefficient.md: %v", readErr)
	}
	content := string(data)
	if !strings.Contains(content, "Violation") {
		t.Errorf("expected coefficient.md to contain 'Violation', got: %s", content)
	}
	if !strings.Contains(content, "p95_latency_ms") {
		t.Errorf("expected coefficient.md to contain metric name, got: %s", content)
	}
}
