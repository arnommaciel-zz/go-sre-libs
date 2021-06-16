package opentelemetry

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"

	"go.opentelemetry.io/contrib/propagators/b3"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpgrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

//Tracing
var (
	serviceVersion = os.Getenv("OTEL_SERVICE_VERSION")
	lsToken        = os.Getenv("OTEL_ACCESS_TOKEN")
)

func initExporter(url string, token string) *otlp.Exporter {

	exporter, err := otlp.NewExporter(
		context.Background(),
		otlpgrpc.NewDriver(
			otlpgrpc.WithInsecure(),
			otlpgrpc.WithEndpoint(url),
		),
	)

	if err != nil {
		log.Fatal(err)
	}
	return exporter
}

type TraceConfig struct {
	ServiceName   string
	OtelCollector string
}

type Span struct {
	span oteltrace.Span
}

// TRACING
func InitTracer(args TraceConfig) {
	b3 := b3.B3{}
	// Register the B3 propagator globally.
	otel.SetTextMapPropagator(b3)
	SamplerRatio, err := strconv.ParseFloat(os.Getenv("OTEL_SAMPLER_RATIO"), 64)
	if err != nil {
		SamplerRatio = 0.25
		log.Println("Sampler ratio value provided invalid, using default value: 0.25")
	}

	if len(serviceVersion) == 0 {
		serviceVersion = "0.0.0"
	}

	exporter := initExporter(args.OtelCollector, lsToken)

	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", args.ServiceName),
			attribute.String("service.version", serviceVersion),
			attribute.String("library.language", "go"),
		),
	)
	if err != nil {
		log.Fatal("Could not set resources: ", err)
	}
	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.ParentBased(trace.TraceIDRatioBased(SamplerRatio))),
		trace.WithSyncer(exporter),
		trace.WithResource(resources),
	)
	otel.SetTracerProvider(tp)
}

func InjectTrace(req *http.Request, newRequest *http.Request) {
	propagator := otel.GetTextMapPropagator()
	ctx := propagator.Extract(req.Context(), propagation.HeaderCarrier(req.Header))
	propagator.Inject(ctx, propagation.HeaderCarrier(newRequest.Header))
}

func StartSpanWithContext(ctx context.Context, spanName string, attrs ...attribute.KeyValue) (context.Context, Span) {
	provider := otel.GetTracerProvider()
	tracer := provider.Tracer("otel-span")

	opts := []oteltrace.SpanOption{
		oteltrace.WithAttributes(attrs...),
		oteltrace.WithSpanKind(oteltrace.SpanKindServer),
	}

	ctx, span := tracer.Start(ctx, spanName, opts...)
	return ctx, Span{
		span: span,
	}
}

func (s *Span) End() {
	if s == nil || s.span == nil {
		return
	}

	defer s.span.End()

	if recovered := recover(); recovered != nil {
		// Record but don't stop the panic.
		defer panic(recovered)
		s.span.SetStatus(codes.Error, "critical error")
	}
}

func RecordError(span Span, erro error) {
	span.span.RecordError(erro)
	span.span.SetStatus(codes.Error, "critical error")
}
