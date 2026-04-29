package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/hironow/dominator/internal/domain"
	"gopkg.in/yaml.v3"
)

// PlanStore persists and retrieves execution plans as JSON files
// in the .pass/.run/plans/ directory.
type PlanStore struct {
	StateDir string
}

// plansDir returns the directory path for storing plan files.
func (s *PlanStore) plansDir() string {
	return filepath.Join(s.StateDir, ".run", "plans")
}

// SavePlan writes a Plan to disk as a JSON file named by its ID.
func (s *PlanStore) SavePlan(plan domain.Plan) error {
	dir := s.plansDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create plans dir: %w", err)
	}
	data, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal plan: %w", err)
	}
	path := filepath.Join(dir, string(plan.ID)+".json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write plan: %w", err)
	}
	return nil
}

// LoadPlan reads a specific plan by its ID.
func (s *PlanStore) LoadPlan(planID domain.PlanID) (*domain.Plan, error) {
	path := filepath.Join(s.plansDir(), string(planID)+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("plan %s not found", planID)
		}
		return nil, fmt.Errorf("read plan: %w", err)
	}
	var plan domain.Plan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, fmt.Errorf("unmarshal plan: %w", err)
	}
	return &plan, nil
}

// LoadLatestPlan finds the most recently created plan file and returns it.
func (s *PlanStore) LoadLatestPlan() (*domain.Plan, error) {
	dir := s.plansDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("no plans found")
		}
		return nil, fmt.Errorf("read plans dir: %w", err)
	}

	var jsonFiles []os.DirEntry
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			jsonFiles = append(jsonFiles, e)
		}
	}
	if len(jsonFiles) == 0 {
		return nil, fmt.Errorf("no plans found")
	}

	// Sort by modification time, newest first.
	sort.Slice(jsonFiles, func(i, j int) bool {
		infoI, _ := jsonFiles[i].Info()
		infoJ, _ := jsonFiles[j].Info()
		return infoI.ModTime().After(infoJ.ModTime())
	})

	data, err := os.ReadFile(filepath.Join(dir, jsonFiles[0].Name()))
	if err != nil {
		return nil, fmt.Errorf("read latest plan: %w", err)
	}
	var plan domain.Plan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, fmt.Errorf("unmarshal latest plan: %w", err)
	}
	return &plan, nil
}

// ApprovePlan loads a plan by ID, approves it, and saves it back.
func (s *PlanStore) ApprovePlan(planID domain.PlanID) (*domain.Plan, error) {
	plan, err := s.LoadPlan(planID)
	if err != nil {
		return nil, err
	}
	plan.Approve()
	if err := s.SavePlan(*plan); err != nil {
		return nil, fmt.Errorf("save approved plan: %w", err)
	}
	return plan, nil
}

// ListScripts returns the filenames of .js files in the k6-scripts directory.
func (s *PlanStore) ListScripts() ([]string, error) {
	dir := filepath.Join(s.StateDir, "k6-scripts")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read k6-scripts dir: %w", err)
	}
	var scripts []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".js") {
			scripts = append(scripts, e.Name())
		}
	}
	return scripts, nil
}

// UpdateConfig loads the config at stateDir, sets a single key to the given
// value, and saves it back. Returns an error for unknown keys or invalid values.
func UpdateConfig(stateDir, key, value string) error {
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
		cfg.Nfr.Performance.P95LatencyMs = v
	case "nfr.performance.error_rate_percent":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return domain.Config{}, fmt.Errorf("invalid float for %s: %w", key, err)
		}
		cfg.Nfr.Performance.ErrorRatePercent = v
	case "nfr.reliability.success_rate_percent":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return domain.Config{}, fmt.Errorf("invalid float for %s: %w", key, err)
		}
		cfg.Nfr.Reliability.SuccessRatePercent = v
	case "nfr.scalability.target_rps":
		v, err := strconv.Atoi(value)
		if err != nil {
			return domain.Config{}, fmt.Errorf("invalid integer for %s: %w", key, err)
		}
		cfg.Nfr.Scalability.TargetRPS = v
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
