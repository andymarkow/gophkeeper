// Package api provides the API.
package api

import (
	"log/slog"

	"github.com/go-chi/chi/v5"

	"github.com/andymarkow/gophkeeper/internal/api/v1/users"
	"github.com/andymarkow/gophkeeper/internal/storage/userrepo"
)

// options represents the options for the API.
type options struct {
	logger *slog.Logger
}

// NewAPI creates a new API router.
func NewAPI(userStorage userrepo.Storage, opts ...Option) chi.Router {
	defOpts := &options{
		logger: slog.New(&slog.JSONHandler{}),
	}

	for _, opt := range opts {
		opt(defOpts)
	}

	usersAPI := users.NewUsers(userStorage, "secret", &users.Options{
		Logger: defOpts.logger,
	})

	r := chi.NewRouter()

	r.Route("/api/v1", func(r chi.Router) {
		r.Mount("/users", usersAPI)
	})

	return r
}

// Option represents an option for the API.
type Option func(*options)

// WithLogger sets the logger for the API.
func WithLogger(logger *slog.Logger) Option {
	return func(o *options) {
		o.logger = logger
	}
}
