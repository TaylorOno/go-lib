package web

import (
	"context"
	"crypto/tls"
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

	defaultShutdownTimeout = 10 * time.Second
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
	cert              string
	key               string
	shutdownTimeout   time.Duration
	httpServer        *http.Server
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
		shutdownTimeout:   defaultShutdownTimeout,
		mux:               http.NewServeMux(),
		readinessRegistry: newHealthRegistry(),
		livenessRegistry:  newHealthRegistry(),
		middleware:        []Middleware{},
		httpServer:        &http.Server{},
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

// SetIdleTimeout Overrides the default SetIdleTimeout
func (s *Server) SetIdleTimeout(timeout time.Duration) {
	s.httpServer.IdleTimeout = timeout
}

// SetReadTimeout Overrides the default ReadTimeout
func (s *Server) SetReadTimeout(timeout time.Duration) {
	s.httpServer.ReadTimeout = timeout
}

// SetWriteTimeout Overrides the default WriteTimeout
func (s *Server) SetWriteTimeout(timeout time.Duration) {
	s.httpServer.WriteTimeout = timeout
}

// SetCustomTLSConfig Allows setting a custom TLS configuration, avoid using this configuration unless truly needed
// as it might compromise security if not handled correctly
func (s *Server) SetCustomTLSConfig(tlsConfig *tls.Config) {
	s.httpServer.TLSConfig = tlsConfig
}

// SetCustomServer Overrides the entire http.Server settings, avoid using this configuration unless truly needed
// as it might compromise security if not handled correctly
func (s *Server) SetCustomServer(customServer *http.Server) {
	s.httpServer = customServer
}

// Start starts the web server with the given context and will block until the context has been canceled. A context cancellation will cause a graceful shutdown.
func (s *Server) Start(ctx context.Context) error {
	var err error

	// Configure Server
	s.httpServer.Handler = s.mux

	listener, err := net.Listen("tcp", net.JoinHostPort("", s.port))
	if err != nil {
		return err
	}

	// Register health endpoints
	s.mux.HandleFunc("/healthz", s.livenessHandler())
	s.mux.HandleFunc("/readyz", s.readinessHandler())

	// Server loop
	go func() {
		s.server(listener)
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
	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()
	if err = s.httpServer.Shutdown(shutdownCtx); err != nil {
		return err
	}

	return nil
}

func (s *Server) server(listener net.Listener) {
	slog.Info(fmt.Sprintf("listening on %s\n", listener.Addr()))
	if s.cert != "" && s.key != "" {
		if err := s.httpServer.ServeTLS(listener, s.cert, s.key); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
	} else {
		if err := s.httpServer.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
	}
}
