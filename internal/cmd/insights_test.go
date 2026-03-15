package cmd_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hironow/dominator/internal/cmd"
)

func TestInsightsCmd_SubcommandExists(t *testing.T) {
	// given
	rootCmd := cmd.NewRootCommand()

	// when
	var found bool
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "insights" {
			found = true
			break
		}
	}

	// then
	if !found {
		t.Fatal("insights subcommand not found")
	}
}

func TestInsightsCmd_FailsWithoutInit(t *testing.T) {
	// given: directory without .pass/
	dir := t.TempDir()
	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"insights"})

	// when
	err := rootCmd.Execute()

	// then
	if err == nil {
		t.Fatal("expected error when .pass/ not initialized, got nil")
	}
	if !strings.Contains(err.Error(), "dominator init") {
		t.Errorf("expected error to mention 'dominator init', got: %v", err)
	}
}

func TestInsightsCmd_EmptyInsights(t *testing.T) {
	// given: directory with .pass/insights/ but no files
	dir := t.TempDir()
	insightsDir := filepath.Join(dir, ".pass", "insights")
	if err := os.MkdirAll(insightsDir, 0o755); err != nil {
		t.Fatalf("create insights dir: %v", err)
	}
	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetArgs([]string{"insights"})

	// when
	err := rootCmd.Execute()

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Hue         []any `json:"hue"`
		Coefficient []any `json:"coefficient"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON output: %v\nraw: %s", err, stdout.String())
	}
	if len(result.Hue) != 0 {
		t.Errorf("expected empty hue array, got %d entries", len(result.Hue))
	}
	if len(result.Coefficient) != 0 {
		t.Errorf("expected empty coefficient array, got %d entries", len(result.Coefficient))
	}
}

func TestInsightsCmd_WithHueData(t *testing.T) {
	// given: directory with .pass/insights/hue.md containing data
	dir := t.TempDir()
	insightsDir := filepath.Join(dir, ".pass", "insights")
	if err := os.MkdirAll(insightsDir, 0o755); err != nil {
		t.Fatalf("create insights dir: %v", err)
	}
	hueContent := `
## 2026-01-01T00:00:00Z — pass

Script: load.js, VUs: 10, Duration: 30s
Deviations: 0
`
	if err := os.WriteFile(filepath.Join(insightsDir, "hue.md"), []byte(hueContent), 0o644); err != nil {
		t.Fatalf("write hue.md: %v", err)
	}
	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetArgs([]string{"insights"})

	// when
	err := rootCmd.Execute()

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Hue []struct {
			Timestamp string `json:"timestamp"`
			Verdict   string `json:"verdict"`
		} `json:"hue"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(result.Hue) != 1 {
		t.Fatalf("expected 1 hue entry, got %d", len(result.Hue))
	}
	if result.Hue[0].Verdict != "pass" {
		t.Errorf("expected verdict 'pass', got %q", result.Hue[0].Verdict)
	}
}

func TestInsightsCmd_OutputJSON(t *testing.T) {
	// given: directory with .pass/insights/ and hue.md
	dir := t.TempDir()
	insightsDir := filepath.Join(dir, ".pass", "insights")
	os.MkdirAll(insightsDir, 0o755)
	hueContent := `
## 2026-01-01T00:00:00Z — pass

Script: test.js, VUs: 10, Duration: 30s
Deviations: 0

## 2026-01-02T00:00:00Z — violation

Script: stress.js, VUs: 50, Duration: 60s
Deviations: 2
`
	os.WriteFile(filepath.Join(insightsDir, "hue.md"), []byte(hueContent), 0o644)
	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetArgs([]string{"insights"})

	// when
	err := rootCmd.Execute()

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !json.Valid(stdout.Bytes()) {
		t.Errorf("stdout is not valid JSON: %s", stdout.String())
	}
}

func TestInsightsCmd_WithCoefficientData(t *testing.T) {
	dir := t.TempDir()
	insightsDir := filepath.Join(dir, ".pass", "insights")
	os.MkdirAll(insightsDir, 0o755)

	coeffContent := `
## 2026-01-01T00:00:00Z — Violation

| Metric | Threshold | Actual | Deviation | Severity |
|--------|-----------|--------|-----------|----------|
| p95_latency_ms | 500.00 | 750.00 | 50.0% | medium |
`
	os.WriteFile(filepath.Join(insightsDir, "coefficient.md"), []byte(coeffContent), 0o644)
	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetArgs([]string{"insights"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Coefficient []any `json:"coefficient"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(result.Coefficient) != 1 {
		t.Errorf("expected 1 coefficient entry, got %d", len(result.Coefficient))
	}
}
