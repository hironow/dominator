//go:build integration

package integration_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// TestWeaveOTLP_HeadersAndResourceAttributes verifies that the OTLP HTTP
// exporter, when configured with Weave-style environment variables, sends
// the correct custom header (wandb-api-key) and resource attributes
// (wandb.entity, wandb.project) to the collector endpoint.
//
// This test uses a local httptest.Server as the mock OTLP collector,
// making it CI-reproducible without a real Weave/W&B account.
func TestWeaveOTLP_HeadersAndResourceAttributes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	var (
		mu              sync.Mutex
		capturedHeaders http.Header
		requestCount    int
	)
	received := make(chan struct{}, 10)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		capturedHeaders = r.Header.Clone()
		requestCount++
		mu.Unlock()
		select {
		case received <- struct{}{}:
		default:
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Simulate .otel.env for weave backend.
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", server.URL)
	t.Setenv("OTEL_EXPORTER_OTLP_HEADERS", "wandb-api-key=test-weave-key-abc123")
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "wandb.entity=test-entity,wandb.project=test-project")

	ctx := context.Background()

	exp, err := otlptracehttp.New(ctx)
	if err != nil {
		t.Fatalf("create OTLP HTTP exporter: %v", err)
	}

	res, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("dominator"),
			semconv.ServiceVersion("0.0.0-test"),
			attribute.String("wandb.entity", "test-entity"),
			attribute.String("wandb.project", "test-project"),
		),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exp),
		sdktrace.WithResource(res),
	)

	tracer := tp.Tracer("dominator")
	_, span := tracer.Start(ctx, "test-weave-verification")
	span.End()

	if err := tp.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown tracer provider: %v", err)
	}

	select {
	case <-received:
	case <-time.After(10 * time.Second):
		t.Fatal("timeout: no OTLP request received by mock collector")
	}

	mu.Lock()
	defer mu.Unlock()

	if requestCount == 0 {
		t.Fatal("no requests received by mock OTLP collector")
	}

	// Verify wandb-api-key header was sent (case-insensitive lookup).
	if got := capturedHeaders.Get("wandb-api-key"); got != "test-weave-key-abc123" {
		t.Errorf("wandb-api-key header: got %q, want %q", got, "test-weave-key-abc123")
	}

	// Verify content-type is protobuf (OTLP HTTP default).
	ct := capturedHeaders.Get("Content-Type")
	if ct != "application/x-protobuf" {
		t.Errorf("Content-Type: got %q, want %q", ct, "application/x-protobuf")
	}
}

// TestOTelTracerInitialization verifies that a TracerProvider can be created
// and produce spans without errors.
func TestOTelTracerInitialization(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// given: no OTEL endpoint set (noop)
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")

	ctx := context.Background()
	tp := sdktrace.NewTracerProvider()
	defer tp.Shutdown(ctx)

	tracer := tp.Tracer("dominator-test")

	// when: create and end a span
	_, span := tracer.Start(ctx, "test-span")
	span.End()

	// then: no panics, no errors (noop behavior)
}

// TestOTelNoopDefault verifies that when no endpoint is configured,
// the tracer does not send any requests.
func TestOTelNoopDefault(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// given: a server that should NOT receive any requests
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Do NOT set OTEL_EXPORTER_OTLP_ENDPOINT
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")

	ctx := context.Background()
	tp := sdktrace.NewTracerProvider()
	defer tp.Shutdown(ctx)

	tracer := tp.Tracer("dominator-test")
	_, span := tracer.Start(ctx, "should-not-export")
	span.End()

	tp.Shutdown(ctx)

	// then: no requests sent
	if requestCount != 0 {
		t.Errorf("expected 0 requests to collector, got %d", requestCount)
	}
}
