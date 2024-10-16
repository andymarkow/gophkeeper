// Package filesvc provides the file service.
package filesvc

import "log/slog"

// Service represents the service.
type Service struct {
	log *slog.Logger
}

// NewService creates a new service.
func NewService(opts ...Option) (*Service, error) {
	return nil, nil
}

// Option is a functional option for the service.
type Option func(*Service)

// WithLogger sets the logger for the service.
func WithLogger(log *slog.Logger) Option {
	return func(s *Service) {
		s.log = log
	}
}
