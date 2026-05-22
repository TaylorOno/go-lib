package rest

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

const (
	metricNameHistory     = "downstream_request_histogram"
	metricNameSummary     = "downstream_request_summary"
	metricClientError     = "downstream_client_error"
	metricConcurrentCalls = "http_current_concurrent_calls"

	errTimeoutLabelValue            = "http_request_timeout"
	errUnhandledHTTPErrorLabelValue = "http_error_unhandled"

	defaultHTTPThrottling = 200
)

var (
	defaultBuckets = []float64{1, 3, 5, 10, 25, 50, 100, 200, 400, 600, 1000, 1500, 2000, 2500, 3000, 4000, 5000, 10000}
)

type ClientBuilder struct {
	clientName      string
	baseClient      *http.Client
	transport       *http.Transport
	clientTimeout   time.Duration
	ThrottleLimit   int
	metricsReporter MetricsReporter
	middlewares     []ClientMiddleware
}

func NewClientBuilder(clientName string) *ClientBuilder {
	return &ClientBuilder{
		clientName:    clientName,
		clientTimeout: 30 * time.Second,
		baseClient:    http.DefaultClient,
		transport: &http.Transport{
			MaxIdleConns:    5,
			IdleConnTimeout: 30 * time.Second,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			MaxConnsPerHost:       100,
			MaxIdleConnsPerHost:   2,
			ForceAttemptHTTP2:     true,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		ThrottleLimit: defaultHTTPThrottling,
		middlewares:   []ClientMiddleware{},
	}
}

func (b *ClientBuilder) WithMetricRegistry(metricsReporter MetricsReporter) *ClientBuilder {
	b.metricsReporter = metricsReporter
	return b
}

func (b *ClientBuilder) WithMiddleware(middlewares ...ClientMiddleware) *ClientBuilder {
	b.middlewares = append(b.middlewares, middlewares...)
	return b
}

func (b *ClientBuilder) WithTimeout(timeout time.Duration) *ClientBuilder {
	if timeout <= 0 {
		return b
	}
	b.clientTimeout = timeout
	return b
}

func (b *ClientBuilder) WithClient(client *http.Client) *ClientBuilder {
	b.baseClient = client
	return b
}

func (b *ClientBuilder) Build() *Client {
	client := &Client{
		clientName: b.clientName,
		client:     b.baseClient,
		OnStats:    defaultStatsHandler,
		OnError:    defaultErrorHandler,
	}

	// Register metrics middleware and stat collector
	if b.metricsReporter != nil {
		// Register metrics
		labelKeys := generateLabelKeys("client_name", "method", "status_code", "host")
		b.metricsReporter.RegisterHistogram(metricNameHistory, "Response time ms", defaultBuckets, "client_name", "method", "host")
		b.metricsReporter.RegisterSummary(metricNameSummary, "Response time ms", map[float64]float64{}, labelKeys...)
		b.metricsReporter.RegisterCounter(metricClientError, "Client errors", labelKeys...)

		// Register stat collector
		client.OnStats = func(host string, method string, headers http.Header, statusCode int, executionTime time.Duration) {
			labelVals := generateLabelVals(headers, b.clientName, method, fmt.Sprintf("%d", statusCode), host)
			if statusCode < 400 && statusCode >= 200 {
				b.metricsReporter.ObserveHistogram(metricNameHistory, toMilliseconds(executionTime), b.clientName, method, host)
			}
			b.metricsReporter.ObserveSummary(metricNameSummary, toMilliseconds(executionTime), labelVals...)
		}

		// Register error collector
		client.OnError = func(method string, headers http.Header, err error) {
			errLabel := errUnhandledHTTPErrorLabelValue
			if isTimeoutError(err) {
				errLabel = errTimeoutLabelValue
			}

			labelVals := generateLabelVals(headers, b.clientName, errLabel)
			b.metricsReporter.IncCounter(metricClientError, 1, labelVals...)
		}

		b.middlewares = append(b.middlewares, Metrics(client))
	}

	// Register throttling middleware with metrics reporting
	if b.metricsReporter != nil {
		b.metricsReporter.RegisterGauge(metricConcurrentCalls, "HTTP in-progress calls", "service")
		b.middlewares = append(b.middlewares, Throttler(client, b.ThrottleLimit, 10*time.Second, MetricOnReport(b.metricsReporter)))
	} else {
		b.middlewares = append(b.middlewares, Throttler(client, b.ThrottleLimit, 10*time.Second, NoOpReport))
	}

	// Wrap the base client with all the middleware
	for _, middleware := range b.middlewares {
		client.client = middleware(client.client)
	}

	return client
}

func isTimeoutError(err error) bool {
	switch typedErr := err.(type) {
	case net.Error:
		return typedErr.Timeout()
	case *url.Error:
		var netErr net.Error
		ok := errors.As(typedErr.Err, &netErr)
		return ok && netErr.Timeout()
	default:
		return false
	}
}
