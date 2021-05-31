package opentelemetry

import (
	"context"
	"log"
	"net/http"
	"os"

	"go.opentelemetry.io/contrib/propagators/b3"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpgrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

//Tracing
var (
	serviceVersion = os.Getenv("OTEL_SERVICE_VERSION")
	lsToken        = os.Getenv("OTEL_ACCESS_TOKEN")
)

func initExporter(url string, token string) *otlp.Exporter {
	//	headers := map[string]string{
	//		"lightstep-access-token": token,
	//	}
	//
	exporter, err := otlp.NewExporter(
		context.Background(),
		otlpgrpc.NewDriver(
			otlpgrpc.WithInsecure(),
			otlpgrpc.WithEndpoint(url),
			//			otlpgrpc.WithHeaders(headers),
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

// TRACING
func InitTracer(args TraceConfig) {
	b3 := b3.B3{}
	// Register the B3 propagator globally.
	otel.SetTextMapPropagator(b3)

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
			attribute.String("library.version", "1.15.5"),
		),
	)
	if err != nil {
		log.Fatal("Could not set resources: ", err)
	}
	tp := trace.NewTracerProvider(
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
