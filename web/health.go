package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

const (
	StatusUp   = "UP"
	StatusDown = "DOWN"
)

// Check defines the function signature for a health check.
type Check func(ctx context.Context) error

// healthResponse defines the JSON structure for health check responses.
type healthResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks,omitempty"`
}

// healthRegistry manages a collection of health checks.
type healthRegistry struct {
	mu     sync.RWMutex
	checks map[string]Check
}

func newHealthRegistry() *healthRegistry {
	return &healthRegistry{
		checks: make(map[string]Check),
	}
}

func (r *healthRegistry) register(name string, check Check) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checks[name] = check
}

func (r *healthRegistry) run(ctx context.Context) map[string]error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make(map[string]error, len(r.checks))
	for name, check := range r.checks {
		results[name] = check(ctx)
	}
	return results
}

// RegisterReadinessCheck adds a new check to the readiness probe.
func (s *Server) RegisterReadinessCheck(name string, check Check) {
	s.readinessRegistry.register(name, check)
}

// ReadinessHandler returns a handler that checks all registered readiness checks.
func (s *Server) readinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		results := s.readinessRegistry.run(r.Context())

		status := http.StatusOK
		response := healthResponse{
			Status: StatusUp,
			Checks: make(map[string]string),
		}

		for name, err := range results {
			if err != nil {
				status = http.StatusServiceUnavailable
				response.Status = StatusDown
				response.Checks[name] = err.Error()
			} else {
				response.Checks[name] = StatusUp
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(response)
	}
}

// RegisterLivenessCheck adds a new check to the liveness probe.
// Use this sparingly, as liveness probes should generally just confirm the process is running.
func (s *Server) RegisterLivenessCheck(name string, check Check) {
	s.livenessRegistry.register(name, check)
}

// LivenessHandler returns a handler that checks all registered liveness checks.
func (s *Server) livenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		results := s.livenessRegistry.run(r.Context())

		for _, err := range results {
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	}
}
