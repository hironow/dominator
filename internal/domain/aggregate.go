package domain

import (
	"fmt"
	"time"
)

// JudgeAggregate is the core domain aggregate for NFR validation.
// It holds the configuration and orchestrates judgment logic.
type JudgeAggregate struct {
	config Config
}

// NewJudgeAggregate creates a JudgeAggregate with the given configuration.
func NewJudgeAggregate(cfg Config) *JudgeAggregate {
	return &JudgeAggregate{config: cfg}
}

// Config returns the aggregate's configuration.
func (j *JudgeAggregate) Config() Config { return j.config }

// RecordJudgment produces an EventJudgmentRecorded for a session-initiated
// judgment (= the dominator.record_result MCP tool). The session judged
// externally via mcp-k6 and reports only the verdict + summary, so this
// method does not consult the aggregate's NFR config. The verdict string
// from the MCP call ("pass" | "fail") is normalised to a domain.Verdict.
func (j *JudgeAggregate) RecordJudgment(targetID, verdict, summary string, now time.Time) (Event, error) {
	if targetID == "" {
		return Event{}, fmt.Errorf("target_id is required")
	}
	var v Verdict
	switch verdict {
	case "pass":
		v = VerdictPass
	case "fail":
		v = VerdictViolation
	default:
		return Event{}, fmt.Errorf("invalid verdict %q: must be 'pass' or 'fail'", verdict)
	}
	return NewEvent(EventJudgmentRecorded, JudgmentRecordedData{
		TargetID: targetID,
		Verdict:  v,
		Summary:  summary,
	}, now)
}
