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

// recordingJudgmentEmitter is a session-local fake implementing
// port.JudgmentEventEmitter (structurally; no usecase import needed,
// honouring layer-session-no-import-usecase). It records the emit calls
// so tests can assert the record_result → emitter wiring.
type recordingJudgmentEmitter struct {
	calls []recordedJudgment
	err   error
}

type recordedJudgment struct {
	targetID string
	verdict  string
	summary  string
}

func (r *recordingJudgmentEmitter) EmitJudgmentRecorded(targetID, verdict, summary string, _ time.Time) error {
	if r.err != nil {
		return r.err
	}
	r.calls = append(r.calls, recordedJudgment{targetID: targetID, verdict: verdict, summary: summary})
	return nil
}

func TestMCPServer_ListsAllPhase2cTools(t *testing.T) {
	// given
	in := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}` + "\n")
	var out bytes.Buffer
	srv := session.NewMCPServer(in, &out, nil)

	// when
	if err := srv.Serve(context.Background()); err != nil {
		t.Fatalf("Serve: %v", err)
	}

	// then: all 3 Phase 2c tools advertised, with stable names
	var resp map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(out.Bytes()), &resp); err != nil {
		t.Fatalf("decode response: %v (raw=%q)", err, out.String())
	}
	if resp["jsonrpc"] != "2.0" {
		t.Errorf("jsonrpc = %v, want 2.0", resp["jsonrpc"])
	}
	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("missing result: %v", resp)
	}
	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatalf("tools list missing: %v", result["tools"])
	}
	want := map[string]bool{
		"dominator.ping":          false,
		"dominator.get_nfr":       false,
		"dominator.record_result": false,
	}
	for _, t0 := range tools {
		entry, _ := t0.(map[string]any)
		if name, _ := entry["name"].(string); name != "" {
			want[name] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("missing Phase 2c tool: %s", name)
		}
	}
}

func TestMCPServer_CallsPingTool(t *testing.T) {
	// given
	in := strings.NewReader(`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"dominator.ping","arguments":{}}}` + "\n")
	var out bytes.Buffer
	srv := session.NewMCPServer(in, &out, nil)

	// when
	if err := srv.Serve(context.Background()); err != nil {
		t.Fatalf("Serve: %v", err)
	}

	// then
	var resp map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(out.Bytes()), &resp); err != nil {
		t.Fatalf("decode response: %v (raw=%q)", err, out.String())
	}
	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("missing result: %v", resp)
	}
	content, ok := result["content"].([]any)
	if !ok || len(content) != 1 {
		t.Fatalf("content list mismatch: %v", result["content"])
	}
	first, _ := content[0].(map[string]any)
	if first["text"] != "pong" {
		t.Errorf("text = %v, want pong", first["text"])
	}
}

func TestMCPServer_RejectsUnknownTool(t *testing.T) {
	// given
	in := strings.NewReader(`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"dominator.does_not_exist","arguments":{}}}` + "\n")
	var out bytes.Buffer
	srv := session.NewMCPServer(in, &out, nil)

	// when
	if err := srv.Serve(context.Background()); err != nil {
		t.Fatalf("Serve: %v", err)
	}

	// then
	var resp map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(out.Bytes()), &resp); err != nil {
		t.Fatalf("decode response: %v (raw=%q)", err, out.String())
	}
	rpcErr, ok := resp["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error, got %v", resp)
	}
	if code, _ := rpcErr["code"].(float64); int(code) != -32601 {
		t.Errorf("error code = %v, want -32601", rpcErr["code"])
	}
}

func TestMCPServer_GetNFR_UninitializedPassDir(t *testing.T) {
	// given: NewMCPServer without WithPassDir → uninitialized.
	in := strings.NewReader(`{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"dominator.get_nfr","arguments":{"target_id":"api-latency-p95"}}}` + "\n")
	var out bytes.Buffer
	srv := session.NewMCPServer(in, &out, nil)

	// when
	if err := srv.Serve(context.Background()); err != nil {
		t.Fatalf("Serve: %v", err)
	}

	// then
	body := decodeFirstText(t, &out)
	if body["initialized"] != false {
		t.Errorf("initialized = %v, want false", body["initialized"])
	}
}

func TestMCPServer_GetNFR_RealImpl_LoadsConfig(t *testing.T) {
	// given: temp passDir with config.yaml containing NFR section.
	passDir := t.TempDir()
	cfgPath := filepath.Join(passDir, domain.ConfigFile)
	cfg := `target:
  url: http://localhost:8080
nfr:
  performance:
    p95_latency_ms: 200
    error_rate_percent: 1.0
  reliability:
    error_budget_percent: 5.0
  scalability:
    target_rps: 100
load:
  vus: 10
  duration: 30s
`
	if err := os.WriteFile(cfgPath, []byte(cfg), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	in := strings.NewReader(`{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"dominator.get_nfr","arguments":{"target_id":"api-latency-p95"}}}` + "\n")
	var out bytes.Buffer
	srv := session.NewMCPServer(in, &out, nil).WithPassDir(passDir)

	// when
	if err := srv.Serve(context.Background()); err != nil {
		t.Fatalf("Serve: %v", err)
	}

	// then
	body := decodeFirstText(t, &out)
	if body["initialized"] != true {
		t.Errorf("initialized = %v, want true (body=%v)", body["initialized"], body)
	}
	if body["target_id"] != "api-latency-p95" {
		t.Errorf("target_id = %v, want api-latency-p95", body["target_id"])
	}
	nfr, _ := body["nfr"].(map[string]any)
	perf, _ := nfr["performance"].(map[string]any)
	if got, _ := perf["p95_latency_ms"].(float64); int(got) != 200 {
		t.Errorf("p95_latency_ms = %v, want 200", perf["p95_latency_ms"])
	}
}

func TestMCPServer_RecordResult_RealImpl_PreviewOnly(t *testing.T) {
	// given
	in := strings.NewReader(`{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"dominator.record_result","arguments":{"target_id":"api-latency-p95","verdict":"pass","summary":"all checks green"}}}` + "\n")
	var out bytes.Buffer
	srv := session.NewMCPServer(in, &out, nil)

	// when
	if err := srv.Serve(context.Background()); err != nil {
		t.Fatalf("Serve: %v", err)
	}

	// then
	body := decodeFirstText(t, &out)
	if body["target_id"] != "api-latency-p95" {
		t.Errorf("target_id = %v, want api-latency-p95", body["target_id"])
	}
	if body["verdict"] != "pass" {
		t.Errorf("verdict = %v, want pass", body["verdict"])
	}
	if got, _ := body["summary_length"].(float64); int(got) != len("all checks green") {
		t.Errorf("summary_length = %v, want %d", body["summary_length"], len("all checks green"))
	}
	if body["persistence"] != "preview-only" {
		t.Errorf("persistence = %v, want preview-only", body["persistence"])
	}
}

func TestMCPServer_RecordResult_EmitterWired_PersistsEventStore(t *testing.T) {
	// given
	rec := &recordingJudgmentEmitter{}
	in := strings.NewReader(`{"jsonrpc":"2.0","id":9,"method":"tools/call","params":{"name":"dominator.record_result","arguments":{"target_id":"api-latency-p95","verdict":"fail","summary":"p95 exceeded"}}}` + "\n")
	var out bytes.Buffer
	srv := session.NewMCPServer(in, &out, nil).WithEmitter(rec)

	// when
	if err := srv.Serve(context.Background()); err != nil {
		t.Fatalf("Serve: %v", err)
	}

	// then
	body := decodeFirstText(t, &out)
	if body["persisted"] != true {
		t.Errorf("persisted = %v, want true", body["persisted"])
	}
	if body["persistence"] != "event-store" {
		t.Errorf("persistence = %v, want event-store", body["persistence"])
	}
	if len(rec.calls) != 1 {
		t.Fatalf("expected 1 emit call, got %d", len(rec.calls))
	}
	if rec.calls[0].targetID != "api-latency-p95" || rec.calls[0].verdict != "fail" {
		t.Errorf("emit call = %+v, want target_id=api-latency-p95 verdict=fail", rec.calls[0])
	}
}

func TestMCPServer_RecordResult_EmitterWired_RejectsInvalidWithoutEmit(t *testing.T) {
	// given: invalid verdict must NOT reach the emitter
	rec := &recordingJudgmentEmitter{}
	in := strings.NewReader(`{"jsonrpc":"2.0","id":10,"method":"tools/call","params":{"name":"dominator.record_result","arguments":{"target_id":"api","verdict":"unknown"}}}` + "\n")
	var out bytes.Buffer
	srv := session.NewMCPServer(in, &out, nil).WithEmitter(rec)

	// when
	if err := srv.Serve(context.Background()); err != nil {
		t.Fatalf("Serve: %v", err)
	}

	// then
	body := decodeFirstText(t, &out)
	if body["persisted"] != false {
		t.Errorf("persisted = %v, want false", body["persisted"])
	}
	if len(rec.calls) != 0 {
		t.Errorf("expected no emit calls for invalid verdict, got %d", len(rec.calls))
	}
}

func TestMCPServer_RecordResult_RealImpl_RejectsMissingFields(t *testing.T) {
	// given: missing verdict
	in := strings.NewReader(`{"jsonrpc":"2.0","id":8,"method":"tools/call","params":{"name":"dominator.record_result","arguments":{"target_id":"api","verdict":"unknown"}}}` + "\n")
	var out bytes.Buffer
	srv := session.NewMCPServer(in, &out, nil)

	// when
	if err := srv.Serve(context.Background()); err != nil {
		t.Fatalf("Serve: %v", err)
	}

	// then
	body := decodeFirstText(t, &out)
	if body["persisted"] != false {
		t.Errorf("persisted = %v, want false", body["persisted"])
	}
	if _, ok := body["reason"]; !ok {
		t.Errorf("reason missing: %v", body)
	}
}

// decodeFirstText extracts the JSON payload from the first content
// item of the MCP tools/call response. Stub responses ship a single
// JSON-string text entry so the body is a JSON object inside a string.
func decodeFirstText(t *testing.T, out *bytes.Buffer) map[string]any {
	t.Helper()
	var resp map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(out.Bytes()), &resp); err != nil {
		t.Fatalf("decode response: %v (raw=%q)", err, out.String())
	}
	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("missing result: %v", resp)
	}
	content, ok := result["content"].([]any)
	if !ok || len(content) == 0 {
		t.Fatalf("missing content: %v", result)
	}
	first, _ := content[0].(map[string]any)
	text, _ := first["text"].(string)
	var body map[string]any
	if err := json.Unmarshal([]byte(text), &body); err != nil {
		t.Fatalf("decode inner JSON: %v (raw=%q)", err, text)
	}
	return body
}

func TestMCPServer_RejectsUnknownMethod(t *testing.T) {
	// given
	in := strings.NewReader(`{"jsonrpc":"2.0","id":4,"method":"completion/complete"}` + "\n")
	var out bytes.Buffer
	srv := session.NewMCPServer(in, &out, nil)

	// when
	if err := srv.Serve(context.Background()); err != nil {
		t.Fatalf("Serve: %v", err)
	}

	// then
	var resp map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(out.Bytes()), &resp); err != nil {
		t.Fatalf("decode response: %v (raw=%q)", err, out.String())
	}
	rpcErr, ok := resp["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error, got %v", resp)
	}
	if code, _ := rpcErr["code"].(float64); int(code) != -32601 {
		t.Errorf("error code = %v, want -32601", rpcErr["code"])
	}
}

func TestMCPServer_Initialize_Handshake(t *testing.T) {
	// given: client sends initialize with a different protocol version
	in := strings.NewReader(`{"jsonrpc":"2.0","id":0,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"claude-code","version":"1.0"}}}` + "\n")
	var out bytes.Buffer
	srv := session.NewMCPServer(in, &out, nil)

	// when
	if err := srv.Serve(context.Background()); err != nil {
		t.Fatalf("Serve: %v", err)
	}

	// then: server returns ITS supported version (not an echo), + tools cap + serverInfo
	var resp struct {
		Result struct {
			ProtocolVersion string                     `json:"protocolVersion"`
			Capabilities    map[string]json.RawMessage `json:"capabilities"`
			ServerInfo      struct {
				Name string `json:"name"`
			} `json:"serverInfo"`
		} `json:"result"`
	}
	if err := json.Unmarshal(bytes.TrimSpace(out.Bytes()), &resp); err != nil {
		t.Fatalf("decode initialize response: %v (raw=%q)", err, out.String())
	}
	if resp.Result.ProtocolVersion != "2024-11-05" {
		t.Errorf("protocolVersion = %q, want 2024-11-05 (server supported, not echo of client 2025-06-18)", resp.Result.ProtocolVersion)
	}
	if _, ok := resp.Result.Capabilities["tools"]; !ok {
		t.Errorf("capabilities.tools missing: %v", resp.Result.Capabilities)
	}
	if resp.Result.ServerInfo.Name != "dominator" {
		t.Errorf("serverInfo.name = %q, want dominator", resp.Result.ServerInfo.Name)
	}
}

func TestMCPServer_NotificationsInitialized_NoResponse(t *testing.T) {
	// given: a JSON-RPC notification (no id)
	in := strings.NewReader(`{"jsonrpc":"2.0","method":"notifications/initialized"}` + "\n")
	var out bytes.Buffer
	srv := session.NewMCPServer(in, &out, nil)

	// when
	if err := srv.Serve(context.Background()); err != nil {
		t.Fatalf("Serve: %v", err)
	}

	// then: notifications must not produce a response
	if strings.TrimSpace(out.String()) != "" {
		t.Errorf("notification must produce no response, got: %q", out.String())
	}
}
