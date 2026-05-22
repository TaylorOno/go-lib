package metrics

import (
	"net/http"
	"strings"
	"time"
)

const (
	metricTypeCounter   = "counter"
	metricTypeGauge     = "gauge"
	metricTypeHistogram = "histogram"
	metricTypeSummary   = "summary"
)

var (
	sanitizer      = strings.NewReplacer(".", "_", "-", "_")
	defaultBuckets = []float64{0, 5, 10, 25, 50, 75, 100, 250, 500, 750, 1000, 2500, 5000, 7500, 10000}
)

type Reporter interface {
	Routes(mux *http.ServeMux)
	GetMetricsDefinition() map[string]MetricDefinition
	Registry
}

type MetricDefinition struct {
	Kind        string
	Description string
	labelCount  int
	Labels      []string
}

func ToMilliseconds(duration time.Duration) float64 {
	return float64(duration) / float64(time.Millisecond)
}

func sanitize(labels []string) {
	for i, label := range labels {
		sanitizedLabels := sanitizer.Replace(label)
		labels[i] = sanitizedLabels
	}
}
