package cmd_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hironow/dominator/internal/cmd"
	"github.com/hironow/dominator/internal/domain"
)

func TestCheckCmd_SubcommandExists(t *testing.T) {
	// given
	rootCmd := cmd.NewRootCommand()

	// when
	var found bool
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "check" {
			found = true
			break
		}
	}

	// then
	if !found {
		t.Fatal("check subcommand not found")
	}
}

func TestCheckCmd_FailsWithoutInit(t *testing.T) {
	// given: directory without .pass/
	dir := t.TempDir()
	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"check"})

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

func TestCheckCmd_SucceedsWithScript(t *testing.T) {
	// given: initialized .pass/ with a k6 script and config
	dir := t.TempDir()
	passDir := filepath.Join(dir, ".pass")
	for _, sub := range []string{".run", "events", "k6-scripts"} {
		if err := os.MkdirAll(filepath.Join(passDir, sub), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
	}
	// Write a k6 script
	scriptPath := filepath.Join(passDir, "k6-scripts", "load-test.js")
	if err := os.WriteFile(scriptPath, []byte("// k6 script"), 0o644); err != nil {
		t.Fatalf("write script: %v", err)
	}
	// Write config
	cfgContent := `lang: en
claude_cmd: claude
model: opus
timeout_sec: 60
load:
  vus: 10
  duration: "30s"
  ramp_up: "5s"
nfr:
  performance:
    p95_latency_ms: 500
    error_rate_percent: 1.0
  reliability:
    success_rate_percent: 99.0
  scalability:
    target_rps: 100
approval:
  required: true
`
	if err := os.WriteFile(filepath.Join(passDir, "config.yaml"), []byte(cfgContent), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetArgs([]string{"check"})

	// when
	err := rootCmd.Execute()

	// then
	if err != nil {
		t.Fatalf("check failed: %v", err)
	}

	// stdout should contain valid plan JSON
	var plan domain.Plan
	if jsonErr := json.Unmarshal(stdout.Bytes(), &plan); jsonErr != nil {
		t.Fatalf("stdout is not valid plan JSON: %v\noutput: %s", jsonErr, stdout.String())
	}
	if plan.Script != "load-test.js" {
		t.Errorf("plan script = %q, want %q", plan.Script, "load-test.js")
	}
	if plan.Approved {
		t.Error("plan should not be approved yet")
	}
}
