package session

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hironow/dominator/internal/domain"
)

// InsightWriter appends judgment results to insight markdown files
// in the .pass/insights/ directory.
type InsightWriter struct {
	StateDir string
	Logger   domain.Logger
}

// insightsDir returns the path to the insights directory.
func (w *InsightWriter) insightsDir() string {
	return filepath.Join(w.StateDir, "insights")
}

// RecordHue appends a summary entry to hue.md for every judgment run.
func (w *InsightWriter) RecordHue(result domain.JudgedData) error {
	dir := w.insightsDir()
	entry := fmt.Sprintf(
		"\n## %s — %s\n\nScript: %s, VUs: %d, Duration: %s\nDeviations: %d\n",
		time.Now().UTC().Format(time.RFC3339),
		result.Verdict,
		result.ScriptPath,
		result.VUs,
		result.Duration,
		len(result.Deviations),
	)

	path := filepath.Join(dir, "hue.md")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open hue.md: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(entry); err != nil {
		return fmt.Errorf("write hue.md: %w", err)
	}

	w.Logger.Info("Recorded hue insight: %s", path)
	return nil
}

// RecordCoefficient appends detailed violation data to coefficient.md.
// Only writes when the verdict is a violation.
func (w *InsightWriter) RecordCoefficient(result domain.JudgedData) error {
	if result.Verdict != domain.VerdictViolation {
		return nil
	}

	dir := w.insightsDir()
	var b strings.Builder
	fmt.Fprintf(&b,
		"\n## %s — Violation\n\n| Metric | Threshold | Actual | Deviation | Severity |\n|--------|-----------|--------|-----------|----------|\n",
		time.Now().UTC().Format(time.RFC3339),
	)
	for _, d := range result.Deviations {
		fmt.Fprintf(&b, "| %s | %.2f | %.2f | %.1f%% | %s |\n",
			d.Metric, d.Threshold, d.Actual, d.Deviation, d.Severity)
	}
	entry := b.String()

	path := filepath.Join(dir, "coefficient.md")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open coefficient.md: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(entry); err != nil {
		return fmt.Errorf("write coefficient.md: %w", err)
	}

	w.Logger.Info("Recorded coefficient insight: %s", path)
	return nil
}
