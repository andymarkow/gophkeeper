package users

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/andymarkow/gophkeeper/internal/api/v1/response"
	"github.com/andymarkow/gophkeeper/internal/auth"
	"github.com/andymarkow/gophkeeper/internal/domain/user"
	"github.com/andymarkow/gophkeeper/internal/httperr"
	"github.com/andymarkow/gophkeeper/internal/storage/userrepo"
)

// Handlers represents user API handlers.
type Handlers struct {
	log     *slog.Logger
	auth    *auth.JWTAuth
	storage userrepo.Storage
}

// NewHandlers returns a new Handlers instance.
func NewHandlers(repo userrepo.Storage, auth *auth.JWTAuth, opts ...Option) *Handlers {
	userAPI := &Handlers{
		log:     slog.New(&slog.JSONHandler{}),
		auth:    auth,
		storage: repo,
	}

	for _, opt := range opts {
		opt(userAPI)
	}

	return userAPI
}

// Option is a functional option type for Handlers.
type Option func(*Handlers)

// WithLogger is an option for Handlers instance that sets logger.
func WithLogger(log *slog.Logger) Option {
	return func(u *Handlers) {
		u.log = log
	}
}

// CreateUser creates a new user.
func (h *Handlers) CreateUser(w http.ResponseWriter, req *http.Request) {
	var payload CreateUserRequest

	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		if errors.Is(err, io.EOF) {
			h.log.Error("json.NewDecoder.Decode", slog.Any("error", err))
			httperr.HandleError(w, ErrReqPayloadEmpty)

			return
		}

		h.log.Error("json.NewDecoder.Decode", slog.Any("error", err))
		httperr.HandleError(w, httperr.NewHTTPError(http.StatusBadRequest, err))

		return
	}

	defer req.Body.Close()

	usr, err := user.CreateUser(payload.Login, payload.Password)
	if err != nil {
		h.log.Error("user.CreateUser", slog.Any("error", err))
		httperr.HandleError(w, httperr.NewHTTPError(http.StatusBadRequest, err))

		return
	}

	if err := h.storage.CreateUser(req.Context(), usr); err != nil {
		if errors.Is(err, userrepo.ErrUsrAlreadyExists) {
			h.log.Error("storage.CreateUser", slog.Any("error", err))
			httperr.HandleError(w, ErrUsrAlreadyExists)

			return
		}

		h.log.Error("storage.CreateUser", slog.Any("error", err))
		httperr.HandleError(w, httperr.NewHTTPError(http.StatusInternalServerError, err))

		return
	}

	token, err := h.auth.CreateJWTString(usr.Login())
	if err != nil {
		h.log.Error("jwtauth.CreateJWTString", slog.Any("error", err))
		httperr.HandleError(w, httperr.NewHTTPError(http.StatusInternalServerError, err))

		return
	}

	// w.Header().Set("Authorization", "Bearer "+token)
	response.JSONResponse(w, http.StatusOK, CreateUserResponse{Token: token})
}

// LoginUser logs in a user.
func (h *Handlers) LoginUser(w http.ResponseWriter, req *http.Request) {
	var payload LoginUserRequest

	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		if errors.Is(err, io.EOF) {
			h.log.Error("json.NewDecoder.Decode", slog.Any("error", err))
			httperr.HandleError(w, ErrReqPayloadEmpty)

			return
		}

		h.log.Error("json.NewDecoder.Decode", slog.Any("error", err))
		httperr.HandleError(w, httperr.NewHTTPError(http.StatusBadRequest, err))

		return
	}

	defer req.Body.Close()

	usr, err := h.storage.GetUser(req.Context(), payload.Login)
	if err != nil {
		if errors.Is(err, userrepo.ErrUsrNotFound) {
			h.log.Error("storage.GetUser", slog.Any("error", err))
			httperr.HandleError(w, ErrUsrNotFound)

			return
		}

		h.log.Error("storage.GetUser", slog.Any("error", err))
		httperr.HandleError(w, httperr.NewHTTPError(http.StatusInternalServerError, err))

		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(usr.Password()), []byte(payload.Password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			h.log.Error("bcrypt.CompareHashAndPassword", slog.Any("error", err))
			httperr.HandleError(w, ErrUsrPassWrong)

			return
		}

		h.log.Error("bcrypt.CompareHashAndPassword", slog.Any("error", err))
		httperr.HandleError(w, httperr.NewHTTPError(http.StatusInternalServerError, err))

		return
	}

	token, err := h.auth.CreateJWTString(usr.Login())
	if err != nil {
		h.log.Error("jwtauth.CreateJWTString", slog.Any("error", err))
		httperr.HandleError(w, httperr.NewHTTPError(http.StatusInternalServerError, err))

		return
	}

	// w.Header().Set("Authorization", "Bearer "+token)
	response.JSONResponse(w, http.StatusOK, LoginUserResponse{Token: token})
}
