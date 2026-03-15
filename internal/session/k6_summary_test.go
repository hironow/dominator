package session_test

import (
	"math"
	"testing"

	"github.com/hironow/dominator/internal/session"
)

func TestParseK6Summary_ValidJSON(t *testing.T) {
	// given
	data := []byte(`{
		"metrics": {
			"http_req_duration": { "values": { "p(95)": 123.45, "avg": 50.0 } },
			"http_req_failed": { "values": { "rate": 0.01 } },
			"http_reqs": { "values": { "rate": 500.0, "count": 15000 } }
		}
	}`)

	// when
	results, err := session.ParseK6Summary(data)

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(results.P95LatencyMs-123.45) > 0.001 {
		t.Errorf("P95LatencyMs = %f, want 123.45", results.P95LatencyMs)
	}
	if math.Abs(results.ErrorRatePercent-1.0) > 0.001 {
		t.Errorf("ErrorRatePercent = %f, want 1.0", results.ErrorRatePercent)
	}
	if math.Abs(results.SuccessRate-99.0) > 0.001 {
		t.Errorf("SuccessRate = %f, want 99.0", results.SuccessRate)
	}
	if math.Abs(results.ActualRPS-500.0) > 0.001 {
		t.Errorf("ActualRPS = %f, want 500.0", results.ActualRPS)
	}
}

func TestParseK6Summary_MissingMetrics(t *testing.T) {
	// given: JSON with no recognized metrics
	data := []byte(`{
		"metrics": {
			"custom_metric": { "values": { "avg": 42.0 } }
		}
	}`)

	// when
	results, err := session.ParseK6Summary(data)

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// All should be zero defaults
	if results.P95LatencyMs != 0 {
		t.Errorf("P95LatencyMs = %f, want 0", results.P95LatencyMs)
	}
	if results.ErrorRatePercent != 0 {
		t.Errorf("ErrorRatePercent = %f, want 0", results.ErrorRatePercent)
	}
	if results.SuccessRate != 0 {
		t.Errorf("SuccessRate = %f, want 0 (unmeasured when http_req_failed absent)", results.SuccessRate)
	}
	if results.ActualRPS != 0 {
		t.Errorf("ActualRPS = %f, want 0", results.ActualRPS)
	}
}

func TestParseK6Summary_InvalidJSON(t *testing.T) {
	// given
	data := []byte(`not valid json`)

	// when
	_, err := session.ParseK6Summary(data)

	// then
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestParseK6Summary_EmptyMetrics(t *testing.T) {
	// given
	data := []byte(`{"metrics": {}}`)

	// when
	results, err := session.ParseK6Summary(data)

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results.SuccessRate != 0 {
		t.Errorf("SuccessRate = %f, want 0 (unmeasured — no http_req_failed metric)", results.SuccessRate)
	}
}

func TestParseK6Summary_FullResponse(t *testing.T) {
	// given: complete k6 summary with all standard HTTP metrics
	data := []byte(`{
		"metrics": {
			"http_req_duration": { "values": { "p(95)": 250.5, "p(90)": 200.0, "avg": 150.0, "min": 10.0, "max": 800.0 } },
			"http_req_failed": { "values": { "rate": 0.02, "passes": 980, "fails": 20 } },
			"http_reqs": { "values": { "rate": 1000.0, "count": 30000 } },
			"http_req_connecting": { "values": { "avg": 5.0 } },
			"http_req_tls_handshaking": { "values": { "avg": 10.0 } }
		}
	}`)

	// when
	results, err := session.ParseK6Summary(data)

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(results.P95LatencyMs-250.5) > 0.001 {
		t.Errorf("P95LatencyMs = %f, want 250.5", results.P95LatencyMs)
	}
	if math.Abs(results.ErrorRatePercent-2.0) > 0.001 {
		t.Errorf("ErrorRatePercent = %f, want 2.0", results.ErrorRatePercent)
	}
	if math.Abs(results.SuccessRate-98.0) > 0.001 {
		t.Errorf("SuccessRate = %f, want 98.0", results.SuccessRate)
	}
	if math.Abs(results.ActualRPS-1000.0) > 0.001 {
		t.Errorf("ActualRPS = %f, want 1000.0", results.ActualRPS)
	}
}

func TestParseK6Summary_HighLatency(t *testing.T) {
	// given
	data := []byte(`{
		"metrics": {
			"http_req_duration": { "values": { "p(95)": 5000.0 } }
		}
	}`)

	// when
	results, err := session.ParseK6Summary(data)

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(results.P95LatencyMs-5000.0) > 0.001 {
		t.Errorf("P95LatencyMs = %f, want 5000.0", results.P95LatencyMs)
	}
}

func TestParseK6Summary_HighErrorRate(t *testing.T) {
	// given: 50% error rate
	data := []byte(`{
		"metrics": {
			"http_req_failed": { "values": { "rate": 0.5 } }
		}
	}`)

	// when
	results, err := session.ParseK6Summary(data)

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(results.ErrorRatePercent-50.0) > 0.001 {
		t.Errorf("ErrorRatePercent = %f, want 50.0", results.ErrorRatePercent)
	}
	if math.Abs(results.SuccessRate-50.0) > 0.001 {
		t.Errorf("SuccessRate = %f, want 50.0", results.SuccessRate)
	}
}

func TestParseK6Summary_ZeroRPS(t *testing.T) {
	// given
	data := []byte(`{
		"metrics": {
			"http_reqs": { "values": { "rate": 0, "count": 0 } }
		}
	}`)

	// when
	results, err := session.ParseK6Summary(data)

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results.ActualRPS != 0 {
		t.Errorf("ActualRPS = %f, want 0", results.ActualRPS)
	}
}

func TestParseK6Summary_WebSocketMetrics(t *testing.T) {
	// given: WebSocket-specific metrics without http_req_*
	data := []byte(`{
		"metrics": {
			"ws_connecting": { "values": { "avg": 15.0 } },
			"ws_msgs_received": { "values": { "count": 5000, "rate": 250 } },
			"ws_msgs_sent": { "values": { "count": 5000, "rate": 250 } }
		}
	}`)

	// when
	results, err := session.ParseK6Summary(data)

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// All HTTP metrics should be zero since only WS metrics present
	if results.P95LatencyMs != 0 {
		t.Errorf("P95LatencyMs = %f, want 0 (no http metrics)", results.P95LatencyMs)
	}
	if results.ErrorRatePercent != 0 {
		t.Errorf("ErrorRatePercent = %f, want 0", results.ErrorRatePercent)
	}
	if results.SuccessRate != 0 {
		t.Errorf("SuccessRate = %f, want 0 (unmeasured)", results.SuccessRate)
	}
	if results.ActualRPS != 0 {
		t.Errorf("ActualRPS = %f, want 0", results.ActualRPS)
	}
}

func TestParseK6Summary_ZeroErrorRate(t *testing.T) {
	// given
	data := []byte(`{
		"metrics": {
			"http_req_failed": { "values": { "rate": 0 } }
		}
	}`)

	// when
	results, err := session.ParseK6Summary(data)

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results.ErrorRatePercent != 0 {
		t.Errorf("ErrorRatePercent = %f, want 0", results.ErrorRatePercent)
	}
	if results.SuccessRate != 100 {
		t.Errorf("SuccessRate = %f, want 100", results.SuccessRate)
	}
}

func TestParseK6Summary_Parameterized(t *testing.T) {
	tests := []struct {
		name            string
		json            string
		wantP95         float64
		wantErrorRate   float64
		wantSuccessRate float64
		wantRPS         float64
	}{
		{
			name:            "all_zeros",
			json:            `{"metrics": {"http_req_duration": {"values": {"p(95)": 0}}, "http_req_failed": {"values": {"rate": 0}}, "http_reqs": {"values": {"rate": 0}}}}`,
			wantP95:         0,
			wantErrorRate:   0,
			wantSuccessRate: 100,
			wantRPS:         0,
		},
		{
			name:            "high_performance",
			json:            `{"metrics": {"http_req_duration": {"values": {"p(95)": 5.5}}, "http_req_failed": {"values": {"rate": 0.001}}, "http_reqs": {"values": {"rate": 10000}}}}`,
			wantP95:         5.5,
			wantErrorRate:   0.1,
			wantSuccessRate: 99.9,
			wantRPS:         10000,
		},
		{
			name:            "degraded",
			json:            `{"metrics": {"http_req_duration": {"values": {"p(95)": 3000}}, "http_req_failed": {"values": {"rate": 0.15}}, "http_reqs": {"values": {"rate": 50}}}}`,
			wantP95:         3000,
			wantErrorRate:   15.0,
			wantSuccessRate: 85.0,
			wantRPS:         50,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := session.ParseK6Summary([]byte(tt.json))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if math.Abs(results.P95LatencyMs-tt.wantP95) > 0.1 {
				t.Errorf("P95LatencyMs = %f, want %f", results.P95LatencyMs, tt.wantP95)
			}
			if math.Abs(results.ErrorRatePercent-tt.wantErrorRate) > 0.1 {
				t.Errorf("ErrorRatePercent = %f, want %f", results.ErrorRatePercent, tt.wantErrorRate)
			}
			if math.Abs(results.SuccessRate-tt.wantSuccessRate) > 0.1 {
				t.Errorf("SuccessRate = %f, want %f", results.SuccessRate, tt.wantSuccessRate)
			}
			if math.Abs(results.ActualRPS-tt.wantRPS) > 0.1 {
				t.Errorf("ActualRPS = %f, want %f", results.ActualRPS, tt.wantRPS)
			}
		})
	}
}

func TestParseK6Summary_EmptyJSON(t *testing.T) {
	data := []byte(`{}`)
	results, err := session.ParseK6Summary(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results.P95LatencyMs != 0 || results.ActualRPS != 0 || results.ErrorRatePercent != 0 {
		t.Error("expected all zeros for empty JSON")
	}
}
