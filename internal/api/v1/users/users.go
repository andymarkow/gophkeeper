// Package users provides the users API.
package users

import (
	"log/slog"

	"github.com/go-chi/chi/v5"

	"github.com/andymarkow/gophkeeper/internal/auth"
	"github.com/andymarkow/gophkeeper/internal/storage/userrepo"
)

// Options represents the options for the NewUsers function.
type Options struct {
	Logger *slog.Logger
}

// NewUsers creates a new users API router.
func NewUsers(repo userrepo.Storage, jwtSecret string, opts *Options) chi.Router {
	if opts == nil {
		opts = defaultOptions()
	}

	jwtAuth := auth.NewJWTAuth(jwtSecret)

	h := NewHandlers(repo, jwtAuth, WithLogger(opts.Logger))

	r := chi.NewRouter()

	r.Post("/signup", h.CreateUser)
	r.Post("/signin", h.LoginUser)

	return r
}

// defaultOptions returns the default options.
func defaultOptions() *Options {
	return &Options{
		Logger: slog.New(&slog.JSONHandler{}),
	}
}
