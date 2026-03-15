package session

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hironow/dominator/internal/domain"
)

// HueEntry represents a single section parsed from hue.md.
type HueEntry struct {
	Timestamp string `json:"timestamp"`
	Verdict   string `json:"verdict"`
	Details   string `json:"details"`
}

// CoefficientEntry represents a single section parsed from coefficient.md.
type CoefficientEntry struct {
	Timestamp string `json:"timestamp"`
	Label     string `json:"label"`
	Table     string `json:"table"`
}

// InsightReader reads insight markdown files from the .pass/insights/ directory.
type InsightReader struct {
	StateDir string
	Logger   domain.Logger
}

// insightsDir returns the path to the insights directory.
func (r *InsightReader) insightsDir() string {
	return filepath.Join(r.StateDir, "insights")
}

// ReadHue parses hue.md sections into structured data.
func (r *InsightReader) ReadHue() ([]HueEntry, error) {
	path := filepath.Join(r.insightsDir(), "hue.md")
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read hue.md: %w", err)
	}

	return parseHueSections(string(content)), nil
}

// ReadCoefficient parses coefficient.md sections into structured data.
func (r *InsightReader) ReadCoefficient() ([]CoefficientEntry, error) {
	path := filepath.Join(r.insightsDir(), "coefficient.md")
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read coefficient.md: %w", err)
	}

	return parseCoefficientSections(string(content)), nil
}

// parseHueSections splits hue.md content at "## " headings.
// Each heading is expected to be "## timestamp — verdict".
func parseHueSections(content string) []HueEntry {
	var entries []HueEntry
	scanner := bufio.NewScanner(strings.NewReader(content))

	var current *HueEntry
	var bodyLines []string

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "## ") {
			// Flush previous entry
			if current != nil {
				current.Details = strings.TrimSpace(strings.Join(bodyLines, "\n"))
				entries = append(entries, *current)
			}
			// Parse heading: "## timestamp — verdict"
			heading := strings.TrimPrefix(line, "## ")
			ts, verdict := splitHeading(heading)
			current = &HueEntry{
				Timestamp: ts,
				Verdict:   verdict,
			}
			bodyLines = nil
			continue
		}
		if current != nil {
			bodyLines = append(bodyLines, line)
		}
	}

	// Flush last entry
	if current != nil {
		current.Details = strings.TrimSpace(strings.Join(bodyLines, "\n"))
		entries = append(entries, *current)
	}

	return entries
}

// parseCoefficientSections splits coefficient.md content at "## " headings.
// Each heading is expected to be "## timestamp — label".
func parseCoefficientSections(content string) []CoefficientEntry {
	var entries []CoefficientEntry
	scanner := bufio.NewScanner(strings.NewReader(content))

	var current *CoefficientEntry
	var bodyLines []string

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "## ") {
			// Flush previous entry
			if current != nil {
				current.Table = strings.TrimSpace(strings.Join(bodyLines, "\n"))
				entries = append(entries, *current)
			}
			// Parse heading: "## timestamp — label"
			heading := strings.TrimPrefix(line, "## ")
			ts, label := splitHeading(heading)
			current = &CoefficientEntry{
				Timestamp: ts,
				Label:     label,
			}
			bodyLines = nil
			continue
		}
		if current != nil {
			bodyLines = append(bodyLines, line)
		}
	}

	// Flush last entry
	if current != nil {
		current.Table = strings.TrimSpace(strings.Join(bodyLines, "\n"))
		entries = append(entries, *current)
	}

	return entries
}

// splitHeading splits "timestamp — label" into (timestamp, label).
// Uses em dash (—) as delimiter; falls back to whole string as timestamp.
func splitHeading(heading string) (string, string) {
	// Try em dash first
	if idx := strings.Index(heading, " — "); idx >= 0 {
		return strings.TrimSpace(heading[:idx]), strings.TrimSpace(heading[idx+len(" — "):])
	}
	// Try en dash
	if idx := strings.Index(heading, " – "); idx >= 0 {
		return strings.TrimSpace(heading[:idx]), strings.TrimSpace(heading[idx+len(" – "):])
	}
	return strings.TrimSpace(heading), ""
}
