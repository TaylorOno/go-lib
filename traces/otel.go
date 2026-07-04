package traces

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// DefaultSampler is the default sampler used by the OTEL provider.
// It uses a parent-based sampler that defaults to always sampling if no parent span is present and is ideal for tail sampling.
var DefaultSampler = sdktrace.ParentBased(sdktrace.AlwaysSample())

// OTELProvider is an implementation of Provider that uses OpenTelemetry.
type OTELProvider struct {
	tracer       trace.Tracer
	shutdownFunc func(ctx context.Context) error
}

// Start creates a span and a context.Context containing the newly created span.
func (o *OTELProvider) Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, Span) {
	return o.tracer.Start(ctx, spanName, opts...)
}

// Shutdown shuts down the provider and releases any resources.
func (o *OTELProvider) Shutdown(ctx context.Context) {
	err := o.shutdownFunc(ctx)
	if err != nil {
		slog.Error("failed to shutdown provider", slog.String("error", err.Error()))
	}
}

// InitOTELProvider initializes the global OTEL provider.
// Call this function at the start of your application to set up OpenTelemetry tracing.
func InitOTELProvider(name string, version string, exporter sdktrace.SpanExporter, sampler sdktrace.Sampler, options ...sdktrace.TracerProviderOption) Provider {
	schemaless := resource.NewSchemaless(ServiceNameKey.String(name), ServiceVersionKey.String(version))
	options = append(options, sdktrace.WithResource(schemaless), sdktrace.WithBatcher(exporter), sdktrace.WithSampler(sampler))

	// Create a new trace provider with configured options.
	tp := sdktrace.NewTracerProvider(options...)

	// SetGlobalTracerProvider sets the global tracer provider.
	otel.SetTracerProvider(tp)

	// SetGlobalTextMapPropagator sets the global propagator used in this process.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	provider = &OTELProvider{
		tracer:       tp.Tracer(name),
		shutdownFunc: tp.Shutdown,
	}

	return provider
}

// GRPCExporter creates a new OTLP gRPC exporter that sends traces to an OTLP-compatible backend (like Jaeger or Otel Collector) via gRPC.
func GRPCExporter(ctx context.Context, options ...otlptracegrpc.Option) sdktrace.SpanExporter {
	exporter, err := otlptracegrpc.New(ctx, options...)
	if err != nil {
		slog.Error("failed to create exporter", slog.String("error", err.Error()))
		return &NoopExporter{}
	}

	return exporter
}
