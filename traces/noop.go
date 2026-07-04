package traces

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Noop is an implementation of Provider that does nothing.
type Noop struct{}

// Start returns the provided context and a NoopSpan.
func (n *Noop) Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, Span) {
	return ctx, &NoopSpan{}
}

// Shutdown does nothing.
func (n *Noop) Shutdown(ctx context.Context) {
	// Do nothing
}

// NoopSpan is an implementation of Span that does nothing.
type NoopSpan struct {
}

// SetAttributes does nothing.
func (n NoopSpan) SetAttributes(kv ...attribute.KeyValue) {
	// Do nothing
}

// End does nothing.
func (n NoopSpan) End(options ...trace.SpanEndOption) {
	// Do nothing
}

// NoopExporter is an implementation of SpanExporter that does nothing.
type NoopExporter struct {
}

// ExportSpans does nothing.
func (n NoopExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	// Do nothing
	return nil
}

// Shutdown does nothing.
func (n NoopExporter) Shutdown(ctx context.Context) error {
	// Do nothing
	return nil
}
