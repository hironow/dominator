package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// StatusReport holds operational status information for the dominator tool.
type StatusReport struct {
	LastJudgment time.Time `json:"last_judgment"`
	Verdict      Verdict   `json:"verdict"`
	PlanCount    int       `json:"plan_count"`
	ApprovedPlan int       `json:"approved_plan_count"`
	PendingPlan  int       `json:"pending_plan_count"`
	K6Scripts    int       `json:"k6_scripts"`
	InboxCount   int       `json:"inbox_count"`
	ArchiveCount int       `json:"archive_count"`
	JudgeCount   int       `json:"judge_count"`
}

// FormatText returns a human-readable status report string suitable for stdout.
func (r StatusReport) FormatText() string {
	var b strings.Builder
	b.WriteString("dominator status\n\n")

	// Last judgment
	if r.LastJudgment.IsZero() {
		fmt.Fprintf(&b, "  %-18s %s\n", "Last judgment:", "no judgments yet")
	} else {
		fmt.Fprintf(&b, "  %-18s %s\n", "Last judgment:", r.LastJudgment.Format(time.RFC3339))
	}

	// Verdict
	if r.Verdict == "" {
		fmt.Fprintf(&b, "  %-18s %s\n", "Verdict:", "none")
	} else {
		fmt.Fprintf(&b, "  %-18s %s\n", "Verdict:", string(r.Verdict))
	}

	fmt.Fprintf(&b, "  %-18s %d total (%d approved, %d pending)\n", "Plans:", r.PlanCount, r.ApprovedPlan, r.PendingPlan)
	fmt.Fprintf(&b, "  %-18s %d\n", "k6 scripts:", r.K6Scripts)
	fmt.Fprintf(&b, "  %-18s %d total\n", "Judgments:", r.JudgeCount)
	fmt.Fprintf(&b, "  %-18s %d pending\n", "Inbox:", r.InboxCount)
	fmt.Fprintf(&b, "  %-18s %d processed\n", "Archive:", r.ArchiveCount)

	return b.String()
}

// FormatJSON returns the status report as a compact JSON string.
func (r StatusReport) FormatJSON() string {
	data, err := json.Marshal(r)
	if err != nil {
		return fmt.Sprintf(`{"error":%q}`, err.Error())
	}
	return string(data)
}
