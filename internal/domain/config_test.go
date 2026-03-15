package domain_test

import (
	"testing"

	"github.com/hironow/dominator/internal/domain"
	"gopkg.in/yaml.v3"
)

func yamlMarshal(v any) ([]byte, error) { return yaml.Marshal(v) }
func yamlUnmarshal(data []byte, v any) error { return yaml.Unmarshal(data, v) }

func TestDefaultConfig_HasReasonableDefaults(t *testing.T) {
	t.Parallel()

	cfg := domain.DefaultConfig()

	if cfg.Lang != "ja" {
		t.Errorf("default lang = %q, want %q", cfg.Lang, "ja")
	}
	if cfg.ClaudeCmd != domain.DefaultClaudeCmd {
		t.Errorf("default claude_cmd = %q, want %q", cfg.ClaudeCmd, domain.DefaultClaudeCmd)
	}
	if cfg.TimeoutSec <= 0 {
		t.Errorf("default timeout_sec = %d, want positive", cfg.TimeoutSec)
	}
	if cfg.Load.VUs <= 0 {
		t.Errorf("default vus = %d, want positive", cfg.Load.VUs)
	}
	if cfg.Load.Duration == "" {
		t.Error("default duration must not be empty")
	}
	if cfg.Nfr.Performance.P95LatencyMs <= 0 {
		t.Errorf("default p95_latency_ms = %d, want positive", cfg.Nfr.Performance.P95LatencyMs)
	}
	if cfg.Nfr.Reliability.SuccessRatePercent <= 0 {
		t.Errorf("default success_rate_percent = %f, want positive", cfg.Nfr.Reliability.SuccessRatePercent)
	}
}

func TestValidateConfig_ValidConfig(t *testing.T) {
	t.Parallel()

	cfg := domain.DefaultConfig()
	errs := domain.ValidateConfig(cfg)
	if len(errs) > 0 {
		t.Errorf("expected no errors for default config, got %v", errs)
	}
}

func TestValidateConfig_InvalidLang(t *testing.T) {
	t.Parallel()

	cfg := domain.DefaultConfig()
	cfg.Lang = "fr"
	errs := domain.ValidateConfig(cfg)
	if len(errs) == 0 {
		t.Error("expected error for invalid lang")
	}
	found := false
	for _, e := range errs {
		if e == "lang must be \"ja\" or \"en\" (got \"fr\")" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected lang error in %v", errs)
	}
}

func TestValidateConfig_NegativeVUs(t *testing.T) {
	t.Parallel()

	cfg := domain.DefaultConfig()
	cfg.Load.VUs = -1
	errs := domain.ValidateConfig(cfg)
	if len(errs) == 0 {
		t.Error("expected error for negative VUs")
	}
}

func TestValidateConfig_EmptyClaudeCmd(t *testing.T) {
	t.Parallel()

	cfg := domain.DefaultConfig()
	cfg.ClaudeCmd = ""
	errs := domain.ValidateConfig(cfg)
	if len(errs) == 0 {
		t.Error("expected error for empty claude_cmd")
	}
}

func TestValidateConfig_EmptyDuration(t *testing.T) {
	t.Parallel()

	cfg := domain.DefaultConfig()
	cfg.Load.Duration = ""
	errs := domain.ValidateConfig(cfg)
	if len(errs) == 0 {
		t.Error("expected error for empty duration")
	}
}

func TestValidateConfig_AllInvalidCombinations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		mutate   func(*domain.Config)
		wantErr  string
	}{
		{
			name:    "empty_model",
			mutate:  func(c *domain.Config) { c.Model = "" },
			wantErr: "model must not be empty",
		},
		{
			name:    "empty_claude_cmd",
			mutate:  func(c *domain.Config) { c.ClaudeCmd = "" },
			wantErr: "claude_cmd must not be empty",
		},
		{
			name:    "invalid_lang_fr",
			mutate:  func(c *domain.Config) { c.Lang = "fr" },
			wantErr: "lang must be",
		},
		{
			name:    "invalid_lang_empty",
			mutate:  func(c *domain.Config) { c.Lang = "" },
			wantErr: "lang must be",
		},
		{
			name:    "zero_vus",
			mutate:  func(c *domain.Config) { c.Load.VUs = 0 },
			wantErr: "load.vus must be positive",
		},
		{
			name:    "negative_vus",
			mutate:  func(c *domain.Config) { c.Load.VUs = -5 },
			wantErr: "load.vus must be positive",
		},
		{
			name:    "empty_duration",
			mutate:  func(c *domain.Config) { c.Load.Duration = "" },
			wantErr: "load.duration must not be empty",
		},
		{
			name:    "zero_p95_latency",
			mutate:  func(c *domain.Config) { c.Nfr.Performance.P95LatencyMs = 0 },
			wantErr: "nfr.performance.p95_latency_ms must be positive",
		},
		{
			name:    "negative_p95_latency",
			mutate:  func(c *domain.Config) { c.Nfr.Performance.P95LatencyMs = -100 },
			wantErr: "nfr.performance.p95_latency_ms must be positive",
		},
		{
			name:    "negative_error_rate",
			mutate:  func(c *domain.Config) { c.Nfr.Performance.ErrorRatePercent = -1.0 },
			wantErr: "nfr.performance.error_rate_percent must be non-negative",
		},
		{
			name:    "success_rate_over_100",
			mutate:  func(c *domain.Config) { c.Nfr.Reliability.SuccessRatePercent = 101.0 },
			wantErr: "nfr.reliability.success_rate_percent must be between 0 and 100",
		},
		{
			name:    "success_rate_negative",
			mutate:  func(c *domain.Config) { c.Nfr.Reliability.SuccessRatePercent = -1.0 },
			wantErr: "nfr.reliability.success_rate_percent must be between 0 and 100",
		},
		{
			name:    "negative_target_rps",
			mutate:  func(c *domain.Config) { c.Nfr.Scalability.TargetRPS = -10 },
			wantErr: "nfr.scalability.target_rps must be non-negative",
		},
		{
			name:    "negative_timeout",
			mutate:  func(c *domain.Config) { c.TimeoutSec = -1 },
			wantErr: "timeout_sec must be non-negative",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := domain.DefaultConfig()
			tt.mutate(&cfg)
			errs := domain.ValidateConfig(cfg)
			if len(errs) == 0 {
				t.Fatal("expected validation error, got none")
			}
			found := false
			for _, e := range errs {
				if contains(e, tt.wantErr) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected error containing %q, got %v", tt.wantErr, errs)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestValidateConfig_MultipleErrors(t *testing.T) {
	t.Parallel()

	// given
	cfg := domain.Config{} // zero-value config has many invalid fields

	// when
	errs := domain.ValidateConfig(cfg)

	// then: should have multiple validation errors
	if len(errs) < 3 {
		t.Errorf("expected at least 3 errors for zero-value config, got %d: %v", len(errs), errs)
	}
}

func TestDefaultConfig_AllFieldsHaveValues(t *testing.T) {
	t.Parallel()

	cfg := domain.DefaultConfig()

	// String fields should not be empty
	if cfg.Lang == "" {
		t.Error("Lang should not be empty")
	}
	if cfg.ClaudeCmd == "" {
		t.Error("ClaudeCmd should not be empty")
	}
	if cfg.Model == "" {
		t.Error("Model should not be empty")
	}

	// Numeric fields should be positive
	if cfg.TimeoutSec <= 0 {
		t.Errorf("TimeoutSec = %d, want positive", cfg.TimeoutSec)
	}
	if cfg.Load.VUs <= 0 {
		t.Errorf("Load.VUs = %d, want positive", cfg.Load.VUs)
	}
	if cfg.Load.Duration == "" {
		t.Error("Load.Duration should not be empty")
	}

	// NFR fields should have reasonable defaults
	if cfg.Nfr.Performance.P95LatencyMs <= 0 {
		t.Errorf("Nfr.Performance.P95LatencyMs = %d, want positive", cfg.Nfr.Performance.P95LatencyMs)
	}
	if cfg.Nfr.Performance.ErrorRatePercent <= 0 {
		t.Errorf("Nfr.Performance.ErrorRatePercent = %f, want positive", cfg.Nfr.Performance.ErrorRatePercent)
	}
	if cfg.Nfr.Reliability.SuccessRatePercent <= 0 {
		t.Errorf("Nfr.Reliability.SuccessRatePercent = %f, want positive", cfg.Nfr.Reliability.SuccessRatePercent)
	}
	if cfg.Nfr.Scalability.TargetRPS <= 0 {
		t.Errorf("Nfr.Scalability.TargetRPS = %d, want positive", cfg.Nfr.Scalability.TargetRPS)
	}
}

func TestConfig_YAMLRoundTrip(t *testing.T) {
	t.Parallel()

	// given
	cfg := domain.DefaultConfig()
	cfg.Target.URL = "http://localhost:8080"
	cfg.Target.Protocol = "openapi"
	cfg.Target.Spec = "https://example.com/spec.json"
	cfg.Target.Docs = "https://example.com/docs"

	// when: marshal then unmarshal
	data, err := yamlMarshal(cfg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var restored domain.Config
	if err := yamlUnmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// then: all fields should be preserved
	if restored.Lang != cfg.Lang {
		t.Errorf("Lang = %q, want %q", restored.Lang, cfg.Lang)
	}
	if restored.ClaudeCmd != cfg.ClaudeCmd {
		t.Errorf("ClaudeCmd = %q, want %q", restored.ClaudeCmd, cfg.ClaudeCmd)
	}
	if restored.Model != cfg.Model {
		t.Errorf("Model = %q, want %q", restored.Model, cfg.Model)
	}
	if restored.TimeoutSec != cfg.TimeoutSec {
		t.Errorf("TimeoutSec = %d, want %d", restored.TimeoutSec, cfg.TimeoutSec)
	}
	if restored.Target.URL != cfg.Target.URL {
		t.Errorf("Target.URL = %q, want %q", restored.Target.URL, cfg.Target.URL)
	}
	if restored.Target.Protocol != cfg.Target.Protocol {
		t.Errorf("Target.Protocol = %q, want %q", restored.Target.Protocol, cfg.Target.Protocol)
	}
	if restored.Load.VUs != cfg.Load.VUs {
		t.Errorf("Load.VUs = %d, want %d", restored.Load.VUs, cfg.Load.VUs)
	}
	if restored.Nfr.Performance.P95LatencyMs != cfg.Nfr.Performance.P95LatencyMs {
		t.Errorf("P95LatencyMs = %d, want %d", restored.Nfr.Performance.P95LatencyMs, cfg.Nfr.Performance.P95LatencyMs)
	}
	if restored.Approval.Required != cfg.Approval.Required {
		t.Errorf("Approval.Required = %v, want %v", restored.Approval.Required, cfg.Approval.Required)
	}
}

func TestValidateConfig_ZeroErrorRateIsValid(t *testing.T) {
	t.Parallel()

	cfg := domain.DefaultConfig()
	cfg.Nfr.Performance.ErrorRatePercent = 0.0
	errs := domain.ValidateConfig(cfg)
	if len(errs) != 0 {
		t.Errorf("zero error rate should be valid, got errors: %v", errs)
	}
}

func TestValidateConfig_ZeroTargetRPSIsValid(t *testing.T) {
	t.Parallel()

	cfg := domain.DefaultConfig()
	cfg.Nfr.Scalability.TargetRPS = 0
	errs := domain.ValidateConfig(cfg)
	if len(errs) != 0 {
		t.Errorf("zero target RPS should be valid, got errors: %v", errs)
	}
}

func TestValidateConfig_ZeroTimeoutIsValid(t *testing.T) {
	t.Parallel()

	cfg := domain.DefaultConfig()
	cfg.TimeoutSec = 0
	errs := domain.ValidateConfig(cfg)
	if len(errs) != 0 {
		t.Errorf("zero timeout should be valid, got errors: %v", errs)
	}
}

func TestValidateConfig_ValidLangs(t *testing.T) {
	t.Parallel()

	for _, lang := range []string{"ja", "en"} {
		t.Run(lang, func(t *testing.T) {
			cfg := domain.DefaultConfig()
			cfg.Lang = lang
			errs := domain.ValidateConfig(cfg)
			if len(errs) != 0 {
				t.Errorf("lang %q should be valid, got errors: %v", lang, errs)
			}
		})
	}
}

func TestDefaultConfig_ModelValue(t *testing.T) {
	t.Parallel()
	cfg := domain.DefaultConfig()
	if cfg.Model != domain.DefaultModel {
		t.Errorf("Model = %q, want %q", cfg.Model, domain.DefaultModel)
	}
}

func TestDefaultConfig_TimeoutValue(t *testing.T) {
	t.Parallel()
	cfg := domain.DefaultConfig()
	if cfg.TimeoutSec != domain.DefaultTimeoutSec {
		t.Errorf("TimeoutSec = %d, want %d", cfg.TimeoutSec, domain.DefaultTimeoutSec)
	}
}

func TestDefaultConfig_ApprovalRequired(t *testing.T) {
	t.Parallel()
	cfg := domain.DefaultConfig()
	if !cfg.Approval.Required {
		t.Error("Approval.Required should default to true")
	}
}

func TestDefaultConfig_LoadRampUp(t *testing.T) {
	t.Parallel()
	cfg := domain.DefaultConfig()
	if cfg.Load.RampUp == "" {
		t.Error("Load.RampUp should have a default value")
	}
}

func TestDefaultConfig_Validates(t *testing.T) {
	t.Parallel()
	cfg := domain.DefaultConfig()
	errs := domain.ValidateConfig(cfg)
	if len(errs) != 0 {
		t.Errorf("default config should validate, got: %v", errs)
	}
}

func TestValidateConfig_SuccessRateBoundary0(t *testing.T) {
	t.Parallel()
	cfg := domain.DefaultConfig()
	cfg.Nfr.Reliability.SuccessRatePercent = 0.0
	errs := domain.ValidateConfig(cfg)
	if len(errs) != 0 {
		t.Errorf("0 percent success rate should be valid, got: %v", errs)
	}
}

func TestValidateConfig_SuccessRateBoundary100(t *testing.T) {
	t.Parallel()
	cfg := domain.DefaultConfig()
	cfg.Nfr.Reliability.SuccessRatePercent = 100.0
	errs := domain.ValidateConfig(cfg)
	if len(errs) != 0 {
		t.Errorf("100 percent success rate should be valid, got: %v", errs)
	}
}
