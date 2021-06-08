package opentelemetry

import (
	"context"
	"fmt"

	otelcontrib "go.opentelemetry.io/contrib"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Function to start a MemCached span. The returned span must be terminated at the end of the operation
// ex  :
// span := CreateMemCachedSpan(ctx, "Get", "key", "value")
// defer span.Finish()
func CreateMemCachedSpan(ctx context.Context, operation string, key string, value string) trace.Span {
	cfg := config{}
	if cfg.TracerProvider == nil {
		cfg.TracerProvider = otel.GetTracerProvider()
	}
	tracer := cfg.TracerProvider.Tracer(
		"otel-MemCached",
		trace.WithInstrumentationVersion(otelcontrib.SemVersion()),
	)
	if cfg.Propagators == nil {
		cfg.Propagators = otel.GetTextMapPropagator()
	}

	spanName := fmt.Sprintf("MemCached: %s", operation)

	attrs := []attribute.KeyValue{
		attribute.String("DB Type", "MemCached"),
		attribute.String("DB key", key),
		attribute.String("DB Values", value),
	}

	opts := []trace.SpanOption{
		trace.WithAttributes(attrs...),
		trace.WithSpanKind(trace.SpanKindServer),
	}

	_, span := tracer.Start(ctx, spanName, opts...)
	return span
}
