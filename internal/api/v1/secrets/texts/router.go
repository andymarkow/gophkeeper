// Package texts provides the texts API.
package texts

import (
	"log/slog"

	"github.com/go-chi/chi/v5"

	"github.com/andymarkow/gophkeeper/internal/services/textsvc"
)

// config represents the configuration for the texts API router.
type config struct {
	logger *slog.Logger
}

// NewRouter creates a new texts API router.
func NewRouter(svc textsvc.Service, opts ...Option) chi.Router {
	cfg := &config{
		logger: slog.New(&slog.JSONHandler{}),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	h := NewHandlers(svc, WithHandlersLogger(cfg.logger))

	r := chi.NewRouter()

	return r.Group(func(r chi.Router) {
		r.Post("/", h.CreateSecret)
		r.Get("/", h.ListSecrets)
		r.Get("/{secretName}", h.GetSecret)
		r.Patch("/{secretName}", h.UpdateSecret)
		r.Delete("/{secretName}", h.DeleteSecret)
		r.Post("/{secretName}/upload", h.UploadSecret)
		r.Get("/{secretName}/download", h.DownloadSecret)
	})
}

// Option represents a router option.
type Option func(*config)

// WithLogger is a router option that sets the logger.
func WithLogger(logger *slog.Logger) Option {
	return func(c *config) {
		c.logger = logger
	}
}
