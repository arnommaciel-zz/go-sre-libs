package opentelemetry

import (
	"context"

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

	spanName := operation

	attrs := []attribute.KeyValue{
		attribute.String("db.mongodb.collection", collection),
		attribute.String("db.operation", operation),
		attribute.String("db.system", "mongodb"),
	}

	opts := []trace.SpanOption{
		trace.WithAttributes(attrs...),
		trace.WithSpanKind(trace.SpanKindClient),
	}

	_, span := tracer.Start(ctx, spanName, opts...)
	return span
}
