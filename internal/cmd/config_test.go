package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hironow/dominator/internal/cmd"
	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/session"
	"gopkg.in/yaml.v3"
)

func TestConfigCmd_ShowExists(t *testing.T) {
	// given
	rootCmd := cmd.NewRootCommand()

	// when
	var configCmd *testing.T
	_ = configCmd // suppress unused
	var found bool
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "config" {
			for _, child := range sub.Commands() {
				if child.Name() == "show" {
					found = true
					break
				}
			}
			break
		}
	}

	// then
	if !found {
		t.Fatal("config show subcommand not found")
	}
}

func TestConfigCmd_SetExists(t *testing.T) {
	// given
	rootCmd := cmd.NewRootCommand()

	// when
	var found bool
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "config" {
			for _, child := range sub.Commands() {
				if child.Name() == "set" {
					found = true
					break
				}
			}
			break
		}
	}

	// then
	if !found {
		t.Fatal("config set subcommand not found")
	}
}

func TestConfigCmd_SetAllKeys(t *testing.T) {
	tests := []struct {
		key      string
		value    string
		validate func(t *testing.T, cfg domain.Config)
	}{
		{
			key: "target.url", value: "https://example.com",
			validate: func(t *testing.T, cfg domain.Config) {
				t.Helper()
				if cfg.Target.URL != "https://example.com" {
					t.Errorf("target.url = %q, want %q", cfg.Target.URL, "https://example.com")
				}
			},
		},
		{
			key: "target.protocol", value: "json-rpc",
			validate: func(t *testing.T, cfg domain.Config) {
				t.Helper()
				if cfg.Target.Protocol != "json-rpc" {
					t.Errorf("target.protocol = %q, want %q", cfg.Target.Protocol, "json-rpc")
				}
			},
		},
		{
			key: "target.spec", value: "https://example.com/spec.json",
			validate: func(t *testing.T, cfg domain.Config) {
				t.Helper()
				if cfg.Target.Spec != "https://example.com/spec.json" {
					t.Errorf("target.spec = %q", cfg.Target.Spec)
				}
			},
		},
		{
			key: "target.docs", value: "https://docs.example.com",
			validate: func(t *testing.T, cfg domain.Config) {
				t.Helper()
				if cfg.Target.Docs != "https://docs.example.com" {
					t.Errorf("target.docs = %q", cfg.Target.Docs)
				}
			},
		},
		{
			key: "nfr.performance.p95_latency_ms", value: "200",
			validate: func(t *testing.T, cfg domain.Config) {
				t.Helper()
				if cfg.Nfr.Performance.P95LatencyMs != 200 {
					t.Errorf("p95_latency_ms = %d, want 200", cfg.Nfr.Performance.P95LatencyMs)
				}
			},
		},
		{
			key: "nfr.performance.error_rate_percent", value: "0.5",
			validate: func(t *testing.T, cfg domain.Config) {
				t.Helper()
				if cfg.Nfr.Performance.ErrorRatePercent != 0.5 {
					t.Errorf("error_rate_percent = %f, want 0.5", cfg.Nfr.Performance.ErrorRatePercent)
				}
			},
		},
		{
			key: "nfr.reliability.success_rate_percent", value: "99.9",
			validate: func(t *testing.T, cfg domain.Config) {
				t.Helper()
				if cfg.Nfr.Reliability.SuccessRatePercent != 99.9 {
					t.Errorf("success_rate_percent = %f, want 99.9", cfg.Nfr.Reliability.SuccessRatePercent)
				}
			},
		},
		{
			key: "nfr.scalability.target_rps", value: "500",
			validate: func(t *testing.T, cfg domain.Config) {
				t.Helper()
				if cfg.Nfr.Scalability.TargetRPS != 500 {
					t.Errorf("target_rps = %d, want 500", cfg.Nfr.Scalability.TargetRPS)
				}
			},
		},
		{
			key: "load.vus", value: "50",
			validate: func(t *testing.T, cfg domain.Config) {
				t.Helper()
				if cfg.Load.VUs != 50 {
					t.Errorf("vus = %d, want 50", cfg.Load.VUs)
				}
			},
		},
		{
			key: "load.duration", value: "60s",
			validate: func(t *testing.T, cfg domain.Config) {
				t.Helper()
				if cfg.Load.Duration != "60s" {
					t.Errorf("duration = %q, want %q", cfg.Load.Duration, "60s")
				}
			},
		},
		{
			key: "load.ramp_up", value: "10s",
			validate: func(t *testing.T, cfg domain.Config) {
				t.Helper()
				if cfg.Load.RampUp != "10s" {
					t.Errorf("ramp_up = %q, want %q", cfg.Load.RampUp, "10s")
				}
			},
		},
		{
			key: "approval.required", value: "false",
			validate: func(t *testing.T, cfg domain.Config) {
				t.Helper()
				if cfg.Approval.Required {
					t.Error("approval.required should be false")
				}
			},
		},
		{
			key: "lang", value: "en",
			validate: func(t *testing.T, cfg domain.Config) {
				t.Helper()
				if cfg.Lang != "en" {
					t.Errorf("lang = %q, want %q", cfg.Lang, "en")
				}
			},
		},
		{
			key: "claude_cmd", value: "/usr/local/bin/claude",
			validate: func(t *testing.T, cfg domain.Config) {
				t.Helper()
				if cfg.ClaudeCmd != "/usr/local/bin/claude" {
					t.Errorf("claude_cmd = %q", cfg.ClaudeCmd)
				}
			},
		},
		{
			key: "model", value: "sonnet",
			validate: func(t *testing.T, cfg domain.Config) {
				t.Helper()
				if cfg.Model != "sonnet" {
					t.Errorf("model = %q, want %q", cfg.Model, "sonnet")
				}
			},
		},
		{
			key: "timeout_sec", value: "3600",
			validate: func(t *testing.T, cfg domain.Config) {
				t.Helper()
				if cfg.TimeoutSec != 3600 {
					t.Errorf("timeout_sec = %d, want 3600", cfg.TimeoutSec)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			// given
			dir := t.TempDir()
			passDir := filepath.Join(dir, ".pass")
			if err := os.MkdirAll(passDir, 0o755); err != nil {
				t.Fatalf("mkdir: %v", err)
			}
			// Write default config
			if err := session.SaveConfig(filepath.Join(passDir, "config.yaml"), domain.DefaultConfig()); err != nil {
				t.Fatalf("write config: %v", err)
			}
			t.Chdir(dir)

			rootCmd := cmd.NewRootCommand()
			buf := new(bytes.Buffer)
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)
			rootCmd.SetArgs([]string{"config", "set", tt.key, tt.value})

			// when
			err := rootCmd.Execute()

			// then
			if err != nil {
				t.Fatalf("config set %s %s failed: %v", tt.key, tt.value, err)
			}

			// Verify the config was updated
			data, readErr := os.ReadFile(filepath.Join(passDir, "config.yaml"))
			if readErr != nil {
				t.Fatalf("read config: %v", readErr)
			}
			var cfg domain.Config
			if yamlErr := yaml.Unmarshal(data, &cfg); yamlErr != nil {
				t.Fatalf("unmarshal config: %v", yamlErr)
			}
			tt.validate(t, cfg)
		})
	}
}

func TestConfigCmd_RejectsUnknownKey(t *testing.T) {
	// given
	dir := t.TempDir()
	passDir := filepath.Join(dir, ".pass")
	if err := os.MkdirAll(passDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := session.SaveConfig(filepath.Join(passDir, "config.yaml"), domain.DefaultConfig()); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"config", "set", "nonexistent.key", "value"})

	// when
	err := rootCmd.Execute()

	// then
	if err == nil {
		t.Fatal("expected error for unknown key, got nil")
	}
	if !strings.Contains(err.Error(), "unknown config key") {
		t.Errorf("expected 'unknown config key' in error, got: %v", err)
	}
}
