package platform

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// RecordCheck increments the dominator.check.total OTel counter.
func RecordCheck(ctx context.Context, status string) {
	c, _ := Meter.Int64Counter("dominator.check.total", // nosemgrep: error-handling.ignored-error-go,error-handling.ignored-error-short-go -- OTel SDK guarantees Int64Counter returns a functional noop instrument on error; Add() is always safe [permanent]
		metric.WithDescription("Total check completions"),
	)
	c.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("status", SanitizeUTF8(status)),
		),
	)
}
