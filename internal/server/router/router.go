// Package router provides HTTP server router.
package router

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"

	"github.com/andymarkow/gophkeeper/internal/api/v1/secrets/bankcards"
	"github.com/andymarkow/gophkeeper/internal/api/v1/users"
	"github.com/andymarkow/gophkeeper/internal/middlewares"
	"github.com/andymarkow/gophkeeper/internal/storage/cardrepo"
	"github.com/andymarkow/gophkeeper/internal/storage/userrepo"
)

// options represents the options for the API.
type options struct {
	logger *slog.Logger
}

// NewRouter creates a new API router.
func NewRouter(userStorage userrepo.Storage, cardStorage cardrepo.Storage,
	jwtSecret []byte, cryproKey []byte, opts ...Option,
) chi.Router {
	defOpts := &options{
		logger: slog.New(&slog.JSONHandler{}),
	}

	for _, opt := range opts {
		opt(defOpts)
	}

	jwtAuth := jwtauth.New("HS256", jwtSecret, nil)

	usersAPI := users.NewRouter(userStorage, jwtSecret, &users.Options{
		Logger: defOpts.logger,
	})

	cardsAPI := bankcards.NewRouter(cardStorage, cryproKey, &bankcards.Options{
		Logger: defOpts.logger,
	})

	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)
	r.Use(middlewares.Logger(defOpts.logger))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Mount("/users", usersAPI)

		r.Route("/secrets", func(r chi.Router) {
			r.Use(jwtauth.Verifier(jwtAuth))
			r.Use(jwtauth.Authenticator(jwtAuth))
			r.Use(middlewares.UserID)

			r.Mount("/bankcards", cardsAPI)
			// r.Mount("/credentials", )
			// r.Mount("/generics", )
		})
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
