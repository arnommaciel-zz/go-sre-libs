package opentelemetry

import (
	"net/http"

	"github.com/urfave/negroni"
	otelcontrib "go.opentelemetry.io/contrib"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/semconv"
	oteltrace "go.opentelemetry.io/otel/trace"
	trace "go.opentelemetry.io/otel/trace"
)

const (
	tracerKey  = "Tracing Example"
	tracerName = "otelnegroni"
)

//// Middleware returns negroni.Handler middleware which will trace incoming requests.
func NegroniMiddleware(service string) negroni.Handler {
	cfg := config{}
	if cfg.TracerProvider == nil {
		cfg.TracerProvider = otel.GetTracerProvider()
	}
	tracer := cfg.TracerProvider.Tracer(
		tracerName,
		trace.WithInstrumentationVersion(otelcontrib.SemVersion()),
	)
	if cfg.Propagators == nil {
		cfg.Propagators = otel.GetTextMapPropagator()
	}

	return negroni.HandlerFunc(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		ctx := cfg.Propagators.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		spanName := ""
		spanName = r.URL.Path
		opts := []oteltrace.SpanOption{
			oteltrace.WithAttributes(semconv.NetAttributesFromHTTPRequest("tcp", r)...),
			oteltrace.WithAttributes(semconv.EndUserAttributesFromHTTPRequest(r)...),
			oteltrace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest(service, spanName, r)...),
			oteltrace.WithSpanKind(oteltrace.SpanKindServer),
		}
		ctx, span := tracer.Start(ctx, spanName, opts...)
		defer span.End()

		cfg.Propagators.Inject(ctx, propagation.HeaderCarrier(r.Header))

		lrw := negroni.NewResponseWriter(w)
		next.ServeHTTP(lrw, r)
		statusCode := lrw.Status()
		attrs := semconv.HTTPAttributesFromHTTPStatusCode(statusCode)
		spanStatus, spanMessage := semconv.SpanStatusFromHTTPStatusCode(statusCode)
		span.SetAttributes(attrs...)
		span.SetStatus(spanStatus, spanMessage)
	})
}
