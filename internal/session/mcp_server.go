package session

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/platform"
	"github.com/hironow/dominator/internal/usecase/port"
)

// MCPServer is a stdio-based Model Context Protocol server for the
// refs/issues/0027 jun15 MCP pivot.
//
// Three tools are exposed with real implementations: ping
// (health check), get_nfr (reads NFR thresholds from
// config.yaml), and record_result (persists an
// EventJudgmentRecorded to the event store when an emitter is wired,
// ADR 0005).
//
// Wire it into a Claude Code interactive session via --mcp-config so
// inference stays on the human-initiated session's subscription quota
// rather than crossing into the Agent SDK credit pool that gates
// `claude -p` from 2026-06-15.
//
// Protocol: JSON-RPC 2.0 over stdio, one envelope per line. Stderr
// carries human-readable diagnostics (per the project stdout/stderr
// separation invariant). Pattern follows paintress Phase 1
// (ADR 0017) + sightjack Phase 2a (ADR 0018) + amadeus Phase 2b
// (ADR 0026) + paintress Phase 3 real impl (= 83cb3ca) WithContinent
// pattern.
//
// passDir is the .pass/ state directory used to resolve config /
// event store paths. When empty, real-impl tools return uninitialized.
type MCPServer struct {
	in      io.Reader
	out     io.Writer
	logger  domain.Logger
	passDir string
	emitter port.JudgmentEventEmitter // optional: when wired, record_result persists EventJudgmentRecorded
}

// NewMCPServer wires explicit I/O so tests can drive the server
// without subprocess overhead. Passing nil for logger uses NopLogger.
func NewMCPServer(in io.Reader, out io.Writer, logger domain.Logger) *MCPServer {
	if logger == nil {
		logger = &domain.NopLogger{}
	}
	return &MCPServer{in: in, out: out, logger: logger}
}

// WithPassDir sets the .pass state directory used by real-impl MCP
// tools to resolve config / event store paths. Returns s for chaining
// (= paintress.WithContinent symmetric).
func (s *MCPServer) WithPassDir(passDir string) *MCPServer {
	s.passDir = passDir
	return s
}

// WithEmitter injects a JudgmentEventEmitter so record_result
// persists an EventJudgmentRecorded to the event store. When nil (the
// default), record_result stays preview-only. The emitter is constructed
// in the cmd composition root (= paintress.WithEmitter symmetric); the
// session layer never builds the aggregate directly
// (session-no-direct-new-aggregate semgrep rule).
func (s *MCPServer) WithEmitter(emitter port.JudgmentEventEmitter) *MCPServer {
	s.emitter = emitter
	return s
}

// jsonrpcMessage is the minimum JSON-RPC 2.0 envelope this skeleton
// understands. Method-specific params decode on demand from
// Params (json.RawMessage).
type jsonrpcMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *jsonrpcError   `json:"error,omitempty"`
}

type jsonrpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Serve reads messages from in line-by-line and writes responses to
// out until ctx cancels or stdin closes. Per-message decode errors
// surface as JSON-RPC error responses; only stream-level read errors
// abort Serve.
func (s *MCPServer) Serve(ctx context.Context) error {
	scanner := bufio.NewScanner(s.in)
	// 4 MiB buffer to comfortably cover D-Mail bodies in later commits.
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		if err := s.handle(ctx, line); err != nil {
			s.logger.Warn("mcp server: handle: %v", err)
		}
	}
	if err := scanner.Err(); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("mcp server: read stdin: %w", err)
	}
	return nil
}

func (s *MCPServer) handle(ctx context.Context, line []byte) error {
	var msg jsonrpcMessage
	if err := json.Unmarshal(line, &msg); err != nil {
		return fmt.Errorf("decode request: %w", err)
	}
	switch msg.Method {
	case "initialize":
		return s.respond(msg.ID, initializeResult())
	case "notifications/initialized":
		// JSON-RPC notification (no id): the client signals it finished
		// the handshake. No response is sent.
		return nil
	case "tools/list":
		return s.respond(msg.ID, map[string]any{"tools": toolDescriptors()})
	case "tools/call":
		return s.handleToolsCall(ctx, msg)
	default:
		// Unknown notifications (no id) are ignored per JSON-RPC; only
		// id-bearing requests get a method-not-found error.
		if len(msg.ID) == 0 {
			return nil
		}
		return s.respondError(msg.ID, -32601, fmt.Sprintf("method not implemented: %s", msg.Method))
	}
}

// mcpProtocolVersion is the single MCP protocol version this server
// implements. Per the MCP lifecycle spec, the server returns the
// version it actually supports (not an echo of the client's request):
// echoing an unsupported client version would falsely claim support
// and break future / draft clients. The client decides compatibility
// from this value.
const mcpProtocolVersion = "2024-11-05"

// initializeResult builds the MCP initialize handshake response. The
// Claude Code session sends `initialize` first; without a valid reply
// it never proceeds to tools/list. The server advertises its supported
// protocol version + the tools capability.
func initializeResult() map[string]any {
	return map[string]any{
		"protocolVersion": mcpProtocolVersion,
		"capabilities":    map[string]any{"tools": map[string]any{"listChanged": false}},
		"serverInfo":      map[string]any{"name": "dominator", "version": "0.1.0"},
	}
}

// handleToolsCall dispatches a single tools/call request and records
// MCP invocation metrics (mcp.tool.invocations counter +
// mcp.tool.duration histogram) for cost-monitoring verification post
// 2026-06-15 (refs/issues/0027 Phase 3 cost monitoring (a)).
func (s *MCPServer) handleToolsCall(ctx context.Context, msg jsonrpcMessage) error {
	start := time.Now()
	var call struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(msg.Params, &call); err != nil {
		platform.RecordMCPInvocation(ctx, "", "error", time.Since(start))
		return s.respondError(msg.ID, -32602, "invalid tools/call params")
	}

	status := "ok"
	var result map[string]any
	switch call.Name {
	case "ping":
		result = textResult("pong")
	case "get_nfr":
		result = realGetNFR(s.passDir, call.Arguments)
	case "record_result":
		result = realRecordResult(s.emitter, call.Arguments)
	default:
		platform.RecordMCPInvocation(ctx, call.Name, "error", time.Since(start))
		return s.respondError(msg.ID, -32601, fmt.Sprintf("unknown tool: %s", call.Name))
	}

	err := s.respond(msg.ID, result)
	if err != nil {
		status = "error"
	}
	platform.RecordMCPInvocation(ctx, call.Name, status, time.Since(start))
	return err
}

// toolDescriptors returns the tool set. Each entry pins the interface
// (name, description, inputSchema) so Claude Code clients see a stable
// contract. All three handler bodies are real implementations
// (ping / realGetNFR / realRecordResult).
func toolDescriptors() []map[string]any {
	return []map[string]any{
		{
			"name":        "ping",
			"description": "Health check. Returns 'pong'.",
			"inputSchema": map[string]any{"type": "object", "properties": map[string]any{}},
		},
		{
			"name":        "get_nfr",
			"description": "Return the NFR thresholds (latency / error rate / success rate / RPS) for the given target, read from the .pass/config.yaml NFR config.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"target_id": map[string]any{"type": "string", "description": "NFR target identifier (e.g. 'api-latency-p95')"},
				},
				"required": []any{"target_id"},
			},
		},
		{
			"name":        "record_result",
			"description": "Persist a k6 run result (verdict 'pass'/'fail' + summary) against the given NFR target as an EventJudgmentRecorded event (persistence='event-store' when an emitter is wired; preview-only otherwise).",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"target_id": map[string]any{"type": "string"},
					"verdict":   map[string]any{"type": "string", "description": "pass / fail"},
					"summary":   map[string]any{"type": "string", "description": "human-readable summary of the run"},
				},
				"required": []any{"target_id", "verdict"},
			},
		},
	}
}

// textResult wraps a plain string into the MCP content envelope.
func textResult(text string) map[string]any {
	return map[string]any{"content": []map[string]any{{"type": "text", "text": text}}}
}

// jsonResult marshals data as JSON and returns an MCP content envelope.
func jsonResult(data any) map[string]any {
	body, err := json.Marshal(data)
	if err != nil {
		return textResult(fmt.Sprintf(`{"error":"marshal failed: %v"}`, err))
	}
	return map[string]any{"content": []map[string]any{{"type": "text", "text": string(body)}}}
}

// realGetNFR reads .pass/config.yaml and returns the NFR thresholds
// (= Performance / Reliability / Scalability sections). target_id is
// echoed back for caller correlation; the actual NFR config is
// global per-repo (= no per-target indexing in current impl).
//
// Pattern: paintress.next_issue (= 83cb3ca) symmetric copy.
func realGetNFR(passDir string, args json.RawMessage) map[string]any {
	var payload struct {
		TargetID string `json:"target_id"`
	}
	if len(args) > 0 {
		_ = json.Unmarshal(args, &payload)
	}
	if passDir == "" {
		return jsonResult(map[string]any{
			"initialized": false,
			"reason":      "dominator mcp passDir not configured (start `dominator mcp` from the project root)",
			"target_id":   payload.TargetID,
		})
	}
	cfgPath := filepath.Join(passDir, domain.ConfigFile)
	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		return jsonResult(map[string]any{
			"initialized": false,
			"reason":      fmt.Sprintf("config load failed: %v", err),
			"target_id":   payload.TargetID,
		})
	}
	return jsonResult(map[string]any{
		"initialized": true,
		"passDir":     passDir,
		"target_id":   payload.TargetID,
		"nfr": map[string]any{
			"performance": map[string]any{
				"p95_latency_ms":     cfg.Nfr.Performance.P95LatencyMs,     // nosemgrep: lod-excessive-dot-chain -- domain.NfrConfig is a JSON wire-format DTO; intermediate accessor would defeat the YAML binding [permanent]
				"error_rate_percent": cfg.Nfr.Performance.ErrorRatePercent, // nosemgrep: lod-excessive-dot-chain -- domain.NfrConfig is a JSON wire-format DTO [permanent]
			},
			"reliability": cfg.Nfr.Reliability,
			"scalability": cfg.Nfr.Scalability,
		},
		"target": map[string]any{
			"url": cfg.Target.URL,
		},
		"instruction": "Run k6 via mcp-k6, compare results against these thresholds, then call record_result with verdict='pass' or 'fail'.",
	})
}

// realRecordResult validates the input and, when an emitter is wired,
// persists an EventJudgmentRecorded to the event store. When emitter is
// nil (event store unavailable), it falls back to a preview-only payload
// for backward compatibility.
//
// LLM firing remains human-initiated: this path only persists an event;
// the emitter never invokes the model. The MCP tool call itself is
// driven by a human-initiated Claude Code session.
//
// Pattern: paintress.append_journal (Phase 4 #4) emitter-injection
// symmetric copy.
func realRecordResult(emitter port.JudgmentEventEmitter, args json.RawMessage) map[string]any {
	var payload struct {
		TargetID string `json:"target_id"`
		Verdict  string `json:"verdict"`
		Summary  string `json:"summary"`
	}
	if len(args) > 0 {
		_ = json.Unmarshal(args, &payload)
	}
	if payload.TargetID == "" || (payload.Verdict != "pass" && payload.Verdict != "fail") {
		return jsonResult(map[string]any{
			"persisted": false,
			"reason":    "missing required fields: target_id and verdict ('pass' or 'fail')",
			"received":  payload,
		})
	}
	if emitter == nil {
		return jsonResult(map[string]any{
			"persisted":      false,
			"target_id":      payload.TargetID,
			"verdict":        payload.Verdict,
			"summary_length": len(payload.Summary),
			"recorded":       false,
			"persistence":    "preview-only",
			"note":           "Preview only (event store unavailable). Run from a directory with a writable .pass/ state dir to persist EventJudgmentRecorded.",
		})
	}
	if err := emitter.EmitJudgmentRecorded(payload.TargetID, payload.Verdict, payload.Summary, time.Now().UTC()); err != nil {
		return jsonResult(map[string]any{
			"persisted": false,
			"target_id": payload.TargetID,
			"verdict":   payload.Verdict,
			"reason":    err.Error(),
		})
	}
	return jsonResult(map[string]any{
		"persisted":      true,
		"target_id":      payload.TargetID,
		"verdict":        payload.Verdict,
		"summary_length": len(payload.Summary),
		"recorded":       true,
		"persistence":    "event-store",
	})
}

func (s *MCPServer) respond(id json.RawMessage, result any) error {
	return s.writeMessage(jsonrpcMessage{JSONRPC: "2.0", ID: id, Result: result})
}

func (s *MCPServer) respondError(id json.RawMessage, code int, message string) error {
	return s.writeMessage(jsonrpcMessage{JSONRPC: "2.0", ID: id, Error: &jsonrpcError{Code: code, Message: message}})
}

func (s *MCPServer) writeMessage(msg jsonrpcMessage) error {
	out, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("encode response: %w", err)
	}
	if _, err := s.out.Write(append(out, '\n')); err != nil {
		return fmt.Errorf("write response: %w", err)
	}
	return nil
}
