package texts

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"sort"

	"github.com/go-chi/chi/v5"

	"github.com/andymarkow/gophkeeper/internal/api"
	"github.com/andymarkow/gophkeeper/internal/httperr"
	"github.com/andymarkow/gophkeeper/internal/services/textsvc"
)

// Handlers represents texts API handlers.
type Handlers struct {
	log     *slog.Logger
	textsvc textsvc.Service
}

// NewHandlers returns a new Handlers instance.
func NewHandlers(svc textsvc.Service, opts ...HandlersOpt) *Handlers {
	h := &Handlers{
		log:     slog.New(&slog.JSONHandler{}),
		textsvc: svc,
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

// CreateSecret handles create text secret request.
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

// processCreateSecretRequest processes create secret request.
func (h *Handlers) processCreateSecretRequest(ctx context.Context, userID string, payload CreateSecretRequest) (*Secret, *httperr.HTTPError) {
	secret, err := h.textsvc.CreateSecret(ctx, userID, payload.Name, payload.Metadata)
	if err != nil {
		if errors.Is(err, textsvc.ErrSecretEntryAlreadyExists) {
			return nil, httperr.NewHTTPError(http.StatusConflict, err)
		}

		h.log.Error("textsvc.CreateSecret", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	return &Secret{
		ID:        secret.ID(),
		Name:      secret.Name(),
		Metadata:  secret.Metadata(),
		CreatedAt: secret.CreatedAt(),
		UpdatedAt: secret.UpdatedAt(),
		Content: &Content{
			Checksum: secret.ContentInfo().Checksum(),
		},
	}, nil
}

// ListSecrets handles list text secrets request.
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

// processListSecretsRequest processes list secrets request.
func (h *Handlers) processListSecretsRequest(ctx context.Context, userID string) (*ListSecretsResponse, *httperr.HTTPError) {
	secrets, err := h.textsvc.ListSecrets(ctx, userID)
	if err != nil {
		h.log.Error("textsvc.ListSecrets", slog.Any("error", err))

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
			Content: &Content{
				Checksum: secret.ContentInfo().Checksum(),
			},
		})
	}

	sort.Slice(resp.Secrets, func(i, j int) bool {
		return resp.Secrets[i].ID < resp.Secrets[j].ID
	})

	return resp, nil
}

// GetSecret handles get secret request.
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

// processGetSecretRequest processes get secret request.
func (h *Handlers) processGetSecretRequest(ctx context.Context, userID, secretName string) (*Secret, *httperr.HTTPError) {
	secret, err := h.textsvc.GetSecret(ctx, userID, secretName)
	if err != nil {
		if errors.Is(err, textsvc.ErrSecretEntryNotFound) {
			return nil, httperr.NewHTTPError(http.StatusNotFound, err)
		}

		h.log.Error("textsvc.GetSecret", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	return &Secret{
		ID:        secret.ID(),
		Name:      secret.Name(),
		Metadata:  secret.Metadata(),
		CreatedAt: secret.CreatedAt(),
		UpdatedAt: secret.UpdatedAt(),
		Content: &Content{
			Checksum: secret.ContentInfo().Checksum(),
		},
	}, nil
}

// UpdateSecret handles update secret request.
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

	resp, httpErr := h.processUpdateSecretRequest(
		req.Context(), userID, secretName, payload.Metadata)
	if httpErr != nil {
		httperr.HandleError(w, httpErr)

		return
	}

	api.JSONResponse(w, http.StatusAccepted, resp)
}

// processUpdateSecretRequest processes update secret request.
func (h *Handlers) processUpdateSecretRequest(ctx context.Context, userID, secretName string,
	metadata map[string]string) (*Secret, *httperr.HTTPError) {
	secret, err := h.textsvc.UpdateSecret(ctx, userID, secretName, metadata)
	if err != nil {
		if errors.Is(err, textsvc.ErrSecretEntryNotFound) {
			return nil, httperr.NewHTTPError(http.StatusNotFound, err)
		}

		h.log.Error("textsvc.UpdateSecret", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	return &Secret{
		ID:        secret.ID(),
		Name:      secret.Name(),
		Metadata:  secret.Metadata(),
		CreatedAt: secret.CreatedAt(),
		UpdatedAt: secret.UpdatedAt(),
		Content: &Content{
			Checksum: secret.ContentInfo().Checksum(),
		},
	}, nil
}

// DeleteSecret handles delete secret request.
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

// processDeleteSecretRequest processes delete secret request.
func (h *Handlers) processDeleteSecretRequest(ctx context.Context, userID, secretName string) *httperr.HTTPError {
	err := h.textsvc.DeleteSecret(ctx, userID, secretName)
	if err != nil {
		if errors.Is(err, textsvc.ErrSecretEntryNotFound) {
			return httperr.NewHTTPError(http.StatusNotFound, err)
		}

		h.log.Error("textsvc.DeleteSecret", slog.Any("error", err))

		return httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	return nil
}

func (h *Handlers) UploadSecret(w http.ResponseWriter, req *http.Request) {
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

	defer req.Body.Close()

	resp, httpErr := h.processUploadSecretRequest(req.Context(), userID, secretName, req.Body)
	if httpErr != nil {
		httperr.HandleError(w, httpErr)

		return
	}

	api.JSONResponse(w, http.StatusAccepted, resp)
}

func (h *Handlers) processUploadSecretRequest(ctx context.Context, userID, secretName string, data io.Reader) (*Secret, *httperr.HTTPError) {
	secret, err := h.textsvc.UploadSecret(ctx, userID, secretName, data)
	if err != nil {
		if errors.Is(err, textsvc.ErrSecretEntryNotFound) {
			return nil, httperr.NewHTTPError(http.StatusNotFound, err)
		}

		h.log.Error("textsvc.UploadSecret", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	return &Secret{
		ID:        secret.ID(),
		Name:      secret.Name(),
		Metadata:  secret.Metadata(),
		CreatedAt: secret.CreatedAt(),
		UpdatedAt: secret.UpdatedAt(),
		Content: &Content{
			Checksum: secret.ContentInfo().Checksum(),
		},
	}, nil
}

func (h *Handlers) DownloadSecret(w http.ResponseWriter, req *http.Request) {
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

	stream, httpErr := h.processDownloadSecretRequest(req.Context(), userID, secretName)
	if httpErr != nil {
		httperr.HandleError(w, httpErr)

		return
	}
	defer stream.Close()

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Transfer-Encoding", "chunked")

	bufWriter := bufio.NewWriterSize(w, 1024)

	if _, err := io.Copy(bufWriter, stream); err != nil {
		h.log.Error("io.Copy", slog.Any("error", err))

		httperr.HandleError(w, httperr.NewHTTPError(http.StatusInternalServerError, err))

		return
	}

	if err := bufWriter.Flush(); err != nil {
		h.log.Error("bufWriter.Flush", slog.Any("error", err))

		httperr.HandleError(w, httperr.NewHTTPError(http.StatusInternalServerError, err))

		return
	}
}

func (h *Handlers) processDownloadSecretRequest(ctx context.Context, userID, secretName string) (io.ReadCloser, *httperr.HTTPError) {
	_, stream, err := h.textsvc.DownloadSecret(ctx, userID, secretName)
	if err != nil {
		if errors.Is(err, textsvc.ErrSecretEntryNotFound) {
			return nil, httperr.NewHTTPError(http.StatusNotFound, err)
		}

		h.log.Error("textsvc.DownloadSecret", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	return stream, nil
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
