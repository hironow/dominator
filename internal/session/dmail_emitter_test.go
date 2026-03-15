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

	// Verify the 3 kinds
	kinds := map[string]bool{"design-feedback": false, "implementation-feedback": false, "verification-feedback": false}
	for _, e := range entries {
		for k := range kinds {
			if strings.HasPrefix(e.Name(), k) {
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
