package bankcards

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
	"github.com/andymarkow/gophkeeper/internal/domain/vault/bankcard"
	"github.com/andymarkow/gophkeeper/internal/httperr"
	"github.com/andymarkow/gophkeeper/internal/storage/cardrepo"
)

// Handlers represents bankcards API handlers.
type Handlers struct {
	log       *slog.Logger
	storage   cardrepo.Storage
	cryptoKey []byte
}

// NewHandlers returns a new Handlers instance.
func NewHandlers(repo cardrepo.Storage, cryptoKey []byte, opts ...Option) *Handlers {
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

// Option is a functional option type for Handlers.
type Option func(*Handlers)

// WithLogger is an option for Handlers instance that sets logger.
func WithLogger(log *slog.Logger) Option {
	return func(h *Handlers) {
		h.log = log
	}
}

// CreateSecret handles create bank card secret request.
func (h *Handlers) CreateSecret(w http.ResponseWriter, req *http.Request) {
	userID := req.Header.Get("X-User-Id")
	if userID == "" {
		httperr.HandleError(w, httperr.ErrUsrIDHeaderEmpty)

		return
	}

	var payload CreateSecretRequest

	if httpErr := h.readBody(req.Body, &payload); httpErr != nil {
		httperr.HandleError(w, httpErr)

		return
	}
	defer req.Body.Close()

	resp, err := h.processCreateSecretRequest(req.Context(), userID, payload)
	if err != nil {
		httperr.HandleError(w, err)

		return
	}

	api.JSONResponse(w, http.StatusCreated, resp)
}

// processCreateSecretRequest processes create bank card secret request.
func (h *Handlers) processCreateSecretRequest(ctx context.Context, userID string, payload CreateSecretRequest) (*Secret, *httperr.HTTPError) {
	data, err := bankcard.CreateData(payload.Data.Number, payload.Data.Name, payload.Data.CVV, payload.Data.ExpireAt)
	if err != nil {
		h.log.Error("failed to create bank card secret data", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusBadRequest, err)
	}

	encData, err := data.Encrypt(h.cryptoKey)
	if err != nil {
		h.log.Error("failed to encrypt bank card secret data", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	secret, err := bankcard.CreateSecret(payload.Name, userID, payload.Metadata, encData)
	if err != nil {
		h.log.Error("failed to create bank card secret", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusBadRequest, err)
	}

	secr, err := h.storage.AddSecret(ctx, secret)
	if err != nil {
		if errors.Is(err, cardrepo.ErrSecretAlreadyExists) {
			return nil, httperr.NewHTTPError(http.StatusConflict, err)
		}

		h.log.Error("failed to add bank card secret to storage", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	return &Secret{
		ID:        secr.ID(),
		Name:      secr.Name(),
		Metadata:  secr.Metadata(),
		CreatedAt: secr.CreatedAt(),
		UpdatedAt: secr.UpdatedAt(),
		Data: &Data{
			Number:   data.Number(),
			Name:     data.Name(),
			ExpireAt: data.ExpireAt(),
			CVV:      data.CVV(),
		},
	}, nil
}

// GetSecret handles get bank card secret request.
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

// processGetSecretRequest processes get bank card secret request.
func (h *Handlers) processGetSecretRequest(ctx context.Context, userID, secretName string) (*Secret, *httperr.HTTPError) {
	secret, err := h.storage.GetSecret(ctx, userID, secretName)
	if err != nil {
		if errors.Is(err, cardrepo.ErrSecretNotFound) {
			return nil, httperr.NewHTTPError(http.StatusNotFound, err)
		}

		h.log.Error("failed to get bank card secret entry from storage", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	decData, err := secret.Data().Decrypt(h.cryptoKey)
	if err != nil {
		h.log.Error("failed to decrypt bank card secret data", slog.Any("error", err))

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
			Number:   secret.Data().Number(),
			Name:     secret.Data().Name(),
			CVV:      secret.Data().CVV(),
			ExpireAt: secret.Data().ExpireAt(),
		},
	}, nil
}

// ListSecrets handles list bank card secrets request.
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

// processListSecretsRequest processes list cards request.
func (h *Handlers) processListSecretsRequest(ctx context.Context, userID string) (*ListSecretsResponse, *httperr.HTTPError) {
	secrets, err := h.storage.ListSecrets(ctx, userID)
	if err != nil {
		h.log.Error("failed to list bank card secret entries from the storage", slog.Any("error", err))

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

// UpdateSecret handles update bank card secret request.
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

	resp, err := h.processUpdateSecretRequest(req.Context(), userID, secretName, &payload)
	if err != nil {
		httperr.HandleError(w, err)

		return
	}

	api.JSONResponse(w, http.StatusAccepted, resp)
}

// processUpdateSecretRequest processes update bank card secret request.
func (h *Handlers) processUpdateSecretRequest(ctx context.Context, userID, secretName string, payload *UpdateSecretRequest) (*Secret, *httperr.HTTPError) {
	currSecret, err := h.storage.GetSecret(ctx, userID, secretName)
	if err != nil {
		if errors.Is(err, cardrepo.ErrSecretNotFound) {
			return nil, httperr.NewHTTPError(http.StatusNotFound, err)
		}

		h.log.Error("failed to get bank card secret entry from storage", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	// Read bank card data from the payload.
	data, err := bankcard.CreateData(payload.Data.Number, payload.Data.Name, payload.Data.CVV, payload.Data.ExpireAt)
	if err != nil {
		h.log.Error("failed to create bank card secret data", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusBadRequest, err)
	}

	// Encrypt bank card data from the payload.
	encData, err := data.Encrypt(h.cryptoKey)
	if err != nil {
		h.log.Error("failed to encrypt bank card secret data", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	currSecret.AddMetadata(payload.Metadata)

	secret, err := bankcard.NewSecret(currSecret.ID(), secretName, userID, currSecret.Metadata(), currSecret.CreatedAt(), time.Now(), encData)
	if err != nil {
		h.log.Error("failed to create bank card secret", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusBadRequest, err)
	}

	secr, err := h.storage.UpdateSecret(ctx, secret)
	if err != nil {
		if errors.Is(err, cardrepo.ErrSecretNotFound) {
			h.log.Error("failed to update bank card secret entry in storage", slog.Any("error", err))

			return nil, httperr.NewHTTPError(http.StatusNotFound, err)
		}

		h.log.Error("failed to update bank card entry in storage", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	return &Secret{
		ID:        secr.ID(),
		Name:      secr.Name(),
		Metadata:  secr.Metadata(),
		CreatedAt: secr.CreatedAt(),
		UpdatedAt: secr.UpdatedAt(),
		Data: &Data{
			Name:     data.Name(),
			Number:   data.Number(),
			CVV:      data.CVV(),
			ExpireAt: data.ExpireAt(),
		},
	}, nil
}

// DeleteSecret handles delete bank card secret request.
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

	err := h.processDeleteSecretRequest(req.Context(), userID, secretName)
	if err != nil {
		httperr.HandleError(w, err)

		return
	}

	api.JSONResponse(w, http.StatusNoContent, nil)
}

// processDeleteSecretRequest processes delete card request.
func (h *Handlers) processDeleteSecretRequest(ctx context.Context, userID, secretName string) *httperr.HTTPError {
	err := h.storage.DeleteSecret(ctx, userID, secretName)
	if err != nil {
		if errors.Is(err, cardrepo.ErrSecretNotFound) {
			return httperr.NewHTTPError(http.StatusNotFound, err)
		}

		h.log.Error("failed to delete bank card secret entry from storage", slog.Any("error", err))

		return httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	return nil
}

// readBody reads request body.
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
