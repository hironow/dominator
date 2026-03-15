//go:build integration

package integration_test

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hironow/dominator/internal/cmd"
	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/session"
)

// TestConfig_SetShowRoundTrip verifies that config set followed by config show
// produces the expected value.
func TestConfig_SetShowRoundTrip(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// given: initialized directory
	dir := t.TempDir()
	passDir := filepath.Join(dir, ".pass")
	if err := session.InitPassDir(passDir, &domain.NopLogger{}); err != nil {
		t.Fatalf("InitPassDir: %v", err)
	}

	// when: set target.url
	setCmd := cmd.NewRootCommand()
	setCmd.SetArgs([]string{"config", "set", "target.url", "https://api.example.com", dir})
	setCmd.SetOut(&bytes.Buffer{})
	setCmd.SetErr(&bytes.Buffer{})
	if err := setCmd.Execute(); err != nil {
		t.Fatalf("config set: %v", err)
	}

	// when: show config
	stdout := &bytes.Buffer{}
	showCmd := cmd.NewRootCommand()
	showCmd.SetArgs([]string{"config", "show", dir})
	showCmd.SetOut(stdout)
	showCmd.SetErr(&bytes.Buffer{})
	if err := showCmd.Execute(); err != nil {
		t.Fatalf("config show: %v", err)
	}

	// then: output contains the set value
	output := stdout.String()
	if !strings.Contains(output, "https://api.example.com") {
		t.Errorf("expected output to contain 'https://api.example.com', got:\n%s", output)
	}
}

// TestConfig_AllKeys_Parameterized tests all supported config keys.
func TestConfig_AllKeys_Parameterized(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testCases := []struct {
		key   string
		value string
		want  string
	}{
		{"target.url", "https://test.example.com", "https://test.example.com"},
		{"target.protocol", "openapi", "openapi"},
		{"target.spec", "https://spec.example.com/openapi.json", "https://spec.example.com/openapi.json"},
		{"target.docs", "https://docs.example.com", "https://docs.example.com"},
		{"load.vus", "50", "50"},
		{"load.duration", "60s", "60s"},
		{"load.ramp_up", "10s", "10s"},
		{"nfr.performance.p95_latency_ms", "200", "200"},
		{"nfr.performance.error_rate_percent", "0.5", "0.5"},
		{"nfr.reliability.success_rate_percent", "99.9", "99.9"},
		{"nfr.scalability.target_rps", "500", "500"},
		{"approval.required", "false", "false"},
		{"lang", "en", "en"},
		{"claude_cmd", "/usr/local/bin/claude", "/usr/local/bin/claude"},
		{"model", "sonnet", "sonnet"},
		{"timeout_sec", "3600", "3600"},
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(t *testing.T) {
			// given
			dir := t.TempDir()
			passDir := filepath.Join(dir, ".pass")
			if err := session.InitPassDir(passDir, &domain.NopLogger{}); err != nil {
				t.Fatalf("InitPassDir: %v", err)
			}

			// when: set
			setCmd := cmd.NewRootCommand()
			setCmd.SetArgs([]string{"config", "set", tc.key, tc.value, dir})
			setCmd.SetOut(&bytes.Buffer{})
			setCmd.SetErr(&bytes.Buffer{})
			if err := setCmd.Execute(); err != nil {
				t.Fatalf("config set %s: %v", tc.key, err)
			}

			// when: show
			stdout := &bytes.Buffer{}
			showCmd := cmd.NewRootCommand()
			showCmd.SetArgs([]string{"config", "show", dir})
			showCmd.SetOut(stdout)
			showCmd.SetErr(&bytes.Buffer{})
			if err := showCmd.Execute(); err != nil {
				t.Fatalf("config show: %v", err)
			}

			// then
			if !strings.Contains(stdout.String(), tc.want) {
				t.Errorf("expected config output to contain %q for key %s, got:\n%s", tc.want, tc.key, stdout.String())
			}
		})
	}
}

// TestConfig_InvalidKey verifies that setting an unknown config key returns an error.
func TestConfig_InvalidKey(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// given
	dir := t.TempDir()
	passDir := filepath.Join(dir, ".pass")
	if err := session.InitPassDir(passDir, &domain.NopLogger{}); err != nil {
		t.Fatalf("InitPassDir: %v", err)
	}

	// when
	setCmd := cmd.NewRootCommand()
	setCmd.SetArgs([]string{"config", "set", "nonexistent.key", "value", dir})
	setCmd.SetOut(&bytes.Buffer{})
	setCmd.SetErr(&bytes.Buffer{})
	err := setCmd.Execute()

	// then
	if err == nil {
		t.Fatal("expected error for unknown config key")
	}
	if !strings.Contains(err.Error(), "unknown config key") {
		t.Errorf("expected 'unknown config key' error, got: %v", err)
	}
}

// TestConfig_InvalidValue verifies that setting a numeric key to a non-numeric
// value returns an error.
func TestConfig_InvalidValue(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testCases := []struct {
		name  string
		key   string
		value string
	}{
		{"vus_not_int", "load.vus", "not-a-number"},
		{"p95_not_int", "nfr.performance.p95_latency_ms", "abc"},
		{"error_rate_not_float", "nfr.performance.error_rate_percent", "xyz"},
		{"approval_not_bool", "approval.required", "maybe"},
		{"timeout_not_int", "timeout_sec", "forever"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			dir := t.TempDir()
			passDir := filepath.Join(dir, ".pass")
			if err := session.InitPassDir(passDir, &domain.NopLogger{}); err != nil {
				t.Fatalf("InitPassDir: %v", err)
			}

			// when
			setCmd := cmd.NewRootCommand()
			setCmd.SetArgs([]string{"config", "set", tc.key, tc.value, dir})
			setCmd.SetOut(&bytes.Buffer{})
			setCmd.SetErr(&bytes.Buffer{})
			err := setCmd.Execute()

			// then
			if err == nil {
				t.Errorf("expected error for invalid value %q on key %s", tc.value, tc.key)
			}
		})
	}
}

// TestConfig_ShowWithoutInit verifies config show fails gracefully.
func TestConfig_ShowWithoutInit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := t.TempDir()
	showCmd := cmd.NewRootCommand()
	showCmd.SetArgs([]string{"config", "show", dir})
	showCmd.SetOut(&bytes.Buffer{})
	showCmd.SetErr(&bytes.Buffer{})
	err := showCmd.Execute()

	if err == nil {
		t.Fatal("expected error when .pass/ does not exist")
	}
}

// TestConfig_PreservesDefaults verifies that setting one key preserves
// default values for other keys.
func TestConfig_PreservesDefaults(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// given
	dir := t.TempDir()
	passDir := filepath.Join(dir, ".pass")
	if err := session.InitPassDir(passDir, &domain.NopLogger{}); err != nil {
		t.Fatalf("InitPassDir: %v", err)
	}

	// when: set only target.url
	setCmd := cmd.NewRootCommand()
	setCmd.SetArgs([]string{"config", "set", "target.url", "https://changed.example.com", dir})
	setCmd.SetOut(&bytes.Buffer{})
	setCmd.SetErr(&bytes.Buffer{})
	if err := setCmd.Execute(); err != nil {
		t.Fatalf("config set: %v", err)
	}

	// then: verify defaults are still present
	cfgPath := filepath.Join(passDir, "config.yaml")
	cfg, err := session.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	defaults := domain.DefaultConfig()
	if cfg.Load.VUs != defaults.Load.VUs {
		t.Errorf("expected default VUs %d, got %d", defaults.Load.VUs, cfg.Load.VUs)
	}
	if cfg.Nfr.Performance.P95LatencyMs != defaults.Nfr.Performance.P95LatencyMs {
		t.Errorf("expected default p95_latency_ms %d, got %d",
			defaults.Nfr.Performance.P95LatencyMs, cfg.Nfr.Performance.P95LatencyMs)
	}
	if cfg.Target.URL != "https://changed.example.com" {
		t.Errorf("expected target.url to be changed, got %q", cfg.Target.URL)
	}
}
