package web

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/taylorono/go-lib/profiler"
)

var (
	port      string
	debugPort string
)

func init() {
	flag.StringVar(&port, "port", "8080", "port to listen on")
	flag.StringVar(&debugPort, "debug-port", "", "when set pprof will be enabled on this port")
}

type Middleware func(next http.HandlerFunc) http.HandlerFunc

// Server represents a web server suitable for kubernetes deployments.
type Server struct {
	port              string
	debugPort         string
	mux               *http.ServeMux
	readinessRegistry *healthRegistry
	livenessRegistry  *healthRegistry
	middleware        []Middleware
}

// NewServer Creates a new web server with the given options.
func NewServer(opts ...OptionFunc) *Server {
	// default server
	s := &Server{
		port:              "8080",
		mux:               http.NewServeMux(),
		readinessRegistry: newHealthRegistry(),
		livenessRegistry:  newHealthRegistry(),
		middleware:        []Middleware{},
	}

	// apply config overrides
	for _, opt := range opts {
		opt(s)
	}

	return s
}

// HandleFunc registers a new route with the given pattern and handler function applying any global middleware.
func (s *Server) HandleFunc(pattern string, handler http.HandlerFunc) {
	// apply any configured middleware
	for _, m := range s.middleware {
		handler = m(handler)
	}

	s.mux.HandleFunc(pattern, handler)
}

// Start starts the web server with the given context and will block until the context has been canceled. A context cancellation will cause a graceful shutdown.
func (s *Server) Start(ctx context.Context) error {
	var err error

	// Configure Server
	httpServer := &http.Server{
		Handler: s.mux,
	}

	listener, err := net.Listen("tcp", net.JoinHostPort("", s.port))
	if err != nil {
		return err
	}

	// Register health endpoints
	s.mux.HandleFunc("/healthz", s.livenessHandler())
	s.mux.HandleFunc("/readyz", s.readinessHandler())

	// Server loop
	go func() {
		slog.Info(fmt.Sprintf("listening on %s\n", listener.Addr()))
		if err := httpServer.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
	}()

	// Launch pprof if the port has been specified
	if s.debugPort != "" {
		profiler.ListenAndServe(ctx, s.debugPort)
	}

	// Allow for a graceful shutdown ctx.Done() will block until the application receives a SIGTERM or SIGINT
	<-ctx.Done()

	// Graceful shutdown operations
	slog.Info("stopping webserver")

	// Wait for 10 seconds before forcing a shutdown.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = httpServer.Shutdown(shutdownCtx); err != nil {
		return err
	}

	return nil
}
