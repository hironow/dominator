package domain_test

import (
	"testing"

	"github.com/hironow/dominator/internal/domain"
)

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
