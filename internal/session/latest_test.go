package session_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/session"
)

func TestWriteReadLatest_RoundTrip(t *testing.T) {
	// given
	dir := t.TempDir()
	judged := domain.JudgedData{
		PlanID:     "plan-latest",
		ScriptPath: "load-test.js",
		Duration:   "60s",
		VUs:        20,
		Results: domain.K6Results{
			P95LatencyMs:     250.5,
			ErrorRatePercent: 0.5,
			SuccessRate:      99.5,
			ActualRPS:        500,
		},
		Verdict: domain.VerdictPass,
	}

	// when
	err := session.WriteLatest(dir, judged)
	if err != nil {
		t.Fatalf("WriteLatest: %v", err)
	}
	loaded, err := session.ReadLatest(dir)
	if err != nil {
		t.Fatalf("ReadLatest: %v", err)
	}

	// then
	if loaded.PlanID != judged.PlanID {
		t.Errorf("PlanID = %q, want %q", loaded.PlanID, judged.PlanID)
	}
	if loaded.ScriptPath != judged.ScriptPath {
		t.Errorf("ScriptPath = %q, want %q", loaded.ScriptPath, judged.ScriptPath)
	}
	if loaded.Duration != judged.Duration {
		t.Errorf("Duration = %q, want %q", loaded.Duration, judged.Duration)
	}
	if loaded.VUs != judged.VUs {
		t.Errorf("VUs = %d, want %d", loaded.VUs, judged.VUs)
	}
	if loaded.Results.P95LatencyMs != judged.Results.P95LatencyMs {
		t.Errorf("P95LatencyMs = %f, want %f", loaded.Results.P95LatencyMs, judged.Results.P95LatencyMs)
	}
	if loaded.Results.ActualRPS != judged.Results.ActualRPS {
		t.Errorf("ActualRPS = %f, want %f", loaded.Results.ActualRPS, judged.Results.ActualRPS)
	}
	if loaded.Verdict != judged.Verdict {
		t.Errorf("Verdict = %q, want %q", loaded.Verdict, judged.Verdict)
	}
}

func TestReadLatest_FileNotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := session.ReadLatest(dir)
	if err == nil {
		t.Fatal("expected error for missing latest.json")
	}
}

func TestWriteLatest_CreatesDir(t *testing.T) {
	// given: dir without .run/ subdirectory
	dir := t.TempDir()
	runDir := filepath.Join(dir, ".run")

	// Verify .run/ does not exist
	if _, err := os.Stat(runDir); err == nil {
		t.Fatal(".run dir should not exist yet")
	}

	// when
	judged := domain.JudgedData{PlanID: "test", Verdict: domain.VerdictPass}
	err := session.WriteLatest(dir, judged)

	// then
	if err != nil {
		t.Fatalf("WriteLatest: %v", err)
	}

	// .run/ should now exist
	if _, err := os.Stat(runDir); err != nil {
		t.Errorf(".run dir should exist after WriteLatest: %v", err)
	}
}

func TestWriteLatest_WithDeviations(t *testing.T) {
	dir := t.TempDir()
	judged := domain.JudgedData{
		PlanID:  "plan-with-devs",
		Verdict: domain.VerdictViolation,
		Deviations: []domain.NfrDeviation{
			{Metric: "p95_latency_ms", Threshold: 500, Actual: 800, Deviation: 60, Severity: domain.SeverityHigh},
			{Metric: "error_rate_percent", Threshold: 1.0, Actual: 5.0, Deviation: 400, Severity: domain.SeverityHigh},
		},
	}

	err := session.WriteLatest(dir, judged)
	if err != nil {
		t.Fatalf("WriteLatest: %v", err)
	}

	loaded, err := session.ReadLatest(dir)
	if err != nil {
		t.Fatalf("ReadLatest: %v", err)
	}

	if len(loaded.Deviations) != 2 {
		t.Fatalf("expected 2 deviations, got %d", len(loaded.Deviations))
	}
	if loaded.Deviations[0].Metric != "p95_latency_ms" {
		t.Errorf("Deviation[0].Metric = %q, want %q", loaded.Deviations[0].Metric, "p95_latency_ms")
	}
}

func TestReadLatest_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	runDir := filepath.Join(dir, ".run")
	os.MkdirAll(runDir, 0o755)
	os.WriteFile(filepath.Join(runDir, "latest.json"), []byte("not json"), 0o644)

	_, err := session.ReadLatest(dir)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestWriteLatest_EmptyResult(t *testing.T) {
	dir := t.TempDir()
	judged := domain.JudgedData{}

	err := session.WriteLatest(dir, judged)
	if err != nil {
		t.Fatalf("WriteLatest: %v", err)
	}

	loaded, err := session.ReadLatest(dir)
	if err != nil {
		t.Fatalf("ReadLatest: %v", err)
	}
	if loaded.PlanID != "" {
		t.Errorf("PlanID = %q, want empty", loaded.PlanID)
	}
}

func TestWriteLatest_Overwrite(t *testing.T) {
	dir := t.TempDir()

	// Write first
	first := domain.JudgedData{PlanID: "first", Verdict: domain.VerdictPass}
	if err := session.WriteLatest(dir, first); err != nil {
		t.Fatalf("WriteLatest first: %v", err)
	}

	// Write second (overwrites)
	second := domain.JudgedData{PlanID: "second", Verdict: domain.VerdictViolation}
	if err := session.WriteLatest(dir, second); err != nil {
		t.Fatalf("WriteLatest second: %v", err)
	}

	// Read should return second
	loaded, err := session.ReadLatest(dir)
	if err != nil {
		t.Fatalf("ReadLatest: %v", err)
	}
	if loaded.PlanID != "second" {
		t.Errorf("PlanID = %q, want %q", loaded.PlanID, "second")
	}
	if loaded.Verdict != domain.VerdictViolation {
		t.Errorf("Verdict = %q, want %q", loaded.Verdict, domain.VerdictViolation)
	}
}

func TestWriteLatest_ResultsPreserved(t *testing.T) {
	dir := t.TempDir()
	judged := domain.JudgedData{
		PlanID: "results-test",
		Results: domain.K6Results{
			P95LatencyMs:     99.99,
			ErrorRatePercent: 0.01,
			SuccessRate:      99.99,
			ActualRPS:        9999.9,
		},
		Verdict: domain.VerdictPass,
	}

	session.WriteLatest(dir, judged)
	loaded, _ := session.ReadLatest(dir)

	if loaded.Results.ErrorRatePercent != 0.01 {
		t.Errorf("ErrorRatePercent = %f, want 0.01", loaded.Results.ErrorRatePercent)
	}
	if loaded.Results.SuccessRate != 99.99 {
		t.Errorf("SuccessRate = %f, want 99.99", loaded.Results.SuccessRate)
	}
}

func TestWriteLatest_ScriptPathPreserved(t *testing.T) {
	dir := t.TempDir()
	judged := domain.JudgedData{
		PlanID:     "path-test",
		ScriptPath: "deep/nested/load-test.js",
		Verdict:    domain.VerdictPass,
	}

	session.WriteLatest(dir, judged)
	loaded, _ := session.ReadLatest(dir)

	if loaded.ScriptPath != "deep/nested/load-test.js" {
		t.Errorf("ScriptPath = %q, want %q", loaded.ScriptPath, "deep/nested/load-test.js")
	}
}

func TestWriteLatest_VUsPreserved(t *testing.T) {
	dir := t.TempDir()
	judged := domain.JudgedData{
		PlanID:  "vus-test",
		VUs:     100,
		Verdict: domain.VerdictPass,
	}

	session.WriteLatest(dir, judged)
	loaded, _ := session.ReadLatest(dir)

	if loaded.VUs != 100 {
		t.Errorf("VUs = %d, want 100", loaded.VUs)
	}
}

func TestWriteLatest_DurationPreserved(t *testing.T) {
	dir := t.TempDir()
	judged := domain.JudgedData{
		PlanID:   "dur-test",
		Duration: "120s",
		Verdict:  domain.VerdictPass,
	}

	session.WriteLatest(dir, judged)
	loaded, _ := session.ReadLatest(dir)

	if loaded.Duration != "120s" {
		t.Errorf("Duration = %q, want %q", loaded.Duration, "120s")
	}
}
