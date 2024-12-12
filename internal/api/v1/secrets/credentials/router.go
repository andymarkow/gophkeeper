// Package credentials provides the credentials API.
package credentials

import (
	"log/slog"

	"github.com/go-chi/chi/v5"

	"github.com/andymarkow/gophkeeper/internal/storage/credrepo"
)

// config represents the configuration for the credentials API router.
type config struct {
	logger *slog.Logger
}

// NewRouter creates a new credentials API router.
func NewRouter(repo credrepo.Storage, cryptoKey []byte, opts ...RouterOpt) chi.Router {
	cfg := &config{
		logger: slog.New(&slog.JSONHandler{}),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	h := NewHandlers(repo, cryptoKey, WithHandlersLogger(cfg.logger))

	r := chi.NewRouter()

	return r.Group(func(r chi.Router) {
		r.Post("/", h.CreateSecret)
		r.Get("/", h.ListSecrets)
		r.Get("/{secretName}", h.GetSecret)
		r.Put("/{secretName}", h.UpdateSecret)
		r.Delete("/{secretName}", h.DeleteSecret)
	})
}

type RouterOpt func(*config)

// WithRouterLogger is a router option that sets the logger.
func WithRouterLogger(logger *slog.Logger) RouterOpt {
	return func(c *config) {
		c.logger = logger
	}
}
