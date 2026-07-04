package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/taylorono/go-lib/traces"
	"go.opentelemetry.io/otel/attribute"

	"go.opentelemetry.io/otel/trace"
)

// SpanDecorator handler signature
type SpanDecorator func(span trace.Span, req *http.Request, res *http.Response, err error)

// TracerOption tracer builder option
type TracerOption func(t *ClientTracer)

// ClientTracer tracer container
type ClientTracer struct {
	spanOpName func(*http.Request, string) string
	tracer     Tracer
}

// Trace initializes a tracer to handle tracing for outgoing HTTP requests
func Trace(clientName string, tracer Tracer, options ...TracerOption) ClientMiddleware {
	t := ClientTracer{
		spanOpName: func(req *http.Request, clientName string) string {
			return fmt.Sprintf("REST %s %s", clientName, req.Method)
		},
		tracer: tracer,
	}

	for _, option := range options {
		option(&t)
	}

	return t.handler(clientName)
}

// SpanOpName overrides the default span operation name
func SpanOpName(spanOpName func(*http.Request, string) string) TracerOption {
	return func(t *ClientTracer) {
		t.spanOpName = spanOpName
	}
}

func (t *ClientTracer) handler(clientName string) ClientMiddleware {
	return func(c Doer) Doer {
		return ClientFunc(func(req *http.Request) (*http.Response, error) {
			span := t.startSpan(req, clientName)

			res, err := c.Do(req)

			t.finishSpan(span, req, res, err)

			return res, err
		})
	}
}

func (t *ClientTracer) startSpan(req *http.Request, clientName string) traces.Span {
	spanName := t.name(req, clientName)
	_, span := t.tracer.Start(req.Context(), spanName,
		trace.WithSpanKind(trace.SpanKindClient),
		traces.AsComponent(fmt.Sprintf("%v-rest-client", clientName)),
		decorate(req),
	)

	return span
}

func decorate(req *http.Request) trace.SpanStartOption {
	attributes := []attribute.KeyValue{
		traces.HttpMethodKey.String(req.Method),
		traces.HttpURLKey.String(req.URL.String()),
		traces.PeerHostnameKey.String(req.URL.Hostname()),
	}

	port, portErr := strconv.ParseUint(req.URL.Port(), 10, 16)
	if portErr == nil {
		attributes = append(attributes, traces.PeerPortKey.Int(int(port)))
	}

	return trace.WithAttributes(attributes...)
}

func (t *ClientTracer) finishSpan(span traces.Span, req *http.Request, res *http.Response, err error) {
	if span != nil {
		if err != nil {
			span.SetAttributes(traces.ErrorKey.Bool(true))
			span.SetAttributes(traces.EventKey.String("error"))
			span.SetAttributes(traces.ErrorObjectKey.String(err.Error()))
			return
		}

		span.SetAttributes(traces.HttpStatusCodeKey.Int(res.StatusCode))
		if res.StatusCode >= 400 && res.StatusCode <= 599 {
			span.SetAttributes(traces.ErrorKey.Bool(true))
		}

		span.End()
	}
}

func (t *ClientTracer) name(req *http.Request, clientName string) string {
	var spanOpName string
	if t.spanOpName != nil {
		spanOpName = t.spanOpName(req, clientName)
	} else {
		spanOpName = req.Method
	}
	return spanOpName
}
