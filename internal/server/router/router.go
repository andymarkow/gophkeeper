// Package router provides HTTP server router.
package router

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"

	"github.com/andymarkow/gophkeeper/internal/api/v1/secrets/bankcards"
	"github.com/andymarkow/gophkeeper/internal/api/v1/secrets/credentials"
	"github.com/andymarkow/gophkeeper/internal/api/v1/secrets/files"
	"github.com/andymarkow/gophkeeper/internal/api/v1/users"
	"github.com/andymarkow/gophkeeper/internal/middlewares"
	"github.com/andymarkow/gophkeeper/internal/services/filesvc"
	"github.com/andymarkow/gophkeeper/internal/storage/cardrepo"
	"github.com/andymarkow/gophkeeper/internal/storage/credrepo"
	"github.com/andymarkow/gophkeeper/internal/storage/userrepo"
)

// options represents the options for the API.
type options struct {
	logger    *slog.Logger
	jwtSecret []byte
	cryptoKey []byte
}

// NewRouter creates a new API router.
func NewRouter(userStorage userrepo.Storage, cardStorage cardrepo.Storage,
	credStorage credrepo.Storage, fileSvc filesvc.Service, opts ...Option,
) chi.Router {
	defOpts := &options{
		logger: slog.New(&slog.JSONHandler{}),
	}

	for _, opt := range opts {
		opt(defOpts)
	}

	jwtAuth := jwtauth.New("HS256", defOpts.jwtSecret, nil)

	usersAPI := users.NewRouter(userStorage, defOpts.jwtSecret, &users.Options{
		Logger: defOpts.logger,
	})

	cardsAPI := bankcards.NewRouter(cardStorage, defOpts.cryptoKey, &bankcards.Options{
		Logger: defOpts.logger,
	})

	credsAPI := credentials.NewRouter(credStorage, defOpts.cryptoKey, credentials.WithRouterLogger(defOpts.logger))

	filesAPI := files.NewRouter(fileSvc, files.WithLogger(defOpts.logger))

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
			r.Mount("/credentials", credsAPI)
			r.Mount("/files", filesAPI)
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

// WithJWTSecret sets the JWT secret for the API.
func WithJWTSecret(jwtSecret []byte) Option {
	return func(o *options) {
		o.jwtSecret = jwtSecret
	}
}

// WithCryptoKey sets the crypto key for the API.
func WithCryptoKey(cryptoKey []byte) Option {
	return func(o *options) {
		o.cryptoKey = cryptoKey
	}
}
