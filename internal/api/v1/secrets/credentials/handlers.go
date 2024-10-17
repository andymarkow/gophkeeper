package credentials

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"sort"

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

// CreateCredential handles create card request.
func (h *Handlers) CreateCredential(w http.ResponseWriter, req *http.Request) {
	userID := req.Header.Get("X-User-Id")
	if userID == "" {
		httperr.HandleError(w, httperr.ErrUsrIDHeaderEmpty)

		return
	}

	var payload CreateCredentialRequest

	if err := h.readBody(req.Body, &payload); err != nil {
		httperr.HandleError(w, err)

		return
	}

	defer req.Body.Close()

	if err := h.processCreateCredentialRequest(req.Context(), userID, payload); err != nil {
		httperr.HandleError(w, err)

		return
	}

	api.JSONResponse(w, http.StatusCreated, &CreateCredentialResponse{Message: "ok"})
}

// processCreateCredentialRequest processes create credential request.
func (h *Handlers) processCreateCredentialRequest(ctx context.Context, userID string, payload CreateCredentialRequest) *httperr.HTTPError {
	data, err := credential.NewData(payload.Data.Login, payload.Data.Password)
	if err != nil {
		h.log.Error("failed to create credentials data", slog.Any("error", err))

		return httperr.NewHTTPError(http.StatusBadRequest, err)
	}

	encData, err := data.Encrypt(h.cryptoKey)
	if err != nil {
		h.log.Error("failed to encrypt credentials data", slog.Any("error", err))

		return httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	cred, err := credential.CreateCredential(payload.ID, userID, payload.Metadata, encData)
	if err != nil {
		h.log.Error("failed to create credentials", slog.Any("error", err))

		return httperr.NewHTTPError(http.StatusBadRequest, err)
	}

	if err := h.storage.AddCredential(ctx, cred); err != nil {
		if errors.Is(err, credrepo.ErrCredAlreadyExists) {
			h.log.Error("failed to add credentials to storage", slog.Any("error", err))

			return httperr.NewHTTPError(http.StatusConflict, err)
		}

		h.log.Error("failed to add credentials to storage", slog.Any("error", err))

		return httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	return nil
}

// ListCredentials handles list credentials request.
func (h *Handlers) ListCredentials(w http.ResponseWriter, req *http.Request) {
	userID := req.Header.Get("X-User-Id")
	if userID == "" {
		httperr.HandleError(w, httperr.ErrUsrIDHeaderEmpty)

		return
	}

	resp, err := h.processListCredentialsRequest(req.Context(), userID)
	if err != nil {
		httperr.HandleError(w, err)

		return
	}

	api.JSONResponse(w, http.StatusOK, resp)
}

func (h *Handlers) GetCredential(w http.ResponseWriter, req *http.Request) {
	userID := req.Header.Get("X-User-Id")
	if userID == "" {
		httperr.HandleError(w, httperr.ErrUsrIDHeaderEmpty)

		return
	}

	credID := chi.URLParam(req, "credID")
	if credID == "" {
		httperr.HandleError(w, ErrCredIDEmpty)

		return
	}

	resp, err := h.processGetCredentialRequest(req.Context(), userID, credID)
	if err != nil {
		httperr.HandleError(w, err)

		return
	}

	api.JSONResponse(w, http.StatusOK, resp)
}

func (h *Handlers) processGetCredentialRequest(ctx context.Context, userID, credID string) (*GetCredentialResponse, *httperr.HTTPError) {
	cred, err := h.storage.GetCredential(ctx, userID, credID)
	if err != nil {
		if errors.Is(err, credrepo.ErrCredNotFound) {
			return nil, httperr.NewHTTPError(http.StatusNotFound, err)
		}

		h.log.Error("failed to get credentials from storage", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	decData, err := cred.Data().Decrypt(h.cryptoKey)
	if err != nil {
		h.log.Error("failed to decrypt credentials data", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	// Set decrypted data.
	cred.SetData(decData)

	return &GetCredentialResponse{
		Credential: Credential{
			ID:        cred.ID(),
			Metadata:  cred.Metadata(),
			CreatedAt: cred.CreatedAt(),
			UpdatedAt: cred.UpdatedAt(),
			Data: &Data{
				Login:    cred.Data().Login(),
				Password: cred.Data().Password(),
			},
		},
	}, nil
}

// processListCredentialsRequest processes list credentials request.
func (h *Handlers) processListCredentialsRequest(ctx context.Context, userID string) (*ListCredentialsResponse, *httperr.HTTPError) {
	creds, err := h.storage.ListCredentials(ctx, userID)
	if err != nil {
		h.log.Error("failed to list credentials", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	if creds == nil {
		return &ListCredentialsResponse{Creds: []*Credential{}}, nil
	}

	crds := make([]*Credential, 0, len(creds))

	for _, cred := range creds {
		crds = append(crds, &Credential{
			ID:        cred.ID(),
			Metadata:  cred.Metadata(),
			CreatedAt: cred.CreatedAt(),
			UpdatedAt: cred.UpdatedAt(),
		})
	}

	sort.Slice(crds, func(i, j int) bool {
		return crds[i].ID < crds[j].ID
	})

	return &ListCredentialsResponse{Creds: crds}, nil
}

// UpdateCredential handles update credentials request.
func (h *Handlers) UpdateCredential(w http.ResponseWriter, req *http.Request) {
	userID := req.Header.Get("X-User-Id")
	if userID == "" {
		httperr.HandleError(w, httperr.ErrUsrIDHeaderEmpty)

		return
	}

	credID := chi.URLParam(req, "credID")
	if credID == "" {
		httperr.HandleError(w, ErrCredIDEmpty)

		return
	}

	var payload UpdateCredentialRequest

	if err := h.readBody(req.Body, &payload); err != nil {
		httperr.HandleError(w, err)

		return
	}

	defer req.Body.Close()

	if err := h.processUpdateCredentialRequest(req.Context(), userID, credID, &payload); err != nil {
		httperr.HandleError(w, err)

		return
	}

	api.JSONResponse(w, http.StatusNoContent, nil)
}

func (h *Handlers) processUpdateCredentialRequest(ctx context.Context, userID, credID string, payload *UpdateCredentialRequest) *httperr.HTTPError {
	data, err := credential.NewData(payload.Data.Login, payload.Data.Password)
	if err != nil {
		h.log.Error("failed to read credentials data", slog.Any("error", err))

		return httperr.NewHTTPError(http.StatusBadRequest, err)
	}

	encData, err := data.Encrypt(h.cryptoKey)
	if err != nil {
		h.log.Error("failed to encrypt credentials data", slog.Any("error", err))

		return httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	cred, err := credential.CreateCredential(credID, userID, payload.Metadata, encData)
	if err != nil {
		h.log.Error("failed to create credentials", slog.Any("error", err))

		return httperr.NewHTTPError(http.StatusBadRequest, err)
	}

	if err := h.storage.UpdateCredential(ctx, cred); err != nil {
		if errors.Is(err, credrepo.ErrCredNotFound) {
			return httperr.NewHTTPError(http.StatusNotFound, err)
		}

		h.log.Error("failed to update credentials", slog.Any("error", err))

		return httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	return nil
}

// DeleteCredential handles delete credentials request.
func (h *Handlers) DeleteCredential(w http.ResponseWriter, req *http.Request) {
	userID := req.Header.Get("X-User-Id")
	if userID == "" {
		httperr.HandleError(w, httperr.ErrUsrIDHeaderEmpty)

		return
	}

	credID := chi.URLParam(req, "credID")
	if credID == "" {
		httperr.HandleError(w, ErrCredIDEmpty)

		return
	}

	err := h.processDeleteCredentialRequest(req.Context(), userID, credID)
	if err != nil {
		httperr.HandleError(w, err)

		return
	}

	api.JSONResponse(w, http.StatusNoContent, nil)
}

// processDeleteCredentialRequest processes delete credentials request.
func (h *Handlers) processDeleteCredentialRequest(ctx context.Context, userID, credID string) *httperr.HTTPError {
	if err := h.storage.DeleteCredential(ctx, userID, credID); err != nil {
		if errors.Is(err, credrepo.ErrCredNotFound) {
			return httperr.NewHTTPError(http.StatusNotFound, err)
		}

		h.log.Error("failed to remove credentials from storage", slog.Any("error", err))

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
