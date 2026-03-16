package session

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hironow/dominator/internal/domain"
)

// DMailEmitter creates D-Mail files in the .pass/outbox/ directory.
// These files are picked up by phonewave for routing to other tools.
type DMailEmitter struct {
	StateDir string
	Logger   domain.Logger
}

// outboxDir returns the path to the outbox directory.
func (e *DMailEmitter) outboxDir() string {
	return filepath.Join(e.StateDir, "outbox")
}

// EmitViolation creates 3 D-Mail files for NFR violations:
// design-feedback (Sightjack), implementation-feedback (Paintress),
// verification-feedback (Amadeus).
func (e *DMailEmitter) EmitViolation(result domain.JudgedData) error {
	dir := e.outboxDir()
	ts := time.Now().UTC().Format("20060102T150405Z")
	maxSeverity := maxDeviationSeverity(result.Deviations)
	table := buildDeviationTable(result.Deviations)

	targets := []struct {
		kind string
		desc string
	}{
		{"design-feedback", "NFR violation detected — design review recommended"},
		{"implementation-feedback", "NFR violation detected — implementation review recommended"},
		{"verification-feedback", "NFR violation detected — verification review recommended"},
	}

	for _, t := range targets {
		name := fmt.Sprintf("%s-%s.md", t.kind, ts)
		content := fmt.Sprintf(`---
name: %s-%s
kind: %s
description: "NFR violation: %d deviation(s) detected"
severity: %s
---

# NFR Violation Report

Plan: %s
Script: %s, VUs: %d, Duration: %s

%s
`, t.kind, ts, t.kind, len(result.Deviations), maxSeverity,
			result.PlanID, result.ScriptPath, result.VUs, result.Duration, table)

		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return fmt.Errorf("write D-Mail %s: %w", name, err)
		}
		e.Logger.Info("Emitted D-Mail: %s", name)
	}

	return nil
}

// EmitPass creates a single informational D-Mail when all NFRs pass.
func (e *DMailEmitter) EmitPass(result domain.JudgedData) error {
	dir := e.outboxDir()
	ts := time.Now().UTC().Format("20060102T150405Z")
	name := fmt.Sprintf("nfr-pass-%s.md", ts)

	content := fmt.Sprintf(`---
name: nfr-pass-%s
kind: report
description: "All NFR thresholds met"
severity: low
---

# NFR Pass Report

Plan: %s
Script: %s, VUs: %d, Duration: %s
Verdict: pass
`, ts, result.PlanID, result.ScriptPath, result.VUs, result.Duration)

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write D-Mail %s: %w", name, err)
	}
	e.Logger.Info("Emitted D-Mail: %s", name)

	return nil
}

// maxDeviationSeverity returns the highest severity among deviations.
func maxDeviationSeverity(deviations []domain.NfrDeviation) string {
	max := domain.SeverityLow
	for _, d := range deviations {
		if severityRank(d.Severity) > severityRank(max) {
			max = d.Severity
		}
	}
	return max
}

// severityRank maps severity string to a comparable integer.
func severityRank(s string) int {
	switch s {
	case domain.SeverityHigh:
		return 3
	case domain.SeverityMedium:
		return 2
	case domain.SeverityLow:
		return 1
	default:
		return 0
	}
}

// buildDeviationTable builds a markdown table from deviations.
func buildDeviationTable(deviations []domain.NfrDeviation) string {
	var b strings.Builder
	b.WriteString("| Metric | Threshold | Actual | Deviation | Severity |\n|--------|-----------|--------|-----------|----------|\n")
	for _, d := range deviations {
		fmt.Fprintf(&b, "| %s | %.2f | %.2f | %.1f%% | %s |\n",
			d.Metric, d.Threshold, d.Actual, d.Deviation, d.Severity)
	}
	return b.String()
}
