package domain

import "fmt"

// Default values for Config fields.
const (
	DefaultModel      = "opus"
	DefaultTimeoutSec = 1980
)

// Config holds the complete dominator configuration.
type Config struct {
	Target     TargetConfig   `yaml:"target"`
	Nfr        NfrConfig      `yaml:"nfr"`
	Load       LoadConfig     `yaml:"load"`
	Approval   ApprovalConfig `yaml:"approval"`
	Lang       string         `yaml:"lang"`
	ClaudeCmd  string         `yaml:"claude_cmd,omitempty"`
	Model      string         `yaml:"model,omitempty"`
	TimeoutSec int            `yaml:"timeout_sec,omitempty"`
}

// TargetConfig describes the system under test.
type TargetConfig struct {
	URL      string `yaml:"url"`
	Protocol string `yaml:"protocol"` // stored as string, validated via Protocol primitive
	Spec     string `yaml:"spec"`
	Docs     string `yaml:"docs"`
}

// NfrConfig holds non-functional requirement thresholds.
type NfrConfig struct {
	Performance PerformanceNfr `yaml:"performance"`
	Reliability ReliabilityNfr `yaml:"reliability"`
	Scalability ScalabilityNfr `yaml:"scalability"`
}

// PerformanceNfr defines latency and error rate thresholds.
type PerformanceNfr struct {
	P95LatencyMs     int     `yaml:"p95_latency_ms"`
	ErrorRatePercent float64 `yaml:"error_rate_percent"`
}

// ReliabilityNfr defines success rate thresholds.
type ReliabilityNfr struct {
	SuccessRatePercent float64 `yaml:"success_rate_percent"`
}

// ScalabilityNfr defines throughput thresholds.
type ScalabilityNfr struct {
	TargetRPS int `yaml:"target_rps"`
}

// LoadConfig defines load test parameters.
type LoadConfig struct {
	VUs      int    `yaml:"vus"`
	Duration string `yaml:"duration"`
	RampUp   string `yaml:"ramp_up"`
}

// ApproverConfig describes how approval behavior is configured.
// Implemented by ApprovalConfig. Used by session.BuildApprover.
type ApproverConfig interface {
	IsAutoApprove() bool
	ApproveCmdString() string
}

// ApprovalConfig controls whether human approval is required.
// ApprovalConfig implements ApproverConfig.
type ApprovalConfig struct {
	Required   bool   `yaml:"required"`
	ApproveCmd string `yaml:"approve_cmd,omitempty"`
}

// IsAutoApprove reports whether auto-approve is enabled (approval not required).
func (a ApprovalConfig) IsAutoApprove() bool { return !a.Required }

// ApproveCmdString returns the approval command string.
func (a ApprovalConfig) ApproveCmdString() string { return a.ApproveCmd }

// DefaultConfig returns a Config populated with reasonable defaults.
func DefaultConfig() Config {
	return Config{
		Lang:       "ja",
		ClaudeCmd:  DefaultClaudeCmd,
		Model:      DefaultModel,
		TimeoutSec: DefaultTimeoutSec,
		Nfr: NfrConfig{
			Performance: PerformanceNfr{
				P95LatencyMs:     500,
				ErrorRatePercent: 1.0,
			},
			Reliability: ReliabilityNfr{
				SuccessRatePercent: 99.0,
			},
			Scalability: ScalabilityNfr{
				TargetRPS: 100,
			},
		},
		Load: LoadConfig{
			VUs:      10,
			Duration: "30s",
			RampUp:   "5s",
		},
		Approval: ApprovalConfig{
			Required: true,
		},
	}
}

// ValidateConfig checks the config for consistency and returns a list of errors.
// An empty slice means the config is valid.
func ValidateConfig(cfg Config) []string {
	var errs []string

	// Required string fields
	if cfg.ClaudeCmd == "" {
		errs = append(errs, "claude_cmd must not be empty")
	}
	if cfg.Model == "" {
		errs = append(errs, "model must not be empty")
	}

	// Language check
	if !ValidLang(cfg.Lang) {
		errs = append(errs, fmt.Sprintf("lang must be \"ja\" or \"en\" (got %q)", cfg.Lang))
	}

	// Load config checks
	if cfg.Load.VUs <= 0 {
		errs = append(errs, fmt.Sprintf("load.vus must be positive (got %d)", cfg.Load.VUs))
	}
	if cfg.Load.Duration == "" {
		errs = append(errs, "load.duration must not be empty")
	}

	// NFR checks
	if cfg.Nfr.Performance.P95LatencyMs <= 0 {
		errs = append(errs, fmt.Sprintf("nfr.performance.p95_latency_ms must be positive (got %d)", cfg.Nfr.Performance.P95LatencyMs))
	}
	if cfg.Nfr.Performance.ErrorRatePercent < 0 {
		errs = append(errs, fmt.Sprintf("nfr.performance.error_rate_percent must be non-negative (got %f)", cfg.Nfr.Performance.ErrorRatePercent))
	}
	if cfg.Nfr.Reliability.SuccessRatePercent < 0 || cfg.Nfr.Reliability.SuccessRatePercent > 100 {
		errs = append(errs, fmt.Sprintf("nfr.reliability.success_rate_percent must be between 0 and 100 (got %f)", cfg.Nfr.Reliability.SuccessRatePercent))
	}
	if cfg.Nfr.Scalability.TargetRPS < 0 {
		errs = append(errs, fmt.Sprintf("nfr.scalability.target_rps must be non-negative (got %d)", cfg.Nfr.Scalability.TargetRPS))
	}

	// TimeoutSec check
	if cfg.TimeoutSec < 0 {
		errs = append(errs, fmt.Sprintf("timeout_sec must be non-negative (got %d)", cfg.TimeoutSec))
	}

	return errs
}
