package session

import (
	"fmt"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"

	"github.com/hironow/dominator/internal/domain"
)

// UpdateConfig loads the config at stateDir, sets a single key to the given
// value, and saves it back. Returns an error for unknown keys or invalid values.
func UpdateConfig(stateDir, key, value string) error { // nosemgrep: domain-primitives.multiple-string-params-go -- stateDir/key/value are semantically distinct: a path, a config key name, and a config value; swapping them causes failures caught by tests [permanent]
	cfgPath := filepath.Join(stateDir, domain.ConfigFile)
	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	updated, err := setConfigField(cfg, key, value)
	if err != nil {
		return err
	}
	return SaveConfig(cfgPath, updated)
}

// setConfigField sets a single field on cfg identified by a dotted key.
// It accepts cfg by value and returns a modified copy, preserving referential
// transparency and avoiding hidden mutation of the caller's data.
func setConfigField(cfg domain.Config, key, value string) (domain.Config, error) {
	switch key {
	case "target.url":
		cfg.Target.URL = value
	case "target.protocol":
		cfg.Target.Protocol = value
	case "target.spec":
		cfg.Target.Spec = value
	case "target.docs":
		cfg.Target.Docs = value
	case "nfr.performance.p95_latency_ms":
		v, err := strconv.Atoi(value)
		if err != nil {
			return domain.Config{}, fmt.Errorf("invalid integer for %s: %w", key, err)
		}
		perf := cfg.Nfr.Performance
		perf.P95LatencyMs = v
		cfg.Nfr.Performance = perf
	case "nfr.performance.error_rate_percent":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return domain.Config{}, fmt.Errorf("invalid float for %s: %w", key, err)
		}
		perf := cfg.Nfr.Performance
		perf.ErrorRatePercent = v
		cfg.Nfr.Performance = perf
	case "nfr.reliability.success_rate_percent":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return domain.Config{}, fmt.Errorf("invalid float for %s: %w", key, err)
		}
		rel := cfg.Nfr.Reliability
		rel.SuccessRatePercent = v
		cfg.Nfr.Reliability = rel
	case "nfr.scalability.target_rps":
		v, err := strconv.Atoi(value)
		if err != nil {
			return domain.Config{}, fmt.Errorf("invalid integer for %s: %w", key, err)
		}
		scal := cfg.Nfr.Scalability
		scal.TargetRPS = v
		cfg.Nfr.Scalability = scal
	case "load.vus":
		v, err := strconv.Atoi(value)
		if err != nil {
			return domain.Config{}, fmt.Errorf("invalid integer for %s: %w", key, err)
		}
		cfg.Load.VUs = v
	case "load.duration":
		cfg.Load.Duration = value
	case "load.ramp_up":
		cfg.Load.RampUp = value
	case "approval.required":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return domain.Config{}, fmt.Errorf("invalid boolean for %s: %w", key, err)
		}
		cfg.Approval.Required = v
	case "lang":
		cfg.Lang = value
	case "claude_cmd":
		cfg.ClaudeCmd = value
	case "model":
		cfg.Model = value
	case "timeout_sec":
		v, err := strconv.Atoi(value)
		if err != nil {
			return domain.Config{}, fmt.Errorf("invalid integer for %s: %w", key, err)
		}
		cfg.TimeoutSec = v
	default:
		return domain.Config{}, fmt.Errorf("unknown config key: %s", key)
	}
	return cfg, nil
}

// ShowConfig loads and returns the config as YAML bytes.
func ShowConfig(stateDir string) ([]byte, error) {
	cfgPath := filepath.Join(stateDir, domain.ConfigFile)
	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("marshal config: %w", err)
	}
	return data, nil
}
