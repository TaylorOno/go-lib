package traces

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// HttpMiddleware returns an HTTP middleware that starts a span for each request.
// Use this to automatically trace incoming HTTP requests in your web server.
// It records the HTTP method, URL, and status code in the span.
func HttpMiddleware(tracer Provider) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx, span := tracer.Start(
				r.Context(),
				name(r),
				decorate(r),
				trace.WithSpanKind(trace.SpanKindServer))

			recorder := newResponseRecorder(w)
			next.ServeHTTP(recorder, r.WithContext(ctx))

			finishSpan(span, recorder.statusCode)
		}
	}
}

// name generates a span name for the given request.
func name(r *http.Request) string {
	var builder strings.Builder
	builder.WriteString("HTTP ")
	builder.WriteString(r.Method)
	builder.WriteString(" ")
	builder.WriteString(getPath(r))
	return builder.String()
}

// decorate returns a SpanStartOption with attributes for the given request.
func decorate(req *http.Request) trace.SpanStartOption {
	attributes := []attribute.KeyValue{
		ComponentKey.String("handler"),
		HttpMethodKey.String(req.Method),
		HttpURLKey.String(getPath(req)),
	}

	return trace.WithAttributes(attributes...)
}

// finishSpan ends the span and sets the status code and error attribute if necessary.
func finishSpan(span Span, statusCode int) {
	if span != nil {
		span.SetAttributes(HttpStatusCodeKey.Int(statusCode))
		if statusCode >= 400 && statusCode <= 599 {
			span.SetAttributes(ErrorKey.Bool(true))
		}

		span.End()
	}
}

// responseRecorder is an implementation of http.ResponseWriter that records the status code.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

// newResponseRecorder returns a new responseRecorder.
func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{w, http.StatusOK}
}

// WriteHeader records the status code and calls the underlying ResponseWriter.WriteHeader.
func (lrw *responseRecorder) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// Flush calls the underlying ResponseWriter.Flush if it supports it.
func (lrw *responseRecorder) Flush() {
	if f, ok := lrw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Hijack calls the underlying ResponseWriter.Hijack if it supports it.
func (lrw *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := lrw.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, errors.New("ResponseWriter does not support Hijacker")
}

// getPath returns the path for the given request.
func getPath(r *http.Request) string {
	path := strings.Split(r.Pattern, " ")
	return path[len(path)-1]
}
