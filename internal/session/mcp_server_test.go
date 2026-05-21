package session_test

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/hironow/dominator/internal/session"
)

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

func TestMCPServer_GetNFRStub_EchoesTargetID(t *testing.T) {
	// given
	in := strings.NewReader(`{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"dominator.get_nfr","arguments":{"target_id":"api-latency-p95"}}}` + "\n")
	var out bytes.Buffer
	srv := session.NewMCPServer(in, &out, nil)

	// when
	if err := srv.Serve(context.Background()); err != nil {
		t.Fatalf("Serve: %v", err)
	}

	// then
	body := decodeFirstText(t, &out)
	if got, _ := body["target_id"].(string); got != "api-latency-p95" {
		t.Errorf("target_id = %v, want api-latency-p95", body["target_id"])
	}
	if _, ok := body["contract"]; !ok {
		t.Errorf("contract descriptor missing: %v", body)
	}
}

func TestMCPServer_RecordResultStub_EchoesVerdict(t *testing.T) {
	// given
	in := strings.NewReader(`{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"dominator.record_result","arguments":{"target_id":"api-latency-p95","verdict":"pass","summary":"all checks green"}}}` + "\n")
	var out bytes.Buffer
	srv := session.NewMCPServer(in, &out, nil)

	// when
	if err := srv.Serve(context.Background()); err != nil {
		t.Fatalf("Serve: %v", err)
	}

	// then
	body := decodeFirstText(t, &out)
	if got, _ := body["target_id"].(string); got != "api-latency-p95" {
		t.Errorf("target_id = %v, want api-latency-p95", body["target_id"])
	}
	if got, _ := body["verdict"].(string); got != "pass" {
		t.Errorf("verdict = %v, want pass", body["verdict"])
	}
	if got, _ := body["summary_length"].(float64); int(got) != len("all checks green") {
		t.Errorf("summary_length = %v, want %d", body["summary_length"], len("all checks green"))
	}
	if got, _ := body["recorded"].(bool); got {
		t.Errorf("recorded = true, want false (stub must not persist)")
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
