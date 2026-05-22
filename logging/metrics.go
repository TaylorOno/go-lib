package logging

import "log/slog"

const (
	_logMetric         = "logger_metric"
	_statusMetric      = "logger"
	_loggedBytesMetric = "logged_bytes_total"
)

var (
	metricsReporter MetricsReporter
	bytesLoggedFunc = func(bytes int) {}
)

// MetricsReporter reports log statistics.
type MetricsReporter interface {
	RegisterCounter(name string, help string, labels ...string)
	IncCounter(name string, value float64, labels ...string)
}

// WithMetricReporter registers a metrics reporter.
func WithMetricReporter(m MetricsReporter) {
	metricsReporter = m
	metricsReporter.RegisterCounter(_statusMetric, "log statistics", "loglevel", "name", "message")
	metricsReporter.RegisterCounter(_loggedBytesMetric, "amount logged in bytes")
	bytesLoggedFunc = func(bytes int) {
		metricsReporter.IncCounter(_loggedBytesMetric, float64(bytes))
	}
}

// Metric adds a metric to the log entry.
func Metric(status Status) slog.Attr {
	return slog.Any(_logMetric, status)
}

// emitMetric will emit a metric if a metric reporter is configured and a metric is present in the log entry.
func emitMetric(lvl slog.Level) func(a slog.Attr) bool {
	return func(a slog.Attr) bool {
		if metricsReporter != nil && a.Key == _logMetric {
			if status, ok := a.Value.Any().(Status); ok {
				go metricsReporter.IncCounter(_statusMetric, 1, lvl.String(), status.Name(), status.Message())
			}
			return false
		}
		return true
	}
}
