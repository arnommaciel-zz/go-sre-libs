package opentelemetry

import (
	"context"
	"fmt"

	"github.com/go-redis/redis"
	otelcontrib "go.opentelemetry.io/contrib"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

//const (
//	tracerKey  = "Tracing Example"
//	tracerName = "otel-redis"
//)

// Client is the interface returned by Wrap.
//
// Client implements redis.UniversalClient
type Client interface {
	redis.UniversalClient

	// ClusterClient returns the wrapped *redis.ClusterClient,
	// or nil if a non-cluster client is wrapped.
	Cluster() *redis.ClusterClient

	// Ring returns the wrapped *redis.Ring,
	// or nil if a non-ring client is wrapped.
	RingClient() *redis.Ring

	// WithContext returns a shallow copy of the client with
	// its context changed to ctx and will add instrumentation
	// with client.WrapProcess and client.WrapProcessPipeline
	//
	// To report commands as spans, ctx must contain a transaction or span.
	WithContext(ctx context.Context) Client
}

// Wrap wraps client such that executed commands are reported as spans to Elastic APM,
// using the client's associated context.
// A context-specific client may be obtained by using Client.WithContext.
func Wrap(client redis.UniversalClient) Client {
	switch client.(type) {
	case *redis.Client:
		return contextClient{Client: client.(*redis.Client)}
	case *redis.ClusterClient:
		return contextClusterClient{ClusterClient: client.(*redis.ClusterClient)}
	case *redis.Ring:
		return contextRingClient{Ring: client.(*redis.Ring)}
	}

	return client.(Client)

}

type contextClient struct {
	*redis.Client
}

// Uses the indicated context with the redis client proxy.
func (c contextClient) WithContext(ctx context.Context) Client {
	c.Client = c.Client.WithContext(ctx)
	c.WrapProcess(process(ctx))
	c.WrapProcessPipeline(processPipeline(ctx))

	return c
}

func (c contextClient) Cluster() *redis.ClusterClient {
	return nil
}

func (c contextClient) RingClient() *redis.Ring {
	return nil
}

type contextClusterClient struct {
	*redis.ClusterClient
}

func (c contextClusterClient) Cluster() *redis.ClusterClient {
	return c.ClusterClient
}

func (c contextClusterClient) RingClient() *redis.Ring {
	return nil
}

func (c contextClusterClient) WithContext(ctx context.Context) Client {
	c.ClusterClient = c.ClusterClient.WithContext(ctx)

	c.WrapProcess(process(ctx))
	c.WrapProcessPipeline(processPipeline(ctx))

	return c
}

type contextRingClient struct {
	*redis.Ring
}

//
func (c contextRingClient) Cluster() *redis.ClusterClient {

	return nil
}

//
func (c contextRingClient) RingClient() *redis.Ring {
	return c.Ring
}

//
func (c contextRingClient) WithContext(ctx context.Context) Client {
	c.Ring = c.Ring.WithContext(ctx)
	c.WrapProcess(process(ctx))
	c.WrapProcessPipeline(processPipeline(ctx))
	return c
}

// Captures a Redis process and creates a tracing span for it.
func process(ctx context.Context) func(oldProcess func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
	cfg := config{}
	if cfg.TracerProvider == nil {
		cfg.TracerProvider = otel.GetTracerProvider()
	}
	tracer := cfg.TracerProvider.Tracer(
		"otel-redis",
		trace.WithInstrumentationVersion(otelcontrib.SemVersion()),
	)
	if cfg.Propagators == nil {
		cfg.Propagators = otel.GetTextMapPropagator()
	}

	return func(oldProcess func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
		return func(cmd redis.Cmder) error {
			spanName := cmd.Name()

			attrs := []attribute.KeyValue{
				attribute.String("db.system", "redis"),
				attribute.String("db.statement", fmt.Sprintf("%s", fmt.Sprintf("%v", cmd.Args()))),
			}

			opts := []trace.SpanOption{
				trace.WithAttributes(attrs...),
				trace.WithSpanKind(trace.SpanKindClient),
			}

			_, span := tracer.Start(ctx, spanName, opts...)
			defer span.End()
			return oldProcess(cmd)
		}
	}
}

//
func processPipeline(ctx context.Context) func(oldProcess func(cmds []redis.Cmder) error) func(cmds []redis.Cmder) error {
	return func(oldProcess func(cmds []redis.Cmder) error) func(cmds []redis.Cmder) error {
		return func(cmds []redis.Cmder) error {

			//ext.DBType.Set(pipelineSpan, "redis")

			for i := len(cmds); i > 0; i-- {
				//	cmdName := strings.ToUpper(cmds[i-1].Name())
				//	if cmdName == "" {
				//		cmdName = "(empty command)"
				//	}
				//
				//	span, _ := opentracing.StartSpanFromContext(ctx, cmdName)
				//	ext.DBType.Set(span, "redis")
				//	ext.DBStatement.Set(span, fmt.Sprintf("%v", cmds[i-1].Args()))
				//	defer span.Finish()
				fmt.Println("redis process pipeline: ", cmds[i-1].Args())
			}
			//
			//defer pipelineSpan.Finish()

			return oldProcess(cmds)
		}
	}
}

// Function to start a Redis span. The returned span must be terminated at the end of the operation
// ex  :
// span := CreateRedisSpan(ctx, "Get", "key", "value")
// defer span.Finish()
func CreateRedisSpan(ctx context.Context, operation string, key string, value string) trace.Span {
	cfg := config{}
	if cfg.TracerProvider == nil {
		cfg.TracerProvider = otel.GetTracerProvider()
	}
	tracer := cfg.TracerProvider.Tracer(
		"otel-redis",
		trace.WithInstrumentationVersion(otelcontrib.SemVersion()),
	)
	if cfg.Propagators == nil {
		cfg.Propagators = otel.GetTextMapPropagator()
	}

	spanName := fmt.Sprintf("%s %s", operation, key)

	attrs := []attribute.KeyValue{
		attribute.String("db.statement", fmt.Sprintf("%s %s %s", operation, key, value)),
		attribute.String("db.system", "redis"),
	}

	opts := []trace.SpanOption{
		trace.WithAttributes(attrs...),
		trace.WithSpanKind(trace.SpanKindClient),
	}

	_, span := tracer.Start(ctx, spanName, opts...)
	return span
}
