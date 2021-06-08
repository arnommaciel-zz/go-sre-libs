package opentelemetry

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	otelcontrib "go.opentelemetry.io/contrib"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Function to start a Mongo span. The returned span must be terminated at the end of the operation
// ex  :
// span := CreateMongoDBSpan(ctx, "Get", "a,b,c", query)
// defer span.Finish()
func CreateMongoDBSpan(ctx context.Context, operation string, collection string, query bson.M) trace.Span {
	cfg := config{}
	if cfg.TracerProvider == nil {
		cfg.TracerProvider = otel.GetTracerProvider()
	}
	tracer := cfg.TracerProvider.Tracer(
		"otel-mongo",
		trace.WithInstrumentationVersion(otelcontrib.SemVersion()),
	)
	if cfg.Propagators == nil {
		cfg.Propagators = otel.GetTextMapPropagator()
	}

	spanName := fmt.Sprintf("Mongo: %s", operation)

	attrs := []attribute.KeyValue{
		attribute.String("DB Type", "MongoDB"),
		attribute.String("DB Query", fmt.Sprintf("%s", fmt.Sprintf("%v", query))),
		attribute.String("DB Values", collection),
	}

	opts := []trace.SpanOption{
		trace.WithAttributes(attrs...),
		trace.WithSpanKind(trace.SpanKindServer),
	}

	_, span := tracer.Start(ctx, spanName, opts...)
	return span
}
