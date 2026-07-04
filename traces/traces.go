package traces

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	// ServiceNameKey is the attribute key for the name of the service.
	// Use this key to identify the service in trace spans.
	ServiceNameKey = attribute.Key("service.name")

	// ServiceVersionKey is the attribute key for the version of the service.
	// Use this key to identify the version of the service in trace spans.
	ServiceVersionKey = attribute.Key("service.version")

	// ComponentKey is the attribute key for the component that generated the span.
	// Use this key to specify which part of the system (e.g., "database", "http-client") created the span.
	ComponentKey = attribute.Key("component")

	// HttpMethodKey is the attribute key for the HTTP method of the request.
	// Use this key in spans representing HTTP requests to record the method (e.g., "GET", "POST").
	HttpMethodKey = attribute.Key("http.method")

	// HttpURLKey is the attribute key for the URL of the HTTP request.
	// Use this key in spans representing HTTP requests to record the request path or full URL.
	HttpURLKey = attribute.Key("http.url")

	// HttpStatusCodeKey is the attribute key for the HTTP status code of the response.
	// Use this key in spans representing HTTP requests to record the status code of the response.
	HttpStatusCodeKey = attribute.Key("http.status_code")

	// ErrorKey is the attribute key for the error flag of the span.
	// Set this to true if the span encountered an error during its execution.
	ErrorKey = attribute.Key("error")

	// EventKey is the attribute key for an event associated with the span.
	// Use this to record specific events or milestones within a span.
	EventKey = attribute.Key("event")

	// ErrorObjectKey is the attribute key for the error object or message.
	// Use this to record the details of an error that occurred.
	ErrorObjectKey = attribute.Key("error.object")

	// PeerHostnameKey is the attribute key for the hostname of the peer.
	// Use this in client spans to record the hostname of the remote service being called.
	PeerHostnameKey = attribute.Key("peer.hostname")

	// PeerPortKey is the attribute key for the port of the peer.
	// Use this in client spans to record the port of the remote service being called.
	PeerPortKey = attribute.Key("peer.port")
)

var provider Provider = &Noop{}

// Provider is the interface that wraps the Start and Shutdown methods.
type Provider interface {
	// Start creates a span and a context.Context containing the newly created span.
	// The returned context should be used for any downstream calls to propagate the trace.
	Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, Span)
	// Shutdown shuts down the provider and releases any resources and ensure all traces are flushed.
	Shutdown(ctx context.Context)
}

// Span represents a single operation within a trace.
type Span interface {
	// SetAttributes sets attributes to the span.
	SetAttributes(kv ...attribute.KeyValue)
	// End completes the span when the operation represented by the span is finished.
	End(options ...trace.SpanEndOption)
}

// GetProvider returns the current global provider.
func GetProvider() Provider {
	return provider
}

// Start creates a span and a context.Context containing the newly created span using the global provider.
func Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, Span) {
	return provider.Start(ctx, spanName, opts...)
}

// AsComponent returns a SpanStartOption that sets the component attribute.
func AsComponent(component string) trace.SpanStartOption {
	return trace.WithAttributes(ComponentKey.String(component))
}
