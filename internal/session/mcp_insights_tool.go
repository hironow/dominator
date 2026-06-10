package session

import (
	"encoding/json"
	"fmt"

	"github.com/hironow/dominator/internal/domain"
)

// realGetInsights exposes the NFR judge's learning loop to the session
// (refs issue 0034 P4). RecordHue / RecordCoefficient have had zero
// production callers since the MCP pivot (verified 2026-06-10), so the
// live judgment summary derived from EventJudgmentRecorded is the
// source of truth; hue.md / coefficient.md are surfaced as the legacy
// ledger when present. Read-only and idempotent; empty state returns
// empty arrays, not an error.
func realGetInsights(passDir string, logger domain.Logger, _ json.RawMessage) map[string]any {
	if passDir == "" {
		return jsonResult(map[string]any{
			"initialized": false,
			"reason":      "dominator mcp passDir not configured (start `dominator mcp` from the project root)",
		})
	}
	if logger == nil {
		logger = &domain.NopLogger{}
	}

	reader := &InsightReader{StateDir: passDir, Logger: logger}
	hue, hueErr := reader.ReadHue()
	coeff, coeffErr := reader.ReadCoefficient()
	hueOut := make([]map[string]any, 0, len(hue))
	for _, h := range hue {
		hueOut = append(hueOut, map[string]any{
			"timestamp": h.Timestamp, "verdict": h.Verdict, "details": h.Details,
		})
	}
	coeffOut := make([]map[string]any, 0, len(coeff))
	for _, c := range coeff {
		coeffOut = append(coeffOut, map[string]any{
			"timestamp": c.Timestamp, "label": c.Label, "table": c.Table,
		})
	}

	live := map[string]any{"count": 0}
	store := NewEventStore(passDir, logger)
	if events, _, err := store.LoadAll(); err == nil {
		count := 0
		var lastTarget, lastVerdict, lastAt string
		fails := 0
		for _, ev := range events {
			if ev.Type != domain.EventJudgmentRecorded {
				continue
			}
			var data domain.JudgmentRecordedData
			if jsonErr := json.Unmarshal(ev.Data, &data); jsonErr != nil {
				continue
			}
			count++
			lastTarget = data.TargetID
			lastVerdict = string(data.Verdict)
			lastAt = ev.Timestamp.UTC().Format("2006-01-02T15:04:05Z")
			if string(data.Verdict) != "pass" {
				fails++
			}
		}
		live["count"] = count
		live["fail_count"] = fails
		if count > 0 {
			live["last_target_id"] = lastTarget
			live["last_verdict"] = lastVerdict
			live["last_at"] = lastAt
		}
	}

	res := map[string]any{
		"initialized":    true,
		"passDir":        passDir,
		"live_judgments": live,
		"hue":            hueOut,
		"coefficient":    coeffOut,
		"instruction":    fmt.Sprintf("Review past verdicts before judging: repeated fail verdicts on a target indicate a persistent NFR gap. %v judgment(s) recorded; hue ledger has %d legacy entrie(s).", live["count"], len(hueOut)),
	}
	if hueErr != nil {
		res["hue_error"] = hueErr.Error()
	}
	if coeffErr != nil {
		res["coefficient_error"] = coeffErr.Error()
	}
	return jsonResult(res)
}
