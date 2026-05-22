package metrics

import (
	"context"
	"net/http"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type OTELReporter struct {
	sync.RWMutex
	meter                metric.Meter
	metricRegistry       map[string]MetricDefinition
	counterDefinitions   map[string]metric.Float64Counter
	gaugeDefinitions     map[string]metric.Float64Gauge
	summaryDefinitions   map[string]metric.Float64Histogram
	histogramDefinitions map[string]metric.Float64Histogram
}

func NewOTELReporter() *OTELReporter {
	return &OTELReporter{
		meter:                otel.GetMeterProvider().Meter("otel-reporter"),
		metricRegistry:       make(map[string]MetricDefinition),
		counterDefinitions:   make(map[string]metric.Float64Counter),
		gaugeDefinitions:     make(map[string]metric.Float64Gauge),
		summaryDefinitions:   make(map[string]metric.Float64Histogram),
		histogramDefinitions: make(map[string]metric.Float64Histogram),
	}
}

func (r *OTELReporter) registerMetrics(name string, description string, kind string, labels []string) {
	r.Lock()
	r.metricRegistry[name] = MetricDefinition{
		Kind:        kind,
		Description: description,
		Labels:      labels,
		labelCount:  len(labels),
	}
	r.Unlock()
}

func (r *OTELReporter) RegisterCounter(name string, description string, labels ...string) {
	sanitize(labels)
	r.registerMetrics(name, description, metricTypeCounter, labels)

	r.Lock()
	counter, err := r.meter.Float64Counter(name, metric.WithDescription(description))
	if err != nil {
		panic(err)
	}
	r.counterDefinitions[name] = counter
	r.Unlock()
}

func (r *OTELReporter) RegisterGauge(name string, description string, labels ...string) {
	sanitize(labels)
	r.registerMetrics(name, description, metricTypeGauge, labels)

	r.Lock()
	gauge, err := r.meter.Float64Gauge(name, metric.WithDescription(description))
	if err != nil {
		panic(err)
	}
	r.gaugeDefinitions[name] = gauge
	r.Unlock()
}

func (r *OTELReporter) RegisterSummary(name string, description string, _ map[float64]float64, labels ...string) {
	sanitize(labels)
	r.registerMetrics(name, description, metricTypeSummary, labels)

	r.Lock()
	histogram, err := r.meter.Float64Histogram(
		name,
		metric.WithDescription(description),
		metric.WithExplicitBucketBoundaries(defaultBuckets...),
	)
	if err != nil {
		panic(err)
	}
	r.summaryDefinitions[name] = histogram
	r.Unlock()
}

func (r *OTELReporter) RegisterHistogram(name string, description string, buckets []float64, labels ...string) {
	sanitize(labels)
	r.registerMetrics(name, description, metricTypeHistogram, labels)

	if len(buckets) == 0 {
		buckets = defaultBuckets
	}

	r.Lock()
	histogram, err := r.meter.Float64Histogram(
		name,
		metric.WithDescription(description),
		metric.WithExplicitBucketBoundaries(buckets...),
	)
	if err != nil {
		panic(err)
	}
	r.histogramDefinitions[name] = histogram
	r.Unlock()
}

func (r *OTELReporter) IncCounter(name string, value float64, labels ...string) {
	r.RLock()
	if meter, ok := r.counterDefinitions[name]; ok {
		labelCount := r.metricRegistry[name].labelCount
		if labelCount == len(labels) {
			attributes := toAttributeSet(r.metricRegistry[name].Labels, labels)
			meter.Add(context.Background(), value, metric.WithAttributeSet(attributes))
		} else {
			// Error
		}
	}
	r.RUnlock()
}

func (r *OTELReporter) SetGauge(name string, value float64, labels ...string) {
	r.RLock()
	if meter, ok := r.gaugeDefinitions[name]; ok {
		labelCount := r.metricRegistry[name].labelCount
		if labelCount == len(labels) {
			attributes := toAttributeSet(r.metricRegistry[name].Labels, labels)
			meter.Record(context.Background(), value, metric.WithAttributeSet(attributes))
		} else {
			// Error
		}
	}
	r.RUnlock()
}

func (r *OTELReporter) ObserveSummary(name string, value float64, labels ...string) {
	r.RLock()
	if meter, ok := r.summaryDefinitions[name]; ok {
		labelCount := r.metricRegistry[name].labelCount
		if labelCount == len(labels) {
			attributes := toAttributeSet(r.metricRegistry[name].Labels, labels)
			meter.Record(context.Background(), value, metric.WithAttributeSet(attributes))
		} else {
			// Error
		}
	}
	r.RUnlock()
}

func (r *OTELReporter) ObserveHistogram(name string, value float64, labels ...string) {
	r.RLock()
	if meter, ok := r.histogramDefinitions[name]; ok {
		labelCount := r.metricRegistry[name].labelCount
		if labelCount == len(labels) {
			attributes := toAttributeSet(r.metricRegistry[name].Labels, labels)
			meter.Record(context.Background(), value, metric.WithAttributeSet(attributes))
		} else {
			// Error
		}
	}
	r.RUnlock()
}

func (r *OTELReporter) Routes(mux *http.ServeMux) {
	mux.HandleFunc("/metrics/docs", MetricDocs(r))
}

// GetMetricsDefinition returns the definition of all the metrics that have been registered using this package
func (r *OTELReporter) GetMetricsDefinition() map[string]MetricDefinition {
	// Creating a copy to avoid exposing the internal map to external manipulation
	metrics := make(map[string]MetricDefinition)
	r.RLock()
	for k, v := range r.metricRegistry {
		metrics[k] = v
	}
	r.RUnlock()
	return metrics
}

func toAttributeSet(labelName []string, labelValue []string) attribute.Set {
	attributes := make([]attribute.KeyValue, len(labelName))
	for i, name := range labelName {
		attributes[i] = attribute.String(name, labelValue[i])
	}

	return attribute.NewSet(attributes...)

}
