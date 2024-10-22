package credentials

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/andymarkow/gophkeeper/internal/api"
	"github.com/andymarkow/gophkeeper/internal/domain/vault/credential"
	"github.com/andymarkow/gophkeeper/internal/httperr"
	"github.com/andymarkow/gophkeeper/internal/storage/credrepo"
)

// Handlers represents bankcards API handlers.
type Handlers struct {
	log       *slog.Logger
	storage   credrepo.Storage
	cryptoKey []byte
}

// NewHandlers returns a new Handlers instance.
func NewHandlers(repo credrepo.Storage, cryptoKey []byte, opts ...HandlersOpt) *Handlers {
	h := &Handlers{
		log:       slog.New(&slog.JSONHandler{}),
		storage:   repo,
		cryptoKey: cryptoKey,
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

// HandlersOpt is a functional option type for Handlers.
type HandlersOpt func(*Handlers)

// WithHandlersLogger is an option for Handlers instance that sets logger.
func WithHandlersLogger(log *slog.Logger) HandlersOpt {
	return func(h *Handlers) {
		h.log = log
	}
}

// CreateSecret handles create credential secret request.
func (h *Handlers) CreateSecret(w http.ResponseWriter, req *http.Request) {
	userID := req.Header.Get("X-User-Id")
	if userID == "" {
		httperr.HandleError(w, httperr.ErrUsrIDHeaderEmpty)

		return
	}

	var payload CreateSecretRequest

	if err := h.readBody(req.Body, &payload); err != nil {
		httperr.HandleError(w, err)

		return
	}
	defer req.Body.Close()

	resp, httpErr := h.processCreateSecretRequest(req.Context(), userID, payload)
	if httpErr != nil {
		httperr.HandleError(w, httpErr)

		return
	}

	api.JSONResponse(w, http.StatusCreated, resp)
}

// processCreateSecretRequest processes create credential secret request.
func (h *Handlers) processCreateSecretRequest(ctx context.Context, userID string, payload CreateSecretRequest) (*Secret, *httperr.HTTPError) {
	data, err := credential.NewData(payload.Data.Login, payload.Data.Password)
	if err != nil {
		h.log.Error("failed to create credentials data", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusBadRequest, err)
	}

	encData, err := data.Encrypt(h.cryptoKey)
	if err != nil {
		h.log.Error("failed to encrypt credentials data", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	cred, err := credential.CreateSecret(payload.Name, userID, payload.Metadata, encData)
	if err != nil {
		h.log.Error("failed to create credentials", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusBadRequest, err)
	}

	secret, err := h.storage.AddSecret(ctx, cred)
	if err != nil {
		if errors.Is(err, credrepo.ErrSecretAlreadyExists) {
			return nil, httperr.NewHTTPError(http.StatusConflict, err)
		}

		h.log.Error("failed to add credentials to storage", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	return &Secret{
		ID:        secret.ID(),
		Name:      secret.Name(),
		Metadata:  secret.Metadata(),
		CreatedAt: secret.CreatedAt(),
		UpdatedAt: secret.UpdatedAt(),
		Data: &Data{
			Login:    data.Login(),
			Password: data.Password(),
		},
	}, nil
}

// ListSecrets handles list credential secrets request.
func (h *Handlers) ListSecrets(w http.ResponseWriter, req *http.Request) {
	userID := req.Header.Get("X-User-Id")
	if userID == "" {
		httperr.HandleError(w, httperr.ErrUsrIDHeaderEmpty)

		return
	}

	resp, err := h.processListSecretsRequest(req.Context(), userID)
	if err != nil {
		httperr.HandleError(w, err)

		return
	}

	api.JSONResponse(w, http.StatusOK, resp)
}

// processListSecretsRequest processes list credential secrets request.
func (h *Handlers) processListSecretsRequest(ctx context.Context, userID string) (*ListSecretsResponse, *httperr.HTTPError) {
	secrets, err := h.storage.ListSecrets(ctx, userID)
	if err != nil {
		h.log.Error("failed to list credential secrets", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	resp := &ListSecretsResponse{
		Secrets: make([]*Secret, 0, len(secrets)),
	}

	for _, secret := range secrets {
		resp.Secrets = append(resp.Secrets, &Secret{
			ID:        secret.ID(),
			Name:      secret.Name(),
			Metadata:  secret.Metadata(),
			CreatedAt: secret.CreatedAt(),
			UpdatedAt: secret.UpdatedAt(),
		})
	}

	sort.Slice(resp.Secrets, func(i, j int) bool {
		return resp.Secrets[i].ID < resp.Secrets[j].ID
	})

	return resp, nil
}

// GetSecret handles get credential secret request.
func (h *Handlers) GetSecret(w http.ResponseWriter, req *http.Request) {
	userID := req.Header.Get("X-User-Id")
	if userID == "" {
		httperr.HandleError(w, httperr.ErrUsrIDHeaderEmpty)

		return
	}

	secretName := chi.URLParam(req, "secretName")
	if secretName == "" {
		httperr.HandleError(w, ErrSecretNameEmpty)

		return
	}

	resp, err := h.processGetSecretRequest(req.Context(), userID, secretName)
	if err != nil {
		httperr.HandleError(w, err)

		return
	}

	api.JSONResponse(w, http.StatusOK, resp)
}

func (h *Handlers) processGetSecretRequest(ctx context.Context, userID, secretName string) (*Secret, *httperr.HTTPError) {
	secret, err := h.storage.GetSecret(ctx, userID, secretName)
	if err != nil {
		if errors.Is(err, credrepo.ErrSecretNotFound) {
			return nil, httperr.NewHTTPError(http.StatusNotFound, err)
		}

		h.log.Error("failed to get credential secret from storage", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	decData, err := secret.Data().Decrypt(h.cryptoKey)
	if err != nil {
		h.log.Error("failed to decrypt credentials data", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	// Set decrypted data.
	secret.SetData(decData)

	return &Secret{
		ID:        secret.ID(),
		Name:      secret.Name(),
		Metadata:  secret.Metadata(),
		CreatedAt: secret.CreatedAt(),
		UpdatedAt: secret.UpdatedAt(),
		Data: &Data{
			Login:    secret.Data().Login(),
			Password: secret.Data().Password(),
		},
	}, nil
}

// UpdateSecret handles update credential secret request.
func (h *Handlers) UpdateSecret(w http.ResponseWriter, req *http.Request) {
	userID := req.Header.Get("X-User-Id")
	if userID == "" {
		httperr.HandleError(w, httperr.ErrUsrIDHeaderEmpty)

		return
	}

	secretName := chi.URLParam(req, "secretName")
	if secretName == "" {
		httperr.HandleError(w, ErrSecretNameEmpty)

		return
	}

	var payload UpdateSecretRequest

	if err := h.readBody(req.Body, &payload); err != nil {
		httperr.HandleError(w, err)

		return
	}

	defer req.Body.Close()

	resp, httpErr := h.processUpdateSecretRequest(req.Context(), userID, secretName, payload)
	if httpErr != nil {
		httperr.HandleError(w, httpErr)

		return
	}

	api.JSONResponse(w, http.StatusAccepted, resp)
}

func (h *Handlers) processUpdateSecretRequest(ctx context.Context, userID, secretName string, payload UpdateSecretRequest) (*Secret, *httperr.HTTPError) {
	currSecret, err := h.storage.GetSecret(ctx, userID, secretName)
	if err != nil {
		if errors.Is(err, credrepo.ErrSecretNotFound) {
			return nil, httperr.NewHTTPError(http.StatusNotFound, err)
		}

		h.log.Error("failed to get credential secret entry from storage", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	data, err := credential.NewData(payload.Data.Login, payload.Data.Password)
	if err != nil {
		h.log.Error("failed to create credential secret data", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusBadRequest, err)
	}

	encData, err := data.Encrypt(h.cryptoKey)
	if err != nil {
		h.log.Error("failed to encrypt credential secret data", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	currSecret.AddMetadata(payload.Metadata)

	secretObj, err := credential.NewSecret(
		currSecret.ID(), secretName, userID, currSecret.Metadata(), currSecret.CreatedAt(), time.Now(), encData)
	if err != nil {
		h.log.Error("failed to create credential secret", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusBadRequest, err)
	}

	secret, err := h.storage.UpdateSecret(ctx, secretObj)
	if err != nil {
		if errors.Is(err, credrepo.ErrSecretNotFound) {
			return nil, httperr.NewHTTPError(http.StatusNotFound, err)
		}

		h.log.Error("failed to update credential secret", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	decData, err := secret.Data().Decrypt(h.cryptoKey)
	if err != nil {
		h.log.Error("failed to decrypt credential secret data", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	return &Secret{
		ID:        secret.ID(),
		Name:      secret.Name(),
		Metadata:  secret.Metadata(),
		CreatedAt: secret.CreatedAt(),
		UpdatedAt: secret.UpdatedAt(),
		Data: &Data{
			Login:    decData.Login(),
			Password: decData.Password(),
		},
	}, nil
}

// DeleteSecret handles delete credential secret request.
func (h *Handlers) DeleteSecret(w http.ResponseWriter, req *http.Request) {
	userID := req.Header.Get("X-User-Id")
	if userID == "" {
		httperr.HandleError(w, httperr.ErrUsrIDHeaderEmpty)

		return
	}

	secretName := chi.URLParam(req, "secretName")
	if secretName == "" {
		httperr.HandleError(w, ErrSecretNameEmpty)

		return
	}

	httpErr := h.processDeleteSecretRequest(req.Context(), userID, secretName)
	if httpErr != nil {
		httperr.HandleError(w, httpErr)

		return
	}

	api.JSONResponse(w, http.StatusNoContent, nil)
}

// processDeleteSecretRequest processes delete credential secret request.
func (h *Handlers) processDeleteSecretRequest(ctx context.Context, userID, secretName string) *httperr.HTTPError {
	if err := h.storage.DeleteSecret(ctx, userID, secretName); err != nil {
		if errors.Is(err, credrepo.ErrSecretNotFound) {
			return httperr.NewHTTPError(http.StatusNotFound, err)
		}

		h.log.Error("failed to delete credential secret from storage", slog.Any("error", err))

		return httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	return nil
}

func (h *Handlers) readBody(body io.ReadCloser, v any) *httperr.HTTPError {
	err := json.NewDecoder(body).Decode(v)
	if err != nil {
		if errors.Is(err, io.EOF) {
			h.log.Error("json.NewDecoder.Decode", slog.Any("error", err))

			return httperr.ErrReqPayloadEmpty
		}

		h.log.Error("json.NewDecoder.Decode", slog.Any("error", err))

		return httperr.NewHTTPError(http.StatusBadRequest, err)
	}

	return nil
}
