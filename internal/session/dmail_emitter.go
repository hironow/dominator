package session

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/platform/projectid"
)

// projectIDMetadataBlock returns the multi-line YAML metadata block to
// embed in a D-Mail frontmatter when project_id can be resolved. Returns
// "" (empty) when unresolved so legacy single-mode renders byte-identical
// to pre-multiplex output.
func projectIDMetadataBlock() string {
	cwd, _ := os.Getwd()
	id, _ := projectid.Resolve(cwd)
	if id == "" {
		return ""
	}
	return fmt.Sprintf("metadata:\n  project_id: %s\n", id)
}

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
// and report (Amadeus context). All three kinds are canonical D-Mail v1
// kinds; the Rival Contract v1 protocol forbids the non-canonical
// 'verification-feedback' kind because strict routing validators reject it.
func (e *DMailEmitter) EmitViolation(result domain.JudgedData) error {
	dir := e.outboxDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create outbox dir: %w", err)
	}
	ts := time.Now().UTC().Format("20060102T150405Z")
	maxSeverity := maxDeviationSeverity(result.Deviations)
	table := buildDeviationTable(result.Deviations)

	targets := []struct {
		kind string
		desc string
	}{
		{"design-feedback", "NFR violation detected — design review recommended"},
		{"implementation-feedback", "NFR violation detected — implementation review recommended"},
		{"report", "NFR violation detected — verifier context for amadeus"},
	}

	pidBlock := projectIDMetadataBlock()
	for _, t := range targets {
		name := fmt.Sprintf("%s-%s.md", t.kind, ts)
		content := fmt.Sprintf(`---
name: %s-%s
kind: %s
description: "NFR violation: %d deviation(s) detected"
severity: %s
%s---

# NFR Violation Report

Plan: %s
Script: %s, VUs: %d, Duration: %s

%s
`, t.kind, ts, t.kind, len(result.Deviations), maxSeverity, pidBlock,
			result.PlanID, result.ScriptPath, result.VUs, result.Duration, table)

		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return fmt.Errorf("write D-Mail %s: %w", name, err)
		}
		e.Logger.Info("Emitted D-Mail: %s", name)
	}

	return nil
}

// EmitDesignFeedbackMissingNfr emits a single design-feedback D-Mail
// requesting that the contract author (sightjack) add the listed
// `nfr.*` thresholds to the current Rival Contract v1. The message is
// deliberately specific: it cites the contract id and lists every key
// that was missing so the contract author can add them deterministically.
//
// This is the dominator's response when a contract exists but does not
// declare the NFR thresholds required to drive a deterministic load
// test. The dominator must NOT invent thresholds; instead it asks the
// contract author to revise the contract.
func (e *DMailEmitter) EmitDesignFeedbackMissingNfr(missing []string, contractID string) error {
	dir := e.outboxDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create outbox dir: %w", err)
	}
	ts := time.Now().UTC().Format("20060102T150405Z")
	name := fmt.Sprintf("design-feedback-%s.md", ts)
	path := filepath.Join(dir, name)

	var bullets strings.Builder
	for _, key := range missing {
		fmt.Fprintf(&bullets, "- %s\n", key)
	}

	content := fmt.Sprintf(`---
name: design-feedback-%s
kind: design-feedback
description: "Rival Contract v1 missing required NFR thresholds for load test"
severity: medium
contract_id: %q
%s---

# Missing NFR Thresholds in Rival Contract v1

The current Rival Contract v1 (contract_id: %s) does not declare the
following required NFR thresholds. The dominator cannot drive a
deterministic load test without them and will not invent values.

Please add the following keys to the contract's Evidence section:

%s
Each bullet must follow the deterministic shape:

    - <key>: <op> <value>

Example: `+"`- nfr.p95_latency_ms: <= 300`"+`
`, ts, contractID, projectIDMetadataBlock(), contractID, bullets.String())

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write D-Mail %s: %w", name, err)
	}
	e.Logger.Info("Emitted D-Mail: %s", name)
	return nil
}

// EmitPass creates a single informational D-Mail when all NFRs pass.
func (e *DMailEmitter) EmitPass(result domain.JudgedData) error {
	dir := e.outboxDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create outbox dir: %w", err)
	}
	ts := time.Now().UTC().Format("20060102T150405Z")
	name := fmt.Sprintf("nfr-pass-%s.md", ts)

	content := fmt.Sprintf(`---
name: nfr-pass-%s
kind: report
description: "All NFR thresholds met"
severity: low
%s---

# NFR Pass Report

Plan: %s
Script: %s, VUs: %d, Duration: %s
Verdict: pass
`, ts, projectIDMetadataBlock(), result.PlanID, result.ScriptPath, result.VUs, result.Duration)

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
