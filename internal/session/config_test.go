package session_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/session"
)

func TestLoadConfig_DefaultValues(t *testing.T) {
	// given: no config file
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	// when
	cfg, err := session.LoadConfig(cfgPath)

	// then: should return defaults, no error
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Lang != "ja" {
		t.Errorf("Lang = %q, want %q", cfg.Lang, "ja")
	}
	if cfg.ClaudeCmd != domain.DefaultClaudeCmd {
		t.Errorf("ClaudeCmd = %q, want %q", cfg.ClaudeCmd, domain.DefaultClaudeCmd)
	}
	if cfg.Model != domain.DefaultModel {
		t.Errorf("Model = %q, want %q", cfg.Model, domain.DefaultModel)
	}
}

func TestSaveLoadConfig_RoundTrip(t *testing.T) {
	// given
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	cfg := domain.DefaultConfig()
	cfg.Target.URL = "http://localhost:9090"
	cfg.Target.Protocol = "json-rpc"
	cfg.Lang = "en"
	cfg.Load.VUs = 50
	cfg.Nfr.Performance.P95LatencyMs = 300

	// when
	if err := session.SaveConfig(cfgPath, cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	loaded, err := session.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	// then: all fields preserved
	if loaded.Target.URL != cfg.Target.URL {
		t.Errorf("Target.URL = %q, want %q", loaded.Target.URL, cfg.Target.URL)
	}
	if loaded.Target.Protocol != cfg.Target.Protocol {
		t.Errorf("Target.Protocol = %q, want %q", loaded.Target.Protocol, cfg.Target.Protocol)
	}
	if loaded.Lang != cfg.Lang {
		t.Errorf("Lang = %q, want %q", loaded.Lang, cfg.Lang)
	}
	if loaded.Load.VUs != cfg.Load.VUs {
		t.Errorf("Load.VUs = %d, want %d", loaded.Load.VUs, cfg.Load.VUs)
	}
	if loaded.Nfr.Performance.P95LatencyMs != cfg.Nfr.Performance.P95LatencyMs {
		t.Errorf("P95LatencyMs = %d, want %d", loaded.Nfr.Performance.P95LatencyMs, cfg.Nfr.Performance.P95LatencyMs)
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	os.WriteFile(cfgPath, []byte("invalid: [yaml: {broken"), 0o644)

	_, err := session.LoadConfig(cfgPath)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestSaveConfig_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	cfg := domain.DefaultConfig()
	if err := session.SaveConfig(cfgPath, cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	if _, err := os.Stat(cfgPath); err != nil {
		t.Errorf("config file should exist: %v", err)
	}
}

func TestFileConfigLoader_Load(t *testing.T) {
	dir := t.TempDir()
	// Write a config
	cfgPath := filepath.Join(dir, domain.ConfigFile)
	cfg := domain.DefaultConfig()
	cfg.Lang = "en"
	session.SaveConfig(cfgPath, cfg)

	loader := &session.FileConfigLoader{}
	loaded, err := loader.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Lang != "en" {
		t.Errorf("Lang = %q, want %q", loaded.Lang, "en")
	}
}

func TestFileConfigLoader_Load_NoFile(t *testing.T) {
	dir := t.TempDir()
	loader := &session.FileConfigLoader{}
	cfg, err := loader.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	// Should return default
	if cfg.Lang != "ja" {
		t.Errorf("Lang = %q, want %q (default)", cfg.Lang, "ja")
	}
}

func TestShowConfig_ReturnsYAML(t *testing.T) {
	dir := t.TempDir()
	cfg := domain.DefaultConfig()
	cfgPath := filepath.Join(dir, domain.ConfigFile)
	session.SaveConfig(cfgPath, cfg)

	data, err := session.ShowConfig(dir)
	if err != nil {
		t.Fatalf("ShowConfig: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty YAML output")
	}
}

func TestLoadConfig_PartialYAML(t *testing.T) {
	// given: YAML that only overrides some fields
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	os.WriteFile(cfgPath, []byte("lang: en\n"), 0o644)

	// when
	cfg, err := session.LoadConfig(cfgPath)

	// then: overridden field should match, rest should be defaults
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Lang != "en" {
		t.Errorf("Lang = %q, want %q", cfg.Lang, "en")
	}
	// Default values preserved for non-overridden fields
	if cfg.ClaudeCmd != domain.DefaultClaudeCmd {
		t.Errorf("ClaudeCmd = %q, want default %q", cfg.ClaudeCmd, domain.DefaultClaudeCmd)
	}
	if cfg.Load.VUs != 10 {
		t.Errorf("Load.VUs = %d, want default 10", cfg.Load.VUs)
	}
}

func TestSaveConfig_Overwrite(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	cfg1 := domain.DefaultConfig()
	cfg1.Lang = "ja"
	session.SaveConfig(cfgPath, cfg1)

	cfg2 := domain.DefaultConfig()
	cfg2.Lang = "en"
	session.SaveConfig(cfgPath, cfg2)

	loaded, err := session.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if loaded.Lang != "en" {
		t.Errorf("Lang = %q, want %q (overwritten)", loaded.Lang, "en")
	}
}

func TestValidateStateDir_Exists(t *testing.T) {
	dir := t.TempDir()
	passDir := filepath.Join(dir, domain.StateDir)
	os.MkdirAll(passDir, 0o755)

	err := session.ValidateStateDir(dir)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateStateDir_Missing(t *testing.T) {
	dir := t.TempDir()
	err := session.ValidateStateDir(dir)
	if err == nil {
		t.Fatal("expected error for missing state dir")
	}
}

func TestLoadConfig_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	os.WriteFile(cfgPath, []byte(""), 0o644)

	cfg, err := session.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	// Empty file unmarshal is a no-op, should return defaults
	if cfg.ClaudeCmd != domain.DefaultClaudeCmd {
		t.Errorf("ClaudeCmd = %q, want default", cfg.ClaudeCmd)
	}
}

func TestValidateStateDir_FileNotDir(t *testing.T) {
	dir := t.TempDir()
	// Create .pass as a file, not a directory
	passPath := filepath.Join(dir, domain.StateDir)
	os.WriteFile(passPath, []byte("not a dir"), 0o644)

	err := session.ValidateStateDir(dir)
	if err == nil {
		t.Fatal("expected error when .pass is a file, not directory")
	}
}

func TestShowConfig_NoFile(t *testing.T) {
	dir := t.TempDir()
	// No config file, should show defaults
	data, err := session.ShowConfig(dir)
	if err != nil {
		t.Fatalf("ShowConfig: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty default YAML output")
	}
}

func TestSaveLoadConfig_TargetDocs(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	cfg := domain.DefaultConfig()
	cfg.Target.Docs = "https://docs.example.com/api"
	session.SaveConfig(cfgPath, cfg)

	loaded, _ := session.LoadConfig(cfgPath)
	if loaded.Target.Docs != cfg.Target.Docs {
		t.Errorf("Target.Docs = %q, want %q", loaded.Target.Docs, cfg.Target.Docs)
	}
}

func TestSaveLoadConfig_LoadDuration(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	cfg := domain.DefaultConfig()
	cfg.Load.Duration = "120s"
	session.SaveConfig(cfgPath, cfg)

	loaded, _ := session.LoadConfig(cfgPath)
	if loaded.Load.Duration != "120s" {
		t.Errorf("Load.Duration = %q, want %q", loaded.Load.Duration, "120s")
	}
}

func TestSaveLoadConfig_ApprovalFalse(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	cfg := domain.DefaultConfig()
	cfg.Approval.Required = false
	session.SaveConfig(cfgPath, cfg)

	loaded, _ := session.LoadConfig(cfgPath)
	if loaded.Approval.Required {
		t.Error("Approval.Required should be false")
	}
}

func TestSaveLoadConfig_AllFields(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	cfg := domain.Config{
		Target: domain.TargetConfig{
			URL:      "http://api.example.com",
			Protocol: "openapi",
			Spec:     "https://api.example.com/spec.json",
			Docs:     "https://docs.example.com",
		},
		Nfr: domain.NfrConfig{
			Performance: domain.PerformanceNfr{P95LatencyMs: 300, ErrorRatePercent: 0.5},
			Reliability: domain.ReliabilityNfr{SuccessRatePercent: 99.9},
			Scalability: domain.ScalabilityNfr{TargetRPS: 500},
		},
		Load:       domain.LoadConfig{VUs: 50, Duration: "120s", RampUp: "15s"},
		Approval:   domain.ApprovalConfig{Required: false},
		Lang:       "en",
		ClaudeCmd:  "claude-dev",
		Model:      "sonnet",
		TimeoutSec: 3600,
	}

	session.SaveConfig(cfgPath, cfg)
	loaded, err := session.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if loaded.Target.URL != cfg.Target.URL {
		t.Errorf("Target.URL mismatch")
	}
	if loaded.Target.Spec != cfg.Target.Spec {
		t.Errorf("Target.Spec mismatch")
	}
	if loaded.Target.Docs != cfg.Target.Docs {
		t.Errorf("Target.Docs mismatch")
	}
	if loaded.Nfr.Performance.ErrorRatePercent != cfg.Nfr.Performance.ErrorRatePercent {
		t.Errorf("ErrorRatePercent mismatch")
	}
	if loaded.Nfr.Reliability.SuccessRatePercent != cfg.Nfr.Reliability.SuccessRatePercent {
		t.Errorf("SuccessRatePercent mismatch")
	}
	if loaded.Nfr.Scalability.TargetRPS != cfg.Nfr.Scalability.TargetRPS {
		t.Errorf("TargetRPS mismatch")
	}
	if loaded.Load.RampUp != cfg.Load.RampUp {
		t.Errorf("RampUp mismatch")
	}
	if loaded.Approval.Required != cfg.Approval.Required {
		t.Errorf("Approval.Required mismatch")
	}
	if loaded.ClaudeCmd != cfg.ClaudeCmd {
		t.Errorf("ClaudeCmd mismatch")
	}
	if loaded.Model != cfg.Model {
		t.Errorf("Model mismatch")
	}
	if loaded.TimeoutSec != cfg.TimeoutSec {
		t.Errorf("TimeoutSec mismatch")
	}
}
