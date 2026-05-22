package metrics

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	_incomingReqHist    = "app_request_latency_histogram"
	_incomingReqSummary = "app_request_latency"
)

type Registry interface {
	RegisterHistogram(name string, description string, buckets []float64, labels ...string)
	RegisterSummary(name string, description string, quantiles map[float64]float64, labels ...string)
	ObserveHistogram(name string, value float64, labels ...string)
	ObserveSummary(name string, value float64, labels ...string)
}

// HttpMiddleware creates http middleware that captures basic response and timing information for http endpoints.
func HttpMiddleware(registry Registry) func(next http.HandlerFunc) http.HandlerFunc {
	registry.RegisterHistogram(_incomingReqHist, "Service response time", defaultBuckets, "method", "path")
	registry.RegisterSummary(_incomingReqSummary, "Service response time with more labels", map[float64]float64{}, "method", "path", "status_code")

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			recorder := newResponseRecorder(w)

			path := strings.Split(r.Pattern, " ")
			defer func(start time.Time) {
				registry.ObserveHistogram(_incomingReqHist, ToMilliseconds(time.Since(start)), r.Method, path[len(path)-1])
				registry.ObserveSummary(_incomingReqSummary, ToMilliseconds(time.Since(start)), r.Method, path[len(path)-1], strconv.Itoa(recorder.statusCode))
			}(time.Now())

			next.ServeHTTP(recorder, r)
		}
	}
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{w, http.StatusOK}
}

func (lrw *responseRecorder) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
