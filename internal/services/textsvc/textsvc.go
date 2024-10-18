// Package textsvc provides the file service.
package textsvc

import "log/slog"

// Service represents the service.
type Service struct {
	log *slog.Logger
}

// NewService creates a new service.
func NewService(opts ...Option) (*Service, error) {
	svc := &Service{
		log: slog.New(&slog.JSONHandler{}),
	}

	for _, opt := range opts {
		opt(svc)
	}

	return svc, nil
}

// Option is a functional option for the service.
type Option func(*Service)

// WithLogger sets the logger for the service.
func WithLogger(log *slog.Logger) Option {
	return func(s *Service) {
		s.log = log
	}
}
