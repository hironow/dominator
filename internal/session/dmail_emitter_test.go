package session_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/session"
)

func TestEmitViolation_Creates3Files(t *testing.T) {
	// given
	dir := t.TempDir()
	emitter := &session.DMailEmitter{StateDir: dir, Logger: &domain.NopLogger{}}
	result := domain.JudgedData{
		PlanID:     "plan-001",
		ScriptPath: "test.js",
		Duration:   "30s",
		VUs:        10,
		Verdict:    domain.VerdictViolation,
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

	// when
	err := emitter.EmitViolation(result)

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outboxDir := filepath.Join(dir, "outbox")
	entries, readErr := os.ReadDir(outboxDir)
	if readErr != nil {
		t.Fatalf("failed to read outbox dir: %v", readErr)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 D-Mail files, got %d", len(entries))
	}

	// Verify the 3 canonical kinds. Per Rival Contract v1 Phase 4, the third
	// emission is "report" (for amadeus context) instead of the non-canonical
	// "verification-feedback".
	kinds := map[string]bool{"design-feedback": false, "implementation-feedback": false, "report": false}
	for _, e := range entries {
		for k := range kinds {
			if strings.HasPrefix(e.Name(), k+"-") {
				kinds[k] = true
			}
		}
	}
	for k, found := range kinds {
		if !found {
			t.Errorf("missing D-Mail kind: %s", k)
		}
	}
}

// TestEmitViolation_UsesOnlyCanonicalDMailKinds enforces the Rival Contract
// v1 Phase 4 invariant: every emitted file must use a canonical D-Mail kind
// (design-feedback, implementation-feedback, or report). No file may use a
// non-canonical kind such as verification-feedback.
func TestEmitViolation_UsesOnlyCanonicalDMailKinds(t *testing.T) {
	// given
	dir := t.TempDir()
	emitter := &session.DMailEmitter{StateDir: dir, Logger: &domain.NopLogger{}}
	result := domain.JudgedData{
		PlanID:     "plan-canonical",
		ScriptPath: "test.js",
		Duration:   "30s",
		VUs:        10,
		Verdict:    domain.VerdictViolation,
		Deviations: []domain.NfrDeviation{
			{Metric: "p95_latency_ms", Threshold: 500, Actual: 750, Deviation: 50, Severity: domain.SeverityHigh},
		},
	}

	// when
	if err := emitter.EmitViolation(result); err != nil {
		t.Fatalf("EmitViolation: %v", err)
	}

	// then
	outboxDir := filepath.Join(dir, "outbox")
	entries, _ := os.ReadDir(outboxDir)
	canonical := map[string]struct{}{
		"design-feedback":         {},
		"implementation-feedback": {},
		"report":                  {},
	}
	for _, e := range entries {
		data, _ := os.ReadFile(filepath.Join(outboxDir, e.Name()))
		fm := string(data)
		var kind string
		for _, line := range strings.Split(fm, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "kind:") {
				kind = strings.TrimSpace(strings.TrimPrefix(line, "kind:"))
				kind = strings.Trim(kind, `"'`)
				break
			}
		}
		if _, ok := canonical[kind]; !ok {
			t.Errorf("file %s: non-canonical kind %q (allowed: design-feedback, implementation-feedback, report)", e.Name(), kind)
		}
	}
}

// TestEmitViolation_DoesNotEmitVerificationFeedback proves that the
// non-canonical "verification-feedback" kind is no longer emitted by any
// Rival Contract v1 path.
func TestEmitViolation_DoesNotEmitVerificationFeedback(t *testing.T) {
	// given
	dir := t.TempDir()
	emitter := &session.DMailEmitter{StateDir: dir, Logger: &domain.NopLogger{}}
	result := domain.JudgedData{
		PlanID:     "plan-no-verify",
		ScriptPath: "test.js",
		Duration:   "30s",
		VUs:        10,
		Verdict:    domain.VerdictViolation,
		Deviations: []domain.NfrDeviation{
			{Metric: "p95_latency_ms", Threshold: 500, Actual: 750, Deviation: 50, Severity: domain.SeverityMedium},
		},
	}

	// when
	if err := emitter.EmitViolation(result); err != nil {
		t.Fatalf("EmitViolation: %v", err)
	}

	// then
	outboxDir := filepath.Join(dir, "outbox")
	entries, _ := os.ReadDir(outboxDir)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "verification-feedback") {
			t.Errorf("file %s: verification-feedback prefix is forbidden in Rival Contract v1", e.Name())
		}
		data, _ := os.ReadFile(filepath.Join(outboxDir, e.Name()))
		if strings.Contains(string(data), "kind: verification-feedback") {
			t.Errorf("file %s: contains forbidden kind 'verification-feedback' in frontmatter", e.Name())
		}
	}
}

func TestEmitPass_Creates1File(t *testing.T) {
	// given
	dir := t.TempDir()
	emitter := &session.DMailEmitter{StateDir: dir, Logger: &domain.NopLogger{}}
	result := domain.JudgedData{
		PlanID:     "plan-001",
		ScriptPath: "test.js",
		Duration:   "30s",
		VUs:        10,
		Verdict:    domain.VerdictPass,
	}

	// when
	err := emitter.EmitPass(result)

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outboxDir := filepath.Join(dir, "outbox")
	entries, readErr := os.ReadDir(outboxDir)
	if readErr != nil {
		t.Fatalf("failed to read outbox dir: %v", readErr)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 D-Mail file, got %d", len(entries))
	}
	if !strings.HasPrefix(entries[0].Name(), "nfr-pass-") {
		t.Errorf("expected file to start with 'nfr-pass-', got: %s", entries[0].Name())
	}
}

func TestDMailFormat_HasYAMLFrontmatter(t *testing.T) {
	// given
	dir := t.TempDir()
	emitter := &session.DMailEmitter{StateDir: dir, Logger: &domain.NopLogger{}}
	result := domain.JudgedData{
		PlanID:     "plan-001",
		ScriptPath: "test.js",
		Duration:   "30s",
		VUs:        10,
		Verdict:    domain.VerdictViolation,
		Deviations: []domain.NfrDeviation{
			{
				Metric:    "p95_latency_ms",
				Threshold: 500,
				Actual:    750,
				Deviation: 50,
				Severity:  domain.SeverityHigh,
			},
		},
	}

	// when
	err := emitter.EmitViolation(result)

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outboxDir := filepath.Join(dir, "outbox")
	entries, _ := os.ReadDir(outboxDir)
	for _, e := range entries {
		data, readErr := os.ReadFile(filepath.Join(outboxDir, e.Name()))
		if readErr != nil {
			t.Fatalf("read file %s: %v", e.Name(), readErr)
		}
		content := string(data)

		// Check YAML frontmatter delimiters
		if !strings.HasPrefix(content, "---\n") {
			t.Errorf("file %s: missing YAML frontmatter start delimiter", e.Name())
		}
		if !strings.Contains(content, "\n---\n") {
			t.Errorf("file %s: missing YAML frontmatter end delimiter", e.Name())
		}

		// Check required frontmatter fields
		for _, field := range []string{"name:", "kind:", "description:", "severity:"} {
			if !strings.Contains(content, field) {
				t.Errorf("file %s: missing frontmatter field '%s'", e.Name(), field)
			}
		}

		// Check severity is the max (high)
		if !strings.Contains(content, "severity: high") {
			t.Errorf("file %s: expected severity 'high', content: %s", e.Name(), content)
		}
	}
}

func TestEmitViolation_SeverityLevels(t *testing.T) {
	tests := []struct {
		name         string
		severity     string
		wantSeverity string
	}{
		{name: "low_severity", severity: domain.SeverityLow, wantSeverity: "low"},
		{name: "medium_severity", severity: domain.SeverityMedium, wantSeverity: "medium"},
		{name: "high_severity", severity: domain.SeverityHigh, wantSeverity: "high"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			emitter := &session.DMailEmitter{StateDir: dir, Logger: &domain.NopLogger{}}
			result := domain.JudgedData{
				PlanID:     "plan-sev",
				ScriptPath: "test.js",
				Duration:   "30s",
				VUs:        10,
				Verdict:    domain.VerdictViolation,
				Deviations: []domain.NfrDeviation{
					{Metric: "p95_latency_ms", Threshold: 500, Actual: 750, Deviation: 50, Severity: tt.severity},
				},
			}

			err := emitter.EmitViolation(result)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			outboxDir := filepath.Join(dir, "outbox")
			entries, _ := os.ReadDir(outboxDir)
			for _, e := range entries {
				data, _ := os.ReadFile(filepath.Join(outboxDir, e.Name()))
				if !strings.Contains(string(data), "severity: "+tt.wantSeverity) {
					t.Errorf("file %s: expected severity %q in content", e.Name(), tt.wantSeverity)
				}
			}
		})
	}
}

func TestEmitViolation_MultipleDeviations(t *testing.T) {
	dir := t.TempDir()
	emitter := &session.DMailEmitter{StateDir: dir, Logger: &domain.NopLogger{}}
	result := domain.JudgedData{
		PlanID:     "plan-multi",
		ScriptPath: "test.js",
		Duration:   "60s",
		VUs:        20,
		Verdict:    domain.VerdictViolation,
		Deviations: []domain.NfrDeviation{
			{Metric: "p95_latency_ms", Threshold: 500, Actual: 800, Deviation: 60, Severity: domain.SeverityHigh},
			{Metric: "error_rate_percent", Threshold: 1.0, Actual: 5.0, Deviation: 400, Severity: domain.SeverityHigh},
			{Metric: "success_rate_percent", Threshold: 99.0, Actual: 95.0, Deviation: 4.04, Severity: domain.SeverityLow},
		},
	}

	err := emitter.EmitViolation(result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify all metrics appear in the table
	outboxDir := filepath.Join(dir, "outbox")
	entries, _ := os.ReadDir(outboxDir)
	if len(entries) != 3 {
		t.Fatalf("expected 3 files, got %d", len(entries))
	}

	// Check first file content
	data, _ := os.ReadFile(filepath.Join(outboxDir, entries[0].Name()))
	content := string(data)
	for _, metric := range []string{"p95_latency_ms", "error_rate_percent", "success_rate_percent"} {
		if !strings.Contains(content, metric) {
			t.Errorf("expected %q in table, not found", metric)
		}
	}

	// 3 deviations should be mentioned in description
	if !strings.Contains(content, "3 deviation(s)") {
		t.Errorf("expected '3 deviation(s)' in content")
	}
}

func TestEmitViolation_FileNaming(t *testing.T) {
	dir := t.TempDir()
	emitter := &session.DMailEmitter{StateDir: dir, Logger: &domain.NopLogger{}}
	result := domain.JudgedData{
		PlanID:     "plan-naming",
		ScriptPath: "test.js",
		Duration:   "30s",
		VUs:        10,
		Verdict:    domain.VerdictViolation,
		Deviations: []domain.NfrDeviation{
			{Metric: "p95_latency_ms", Threshold: 500, Actual: 750, Deviation: 50, Severity: domain.SeverityMedium},
		},
	}

	err := emitter.EmitViolation(result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outboxDir := filepath.Join(dir, "outbox")
	entries, _ := os.ReadDir(outboxDir)
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".md") {
			t.Errorf("file %s should end with .md", name)
		}
		// Should contain timestamp pattern YYYYMMDDTHHMMSSZ
		if !strings.Contains(name, "T") || !strings.Contains(name, "Z") {
			t.Errorf("file %s should contain timestamp pattern", name)
		}
	}
}

func TestEmitPass_Content(t *testing.T) {
	dir := t.TempDir()
	emitter := &session.DMailEmitter{StateDir: dir, Logger: &domain.NopLogger{}}
	result := domain.JudgedData{
		PlanID:     "plan-pass-content",
		ScriptPath: "load-test.js",
		Duration:   "60s",
		VUs:        50,
		Verdict:    domain.VerdictPass,
	}

	err := emitter.EmitPass(result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outboxDir := filepath.Join(dir, "outbox")
	entries, _ := os.ReadDir(outboxDir)
	if len(entries) != 1 {
		t.Fatalf("expected 1 file, got %d", len(entries))
	}

	data, _ := os.ReadFile(filepath.Join(outboxDir, entries[0].Name()))
	content := string(data)

	// Check frontmatter
	if !strings.Contains(content, "kind: report") {
		t.Error("expected kind: report in pass D-Mail")
	}
	if !strings.Contains(content, "severity: low") {
		t.Error("expected severity: low in pass D-Mail")
	}
	if !strings.Contains(content, "All NFR thresholds met") {
		t.Error("expected 'All NFR thresholds met' in description")
	}

	// Check body
	if !strings.Contains(content, "plan-pass-content") {
		t.Error("expected plan ID in body")
	}
	if !strings.Contains(content, "load-test.js") {
		t.Error("expected script path in body")
	}
	if !strings.Contains(content, "VUs: 50") {
		t.Error("expected VUs in body")
	}
	if !strings.Contains(content, "Verdict: pass") {
		t.Error("expected verdict in body")
	}
}

func TestEmitPass_FileNaming(t *testing.T) {
	dir := t.TempDir()
	emitter := &session.DMailEmitter{StateDir: dir, Logger: &domain.NopLogger{}}
	result := domain.JudgedData{
		PlanID:  "plan-pass-naming",
		Verdict: domain.VerdictPass,
	}

	err := emitter.EmitPass(result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outboxDir := filepath.Join(dir, "outbox")
	entries, _ := os.ReadDir(outboxDir)
	if len(entries) != 1 {
		t.Fatalf("expected 1 file, got %d", len(entries))
	}
	name := entries[0].Name()
	if !strings.HasPrefix(name, "nfr-pass-") {
		t.Errorf("expected prefix 'nfr-pass-', got %q", name)
	}
	if !strings.HasSuffix(name, ".md") {
		t.Errorf("expected suffix '.md', got %q", name)
	}
}

func TestEmitViolation_OutboxDirCreated(t *testing.T) {
	dir := t.TempDir()
	emitter := &session.DMailEmitter{StateDir: dir, Logger: &domain.NopLogger{}}
	result := domain.JudgedData{
		PlanID:  "plan-dir",
		Verdict: domain.VerdictViolation,
		Deviations: []domain.NfrDeviation{
			{Metric: "m1", Severity: domain.SeverityLow},
		},
	}

	// Outbox dir should not exist yet
	outboxDir := filepath.Join(dir, "outbox")
	if _, err := os.Stat(outboxDir); err == nil {
		t.Fatal("outbox dir should not exist yet")
	}

	err := emitter.EmitViolation(result)
	if err != nil {
		t.Fatalf("EmitViolation: %v", err)
	}

	if _, err := os.Stat(outboxDir); err != nil {
		t.Errorf("outbox dir should exist: %v", err)
	}
}

func TestEmitViolation_HighestSeverityWins(t *testing.T) {
	dir := t.TempDir()
	emitter := &session.DMailEmitter{StateDir: dir, Logger: &domain.NopLogger{}}
	result := domain.JudgedData{
		PlanID:  "plan-max-sev",
		Verdict: domain.VerdictViolation,
		Deviations: []domain.NfrDeviation{
			{Metric: "m1", Severity: domain.SeverityLow},
			{Metric: "m2", Severity: domain.SeverityHigh},
			{Metric: "m3", Severity: domain.SeverityMedium},
		},
	}

	err := emitter.EmitViolation(result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outboxDir := filepath.Join(dir, "outbox")
	entries, _ := os.ReadDir(outboxDir)
	data, _ := os.ReadFile(filepath.Join(outboxDir, entries[0].Name()))
	if !strings.Contains(string(data), "severity: high") {
		t.Error("expected max severity 'high' in frontmatter")
	}
}
