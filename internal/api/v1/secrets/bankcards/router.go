// Package bankcards provides the bankcards API.
package bankcards

import (
	"log/slog"

	"github.com/go-chi/chi/v5"

	"github.com/andymarkow/gophkeeper/internal/storage/cardrepo"
)

// Options represents the options for the bankcards API router.
type Options struct {
	Logger *slog.Logger
}

// NewRouter creates a new bankcards API router.
func NewRouter(repo cardrepo.Storage, cryptoKey []byte, opts *Options) chi.Router {
	if opts == nil {
		opts = defaultOptions()
	}

	h := NewHandlers(repo, cryptoKey, WithLogger(opts.Logger))

	r := chi.NewRouter()

	return r.Group(func(r chi.Router) {
		r.Post("/", h.CreateSecret)
		r.Get("/", h.ListSecrets)
		r.Get("/{secretName}", h.GetSecret)
		r.Put("/{secretName}", h.UpdateSecret)
		r.Delete("/{secretName}", h.DeleteSecret)
	})
}

// defaultOptions returns the default options.
func defaultOptions() *Options {
	return &Options{
		Logger: slog.New(&slog.JSONHandler{}),
	}
}
