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
		"otel-memcached",
		trace.WithInstrumentationVersion(otelcontrib.SemVersion()),
	)
	if cfg.Propagators == nil {
		cfg.Propagators = otel.GetTextMapPropagator()
	}

	spanName := fmt.Sprintf("%s %s", operation, key)

	attrs := []attribute.KeyValue{
		attribute.String("db.system", "memcached"),
		attribute.String("db.statement", fmt.Sprintf("%s %s %s", operation, key, value)),
	}

	opts := []trace.SpanOption{
		trace.WithAttributes(attrs...),
		trace.WithSpanKind(trace.SpanKindClient),
	}

	_, span := tracer.Start(ctx, spanName, opts...)
	return span
}
