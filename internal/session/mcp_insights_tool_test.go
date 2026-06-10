package session_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/session"
)

// refs issue 0034 (P4): get_insights exposes the NFR judge's learning
// loop. RecordHue/RecordCoefficient have had zero production callers
// since the MCP pivot (verified 2026-06-10), so the live judgment
// summary derived from EventJudgmentRecorded is the source of truth;
// hue.md / coefficient.md are returned as the legacy ledger when
// present.

func callDomInsights(t *testing.T, passDir, args string) map[string]any {
	t.Helper()
	req := `{"jsonrpc":"2.0","id":90,"method":"tools/call","params":{"name":"get_insights","arguments":` + args + `}}` + "\n"
	var out bytes.Buffer
	srv := session.NewMCPServer(strings.NewReader(req), &out, nil).WithPassDir(passDir)
	if err := srv.Serve(context.Background()); err != nil {
		t.Fatalf("Serve: %v", err)
	}
	return decodeDMailToolJSON(t, out.Bytes())
}

func seedJudgment(t *testing.T, passDir, targetID, verdict string) {
	t.Helper()
	ev, err := domain.NewEvent(domain.EventJudgmentRecorded, domain.JudgmentRecordedData{
		TargetID: targetID, Verdict: domain.Verdict(verdict), Summary: "s",
	}, time.Now().UTC())
	if err != nil {
		t.Fatalf("NewEvent: %v", err)
	}
	store := session.NewEventStore(passDir, nil)
	if _, err := store.Append(ev); err != nil {
		t.Fatalf("seed append: %v", err)
	}
}

func TestMCPServer_ToolsList_IncludesGetInsights(t *testing.T) {
	// given
	in := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}` + "\n")
	var out bytes.Buffer
	srv := session.NewMCPServer(in, &out, nil)

	// when
	if err := srv.Serve(context.Background()); err != nil {
		t.Fatalf("Serve: %v", err)
	}

	// then
	var resp map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(out.Bytes()), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	tools := resp["result"].(map[string]any)["tools"].([]any)
	found := false
	for _, t0 := range tools {
		entry, _ := t0.(map[string]any)
		if entry["name"] == "get_insights" {
			found = true
		}
	}
	if !found {
		t.Error("missing get_insights in tools/list")
	}
}

func TestMCPServer_GetInsights_EmptyStateIsNotAnError(t *testing.T) {
	// given
	passDir := filepath.Join(t.TempDir(), ".pass")

	// when
	body := callDomInsights(t, passDir, `{}`)

	// then
	if body["initialized"] != true {
		t.Fatalf("initialized = %v, want true (body=%v)", body["initialized"], body)
	}
	hue, _ := body["hue"].([]any)
	if len(hue) != 0 {
		t.Errorf("hue = %v, want empty", body["hue"])
	}
}

func TestMCPServer_GetInsights_LiveJudgmentSummaryFromEvents(t *testing.T) {
	// given: two recorded judgments (the live post-pivot source)
	passDir := filepath.Join(t.TempDir(), ".pass")
	seedJudgment(t, passDir, "api", "pass")
	seedJudgment(t, passDir, "api", "fail")

	// when
	body := callDomInsights(t, passDir, `{}`)

	// then
	live, _ := body["live_judgments"].(map[string]any)
	if live == nil {
		t.Fatalf("live_judgments missing: %v", body)
	}
	if int(live["count"].(float64)) != 2 {
		t.Errorf("count = %v, want 2", live["count"])
	}
	if live["last_verdict"] != "fail" {
		t.Errorf("last_verdict = %v, want fail", live["last_verdict"])
	}
	if live["last_target_id"] != "api" {
		t.Errorf("last_target_id = %v, want api", live["last_target_id"])
	}
}

func TestMCPServer_GetInsights_ReadsLegacyHueLedger(t *testing.T) {
	// given: a legacy hue.md (pre-pivot ledger format)
	passDir := filepath.Join(t.TempDir(), ".pass")
	insightsDir := filepath.Join(passDir, "insights")
	if err := os.MkdirAll(insightsDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	hue := `# Hue

## 2026-05-01T10:00:00Z — pass

- Script: k6-scripts/api.js, VUs: 10, Duration: 1m
`
	if err := os.WriteFile(filepath.Join(insightsDir, "hue.md"), []byte(hue), 0o644); err != nil {
		t.Fatalf("write hue: %v", err)
	}

	// when
	body := callDomInsights(t, passDir, `{}`)

	// then
	entries, _ := body["hue"].([]any)
	if len(entries) != 1 {
		t.Fatalf("hue = %v, want 1 legacy entry (body=%v)", body["hue"], body)
	}
	first, _ := entries[0].(map[string]any)
	if first["verdict"] != "pass" {
		t.Errorf("hue[0].verdict = %v, want pass", first["verdict"])
	}
}

func TestMCPServer_GetInsights_UninitializedWithoutPassDir(t *testing.T) {
	// given
	req := `{"jsonrpc":"2.0","id":91,"method":"tools/call","params":{"name":"get_insights","arguments":{}}}` + "\n"
	var out bytes.Buffer
	srv := session.NewMCPServer(strings.NewReader(req), &out, nil)

	// when
	if err := srv.Serve(context.Background()); err != nil {
		t.Fatalf("Serve: %v", err)
	}

	// then
	body := decodeDMailToolJSON(t, out.Bytes())
	if body["initialized"] != false {
		t.Errorf("initialized = %v, want false", body["initialized"])
	}
}
