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

func TestRecordHue_MultipleEntries(t *testing.T) {
	dir := t.TempDir()
	writer := &session.InsightWriter{StateDir: dir, Logger: &domain.NopLogger{}}

	entries := []domain.JudgedData{
		{PlanID: "plan-1", ScriptPath: "test1.js", Duration: "30s", VUs: 10, Verdict: domain.VerdictPass},
		{PlanID: "plan-2", ScriptPath: "test2.js", Duration: "60s", VUs: 20, Verdict: domain.VerdictViolation, Deviations: []domain.NfrDeviation{{Metric: "m1"}}},
		{PlanID: "plan-3", ScriptPath: "test3.js", Duration: "90s", VUs: 30, Verdict: domain.VerdictPass},
	}

	for _, e := range entries {
		if err := writer.RecordHue(e); err != nil {
			t.Fatalf("RecordHue: %v", err)
		}
	}

	path := filepath.Join(dir, "insights", "hue.md")
	data, _ := os.ReadFile(path)
	content := string(data)

	// Verify all 3 entries
	count := strings.Count(content, "## ")
	if count != 3 {
		t.Errorf("expected 3 entries, got %d", count)
	}
	if !strings.Contains(content, "test1.js") {
		t.Error("missing test1.js")
	}
	if !strings.Contains(content, "test2.js") {
		t.Error("missing test2.js")
	}
	if !strings.Contains(content, "test3.js") {
		t.Error("missing test3.js")
	}
}

func TestRecordHue_PassEntry(t *testing.T) {
	dir := t.TempDir()
	writer := &session.InsightWriter{StateDir: dir, Logger: &domain.NopLogger{}}

	result := domain.JudgedData{
		PlanID:     "plan-pass",
		ScriptPath: "pass-test.js",
		Duration:   "30s",
		VUs:        10,
		Verdict:    domain.VerdictPass,
	}
	if err := writer.RecordHue(result); err != nil {
		t.Fatalf("RecordHue: %v", err)
	}

	path := filepath.Join(dir, "insights", "hue.md")
	data, _ := os.ReadFile(path)
	content := string(data)

	if !strings.Contains(content, "pass") {
		t.Error("pass entry should contain 'pass'")
	}
	if !strings.Contains(content, "Deviations: 0") {
		t.Error("pass entry should have Deviations: 0")
	}
}

func TestRecordHue_ViolationEntry(t *testing.T) {
	dir := t.TempDir()
	writer := &session.InsightWriter{StateDir: dir, Logger: &domain.NopLogger{}}

	result := domain.JudgedData{
		PlanID:     "plan-viol",
		ScriptPath: "viol-test.js",
		Duration:   "30s",
		VUs:        10,
		Verdict:    domain.VerdictViolation,
		Deviations: []domain.NfrDeviation{
			{Metric: "m1"}, {Metric: "m2"},
		},
	}
	if err := writer.RecordHue(result); err != nil {
		t.Fatalf("RecordHue: %v", err)
	}

	path := filepath.Join(dir, "insights", "hue.md")
	data, _ := os.ReadFile(path)
	content := string(data)

	if !strings.Contains(content, "violation") {
		t.Error("violation entry should contain 'violation'")
	}
	if !strings.Contains(content, "Deviations: 2") {
		t.Error("violation entry should have Deviations: 2")
	}
}

func TestRecordCoefficient_MultipleDeviations(t *testing.T) {
	dir := t.TempDir()
	writer := &session.InsightWriter{StateDir: dir, Logger: &domain.NopLogger{}}

	result := domain.JudgedData{
		PlanID:  "plan-multi-coeff",
		Verdict: domain.VerdictViolation,
		Deviations: []domain.NfrDeviation{
			{Metric: "p95_latency_ms", Threshold: 500, Actual: 800, Deviation: 60, Severity: domain.SeverityHigh},
			{Metric: "error_rate_percent", Threshold: 1.0, Actual: 5.0, Deviation: 400, Severity: domain.SeverityHigh},
			{Metric: "success_rate_percent", Threshold: 99.0, Actual: 95.0, Deviation: 4.04, Severity: domain.SeverityLow},
			{Metric: "target_rps", Threshold: 100, Actual: 50, Deviation: 50, Severity: domain.SeverityMedium},
		},
	}

	if err := writer.RecordCoefficient(result); err != nil {
		t.Fatalf("RecordCoefficient: %v", err)
	}

	path := filepath.Join(dir, "insights", "coefficient.md")
	data, _ := os.ReadFile(path)
	content := string(data)

	for _, metric := range []string{"p95_latency_ms", "error_rate_percent", "success_rate_percent", "target_rps"} {
		if !strings.Contains(content, metric) {
			t.Errorf("expected %q in coefficient.md", metric)
		}
	}
	// Should have a markdown table header
	if !strings.Contains(content, "| Metric |") {
		t.Error("expected markdown table header")
	}
}

func TestRecordHue_CreatesInsightsDir(t *testing.T) {
	dir := t.TempDir()
	writer := &session.InsightWriter{StateDir: dir, Logger: &domain.NopLogger{}}
	insightsDir := filepath.Join(dir, "insights")

	// Should not exist yet
	if _, err := os.Stat(insightsDir); err == nil {
		t.Fatal("insights dir should not exist yet")
	}

	result := domain.JudgedData{PlanID: "dir-test", Verdict: domain.VerdictPass}
	writer.RecordHue(result)

	if _, err := os.Stat(insightsDir); err != nil {
		t.Errorf("insights dir should exist: %v", err)
	}
}

func TestRecordCoefficient_PassDoesNotWrite(t *testing.T) {
	dir := t.TempDir()
	writer := &session.InsightWriter{StateDir: dir, Logger: &domain.NopLogger{}}

	result := domain.JudgedData{
		PlanID:  "plan-pass-coeff",
		Verdict: domain.VerdictPass,
	}

	if err := writer.RecordCoefficient(result); err != nil {
		t.Fatalf("RecordCoefficient: %v", err)
	}

	path := filepath.Join(dir, "insights", "coefficient.md")
	if _, err := os.Stat(path); err == nil {
		t.Error("coefficient.md should not exist for pass verdict")
	}
}
