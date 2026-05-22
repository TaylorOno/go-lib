package rest

import (
	"errors"
	"net/http"
	"sync"
	"time"
)

// ErrMaxCallsReached Thrown when the downstream system is still processing a given number of requests
var ErrMaxCallsReached = errors.New("max concurrency reached")

type OnReport func(callAlias string, currentConcurrency int)

func NoOpReport(_ string, _ int) {}

func MetricOnReport(metricsReporter MetricsReporter) func(callAlias string, currentConcurrency int) {
	return func(callAlias string, currentConcurrency int) {
		metricsReporter.SetGauge(metricConcurrentCalls, float64(currentConcurrency), callAlias)
	}
}

type throttle struct {
	max     int
	current int
	lock    sync.Mutex
}

func (t *throttle) AtMaxConcurrency() bool {
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.current < t.max {
		t.current = t.current + 1
		return false
	}

	return true
}

// Throttler middleware to control the maximum number of inflight requests this client is allowed to have
func Throttler(client *Client, maxCalls int, reportingInterval time.Duration, report OnReport) ClientMiddleware {
	control := &throttle{
		max: maxCalls,
	}

	go func() {
		for {
			time.Sleep(reportingInterval)
			control.lock.Lock()
			current := control.current
			control.lock.Unlock()
			report(client.clientName, current)
		}
	}()

	return func(c Doer) Doer {
		return ClientFunc(func(r *http.Request) (*http.Response, error) {
			if control.AtMaxConcurrency() {
				client.OnError(r.Method, r.Header, ErrMaxCallsReached)
				return nil, ErrMaxCallsReached
			}

			defer func() {
				control.lock.Lock()
				control.current = control.current - 1
				control.lock.Unlock()
			}()

			return c.Do(r)
		})
	}
}
