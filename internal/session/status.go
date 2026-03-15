package session

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/hironow/dominator/internal/domain"
)

// Status collects current operational status from the event store and filesystem.
// passDir is the .pass/ directory path (e.g. "<repo>/.pass").
func Status(ctx context.Context, passDir string, logger domain.Logger) domain.StatusReport {
	var report domain.StatusReport

	// Count inbox files
	report.InboxCount = countDirFiles(passDir, "inbox")

	// Count archive files
	report.ArchiveCount = countDirFiles(passDir, "archive")

	// Count k6 scripts
	report.K6Scripts = countDirFiles(passDir, "k6")

	// Count plans (approved vs pending)
	plansDir := filepath.Join(passDir, ".run", "plans")
	entries, err := os.ReadDir(plansDir)
	if err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			report.PlanCount++
			data, readErr := os.ReadFile(filepath.Join(plansDir, e.Name()))
			if readErr != nil {
				continue
			}
			var plan domain.Plan
			if jsonErr := json.Unmarshal(data, &plan); jsonErr != nil {
				continue
			}
			if plan.Approved {
				report.ApprovedPlan++
			} else {
				report.PendingPlan++
			}
		}
	}

	// Load all events for judgment stats
	store := NewEventStore(passDir, logger)
	allEvents, _, loadErr := store.LoadAll()
	if loadErr != nil || len(allEvents) == 0 {
		return report
	}

	var lastJudgment time.Time
	var lastVerdict domain.Verdict
	for _, ev := range allEvents {
		switch ev.Type {
		case domain.EventJudged:
			report.JudgeCount++
			var data domain.JudgedData
			if jsonErr := json.Unmarshal(ev.Data, &data); jsonErr == nil {
				if ev.Timestamp.After(lastJudgment) {
					lastJudgment = ev.Timestamp
					lastVerdict = data.Verdict
				}
			}
		case domain.EventPassConfirmed:
			report.JudgeCount++
			if ev.Timestamp.After(lastJudgment) {
				lastJudgment = ev.Timestamp
				lastVerdict = domain.VerdictPass
			}
		case domain.EventViolationDetected:
			report.JudgeCount++
			if ev.Timestamp.After(lastJudgment) {
				lastJudgment = ev.Timestamp
				lastVerdict = domain.VerdictViolation
			}
		}
	}

	report.LastJudgment = lastJudgment
	report.Verdict = lastVerdict

	return report
}

// countDirFiles returns the number of non-directory entries in a subdirectory of baseDir.
// Returns 0 if the directory does not exist or cannot be read.
func countDirFiles(baseDir string, sub string) int {
	dir := baseDir
	if sub != "" {
		dir = filepath.Join(baseDir, sub)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if !e.IsDir() {
			count++
		}
	}
	return count
}
