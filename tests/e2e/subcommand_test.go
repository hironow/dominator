//go:build e2e

package e2e

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestE2E_Version(t *testing.T) {
	stdout, _, err := runCmd(t, t.TempDir(), "version")
	if err != nil {
		t.Fatalf("version: %v", err)
	}
	if !strings.Contains(stdout, "dominator") {
		t.Errorf("expected 'dominator' in version output, got: %s", stdout)
	}
}

func TestE2E_VersionJSON(t *testing.T) {
	stdout, _, err := runCmd(t, t.TempDir(), "version", "--json")
	if err != nil {
		t.Fatalf("version --json: %v", err)
	}
	var v map[string]string
	parseJSONOutput(t, stdout, &v)
	for _, key := range []string{"version", "commit", "date"} {
		if _, ok := v[key]; !ok {
			t.Errorf("missing key %q in JSON output", key)
		}
	}
}

func TestE2E_Help(t *testing.T) {
	stdout, _, err := runCmd(t, t.TempDir(), "--help")
	if err != nil {
		t.Fatalf("--help: %v", err)
	}
	for _, sub := range []string{"init", "check", "approve", "run", "generate", "doctor", "status", "validate", "version"} {
		if !strings.Contains(stdout, sub) {
			t.Errorf("expected %q in help output", sub)
		}
	}
}

func TestE2E_UnknownCommand(t *testing.T) {
	_, _, err := runCmd(t, t.TempDir(), "nonexistent-cmd")
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
}

func TestE2E_NoSubcommand(t *testing.T) {
	_, _, err := runCmd(t, t.TempDir())
	if err == nil {
		t.Fatal("expected error when no subcommand given")
	}
}

func TestE2E_Init(t *testing.T) {
	dir := initTestRepo(t)

	for _, sub := range []string{".run", "events", "outbox", "inbox", "archive", "k6-scripts"} {
		assertFileExists(t, dir+"/.pass/"+sub)
	}
	assertFileExists(t, dir+"/.pass/config.yaml")
	assertFileExists(t, dir+"/.pass/.gitignore")
}

func TestE2E_Init_AlreadyExists(t *testing.T) {
	dir := initTestRepo(t)
	_, stderr, err := runCmd(t, dir, "init")
	if err == nil {
		t.Fatal("expected error on second init")
	}
	if !strings.Contains(stderr, "already exists") {
		t.Errorf("expected 'already exists' error, got: %s", stderr)
	}
}

func TestE2E_Validate_ValidConfig(t *testing.T) {
	// validate requires real Claude + mcp-k6 (fake-claude does not implement
	// tool_result stream-json protocol). Covered by scenario tests instead.
	dir := initTestRepo(t)
	writeConfig(t, dir, defaultTestConfig())
	writeK6Script(t, dir, "load-test.js")

	// Verify validate runs without panic (exit code may be non-zero with fake-claude)
	_, stderr, _ := runCmd(t, dir, "validate")
	if stderr == "" {
		t.Error("expected validate output on stderr")
	}
}

func TestE2E_Validate_InvalidConfig(t *testing.T) {
	dir := initTestRepo(t)
	cfg := defaultTestConfig()
	nfr := cfg["nfr"].(map[string]any)
	perf := nfr["performance"].(map[string]any)
	perf["p95_latency_ms"] = -1 // invalid
	writeConfig(t, dir, cfg)

	_, _, err := runCmd(t, dir, "validate")
	if err == nil {
		t.Fatal("expected validation error for negative latency")
	}
}

func TestE2E_Doctor(t *testing.T) {
	dir := initTestRepo(t)
	writeConfig(t, dir, defaultTestConfig())

	_, stderr, err := runCmd(t, dir, "doctor")
	_ = err
	if stderr == "" {
		t.Error("expected doctor output on stderr")
	}
}

func TestE2E_DoctorJSON(t *testing.T) {
	dir := initTestRepo(t)
	writeConfig(t, dir, defaultTestConfig())

	stdout, _, _ := runCmd(t, dir, "doctor", "--json")
	var result struct {
		Checks []struct {
			Name    string `json:"name"`
			Status  string `json:"status"`
			Message string `json:"message"`
		} `json:"checks"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("parse doctor JSON: %v\nraw: %s", err, stdout)
	}
	if len(result.Checks) == 0 {
		t.Error("expected at least one check")
	}
}

func TestE2E_Status_Empty(t *testing.T) {
	dir := initTestRepo(t)
	writeConfig(t, dir, defaultTestConfig())

	stdout, _, err := runCmd(t, dir, "status")
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if !strings.Contains(stdout, "Plans:") {
		t.Errorf("expected 'Plans:' in status output, got: %s", stdout)
	}
}

func TestE2E_Status_JSON(t *testing.T) {
	dir := initTestRepo(t)
	writeConfig(t, dir, defaultTestConfig())

	stdout, _, err := runCmd(t, dir, "status", "--output", "json")
	if err != nil {
		t.Fatalf("status --json: %v", err)
	}
	var result map[string]any
	parseJSONOutput(t, stdout, &result)
	if _, ok := result["plan_count"]; !ok {
		t.Error("expected 'plan_count' in JSON output")
	}
}

func TestE2E_Clean(t *testing.T) {
	dir := initTestRepo(t)

	_, stderr, err := runCmd(t, dir, "clean", "--yes")
	if err != nil {
		t.Fatalf("clean --yes: %v\nstderr: %s", err, stderr)
	}
	if _, statErr := statFile(dir + "/.pass"); statErr == nil {
		t.Error("expected .pass/ to be removed after clean")
	}
}

func TestE2E_ArchivePrune_Empty(t *testing.T) {
	dir := initTestRepo(t)
	writeConfig(t, dir, defaultTestConfig())

	stdout, _, err := runCmd(t, dir, "archive-prune")
	if err != nil {
		t.Fatalf("archive-prune: %v", err)
	}
	// No archives to prune — should succeed with 0 pruned
	if !strings.Contains(stdout, "0") {
		t.Logf("archive-prune output: %s", stdout)
	}
}
