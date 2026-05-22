package metrics

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusReporter struct {
	sync.RWMutex
	metricRegistry       map[string]MetricDefinition
	counterDefinitions   map[string]*prometheus.CounterVec
	gaugeDefinitions     map[string]*prometheus.GaugeVec
	summaryDefinitions   map[string]*prometheus.SummaryVec
	histogramDefinitions map[string]*prometheus.HistogramVec
}

func NewPrometheusReporter() *PrometheusReporter {
	return &PrometheusReporter{
		metricRegistry:       make(map[string]MetricDefinition),
		counterDefinitions:   make(map[string]*prometheus.CounterVec),
		gaugeDefinitions:     make(map[string]*prometheus.GaugeVec),
		summaryDefinitions:   make(map[string]*prometheus.SummaryVec),
		histogramDefinitions: make(map[string]*prometheus.HistogramVec),
	}
}

func (p *PrometheusReporter) registerMetrics(name string, description string, kind string, labels []string) {
	p.Lock()
	p.metricRegistry[name] = MetricDefinition{
		Kind:        kind,
		Description: description,
		Labels:      labels,
		labelCount:  len(labels),
	}
	p.Unlock()
}

func (p *PrometheusReporter) RegisterCounter(name string, description string, labels ...string) {
	sanitize(labels)

	opts := prometheus.CounterOpts{Name: name, Help: description}
	counter := prometheus.NewCounterVec(opts, labels)
	p.registerMetrics(name, description, metricTypeCounter, labels)

	p.Lock()
	prometheus.MustRegister(counter)
	p.counterDefinitions[name] = counter
	p.Unlock()
}

func (p *PrometheusReporter) RegisterGauge(name string, description string, labels ...string) {
	sanitize(labels)

	opts := prometheus.GaugeOpts{Name: name, Help: description}
	gauge := prometheus.NewGaugeVec(opts, labels)
	p.registerMetrics(name, description, metricTypeGauge, labels)

	p.Lock()
	prometheus.MustRegister(gauge)
	p.gaugeDefinitions[name] = gauge
	p.Unlock()
}

func (p *PrometheusReporter) RegisterSummary(name string, description string, quantiles map[float64]float64, labels ...string) {
	sanitize(labels)

	opts := prometheus.SummaryOpts{Name: name, Help: description, Objectives: quantiles}
	summary := prometheus.NewSummaryVec(opts, labels)
	p.registerMetrics(name, description, metricTypeSummary, labels)

	p.Lock()
	prometheus.MustRegister(summary)
	p.summaryDefinitions[name] = summary
	p.Unlock()
}

func (p *PrometheusReporter) RegisterHistogram(name string, description string, buckets []float64, labels ...string) {
	sanitize(labels)

	if len(buckets) == 0 {
		buckets = prometheus.DefBuckets
	}

	opts := prometheus.HistogramOpts{Name: name, Help: description, Buckets: buckets}
	histogram := prometheus.NewHistogramVec(opts, labels)
	p.registerMetrics(name, description, metricTypeHistogram, labels)

	p.Lock()
	prometheus.MustRegister(histogram)
	p.histogramDefinitions[name] = histogram
	p.Unlock()
}

func (p *PrometheusReporter) IncCounter(name string, value float64, labels ...string) {
	p.RLock()
	if metric, ok := p.counterDefinitions[name]; ok {
		labelCount := p.metricRegistry[name].labelCount
		if labelCount == len(labels) {
			metric.WithLabelValues(labels...).Add(value)
		} else {
			// Error
		}
	}
	p.RUnlock()
}

func (p *PrometheusReporter) SetGauge(name string, value float64, labels ...string) {
	p.RLock()
	if metric, ok := p.gaugeDefinitions[name]; ok {
		labelCount := p.metricRegistry[name].labelCount
		if labelCount == len(labels) {
			metric.WithLabelValues(labels...).Set(value)
		}
	} else {
		// Error
	}
	p.RUnlock()
}

func (p *PrometheusReporter) ObserveSummary(name string, value float64, labels ...string) {
	p.RLock()
	if metric, ok := p.summaryDefinitions[name]; ok {
		labelCount := p.metricRegistry[name].labelCount
		if labelCount == len(labels) {
			metric.WithLabelValues(labels...).Observe(value)
		}
	} else {
		// Error
	}
	p.RUnlock()
}

func (p *PrometheusReporter) ObserveHistogram(name string, value float64, labels ...string) {
	p.RLock()
	if metric, ok := p.histogramDefinitions[name]; ok {
		labelCount := p.metricRegistry[name].labelCount
		if labelCount == len(labels) {
			metric.WithLabelValues(labels...).Observe(value)
		}
	} else {
		// Error
	}
	p.RUnlock()
}

func (p *PrometheusReporter) Routes(mux *http.ServeMux) {
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/metrics/docs", MetricDocs(p))
}

// GetMetricsDefinition returns the definition of all the metrics that have been registered using this package
func (p *PrometheusReporter) GetMetricsDefinition() map[string]MetricDefinition {
	// Creating a copy to avoid exposing the internal map to external manipulation
	metrics := make(map[string]MetricDefinition)
	p.RLock()
	for k, v := range p.metricRegistry {
		metrics[k] = v
	}
	p.RUnlock()
	return metrics
}
