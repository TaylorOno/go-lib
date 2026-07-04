package web

import (
	"github.com/taylorono/go-lib/metrics"
	"github.com/taylorono/go-lib/traces"
)

type OptionFunc func(*Server)

func WithPort(port string) OptionFunc {
	return func(o *Server) {
		o.port = port
	}
}

func WithDebugPort(port string) OptionFunc {
	return func(o *Server) {
		o.debugPort = port
	}
}

func WithMiddleware(middleware ...Middleware) OptionFunc {
	return func(o *Server) {
		o.middleware = append(o.middleware, middleware...)
	}
}

func WithMetricRegistry(registry metrics.Reporter) OptionFunc {
	return func(o *Server) {
		// Register metrics routes before middleware to avoid instrumentation.
		registry.Routes(o.mux)

		// Add default instrumentation middleware
		o.middleware = append(o.middleware, metrics.HttpMiddleware(registry))
	}
}

func WithTracer(provider traces.Provider) OptionFunc {
	return func(o *Server) {
		o.middleware = append(o.middleware, traces.HttpMiddleware(provider))
	}
}

// WithReadinessCheck adds a new check to the readiness probe.
func WithReadinessCheck(name string, check Check) OptionFunc {
	return func(s *Server) {
		s.RegisterReadinessCheck(name, check)
	}
}

// WithLivenessCheck adds a new check to the liveness probe. Use this sparingly, as liveness probes should generally just confirm the process is running.
func WithLivenessCheck(name string, check Check) OptionFunc {
	return func(s *Server) {
		s.RegisterLivenessCheck(name, check)
	}
}
