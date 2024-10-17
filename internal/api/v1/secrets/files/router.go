// Package files provides the files API.
package files

import (
	"log/slog"

	"github.com/go-chi/chi/v5"

	"github.com/andymarkow/gophkeeper/internal/services/filesvc"
)

// config represents the configuration for the files API router.
type config struct {
	logger *slog.Logger
}

// NewRouter creates a new files API router.
func NewRouter(svc *filesvc.Service, opts ...Option) chi.Router {
	cfg := &config{
		logger: slog.New(&slog.JSONHandler{}),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	h := NewHandlers(svc, WithHandlersLogger(cfg.logger))

	r := chi.NewRouter()

	return r.Group(func(r chi.Router) {
		r.Post("/", h.CreateFile)
		// r.Get("/", h.ListFiles)
		// r.Get("/{fileID}", h.GetFile)
		// r.Put("/{fileID}", h.UpdateFile)
		// r.Delete("/{fileID}", h.DeleteFile)
	})
}

type Option func(*config)

// WithLogger is a router option that sets the logger.
func WithLogger(logger *slog.Logger) Option {
	return func(c *config) {
		c.logger = logger
	}
}
