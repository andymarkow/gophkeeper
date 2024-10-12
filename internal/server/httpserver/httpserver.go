// Package httpserver provides the HTTP server.
package httpserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// HTTPServer represents an HTTP server.
type HTTPServer struct {
	log    *slog.Logger
	server *http.Server
}

// NewHTTPServer creates a new HTTP server.
func NewHTTPServer(h http.Handler, opts ...Option) *HTTPServer {
	srv := &HTTPServer{
		server: &http.Server{
			Addr:              ":8080",
			Handler:           h,
			ReadTimeout:       30 * time.Second,
			ReadHeaderTimeout: 10 * time.Second,
			WriteTimeout:      30 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(srv)
	}

	return srv
}

// Option represents an HTTP server option.
type Option func(s *HTTPServer)

// WithLogger sets the logger for the HTTP server.
func WithLogger(log *slog.Logger) Option {
	return func(s *HTTPServer) {
		s.log = log
	}
}

// WithServerAddr sets the HTTP server address.
func WithServerAddr(addr string) Option {
	return func(s *HTTPServer) {
		s.server.Addr = addr
	}
}

// WithReadTimeout sets the HTTP server read timeout.
func WithReadTimeout(timeout time.Duration) Option {
	return func(s *HTTPServer) {
		s.server.ReadTimeout = timeout
	}
}

// WithReadHeaderTimeout sets the HTTP server read header timeout.
func WithReadHeaderTimeout(timeout time.Duration) Option {
	return func(s *HTTPServer) {
		s.server.ReadHeaderTimeout = timeout
	}
}

// WithWriteTimeout sets the HTTP server write timeout.
func WithWriteTimeout(timeout time.Duration) Option {
	return func(s *HTTPServer) {
		s.server.WriteTimeout = timeout
	}
}

// Serve starts the HTTP server.
func (s *HTTPServer) Serve() error {
	s.log.Info("Starting HTTP server", slog.String("address", s.server.Addr))

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server.ListenAndServe: %w", err)
	}

	return nil
}

// Shutdown stops the HTTP server.
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	s.log.Info("Shutting down HTTP server")

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server.Shutdown: %w", err)
	}

	return nil
}
