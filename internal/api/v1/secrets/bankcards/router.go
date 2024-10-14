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
		r.Post("/", h.CreateCard)
		r.Get("/", h.ListCards)
		r.Get("/{cardID}", h.GetCard)
		r.Put("/{cardID}", h.UpdateCard)
		r.Delete("/{cardID}", h.DeleteCard)
	})
}

// defaultOptions returns the default options.
func defaultOptions() *Options {
	return &Options{
		Logger: slog.New(&slog.JSONHandler{}),
	}
}
