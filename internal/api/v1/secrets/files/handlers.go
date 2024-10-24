package files

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"sort"

	"github.com/go-chi/chi/v5"

	"github.com/andymarkow/gophkeeper/internal/api"
	"github.com/andymarkow/gophkeeper/internal/httperr"
	"github.com/andymarkow/gophkeeper/internal/services/filesvc"
)

// Handlers represents files API handlers.
type Handlers struct {
	log     *slog.Logger
	filesvc filesvc.Service
}

// NewHandlers returns a new Handlers instance.
func NewHandlers(svc filesvc.Service, opts ...HandlersOpt) *Handlers {
	h := &Handlers{
		log:     slog.New(&slog.JSONHandler{}),
		filesvc: svc,
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

// CreateSecret handles create file request.
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

	resp, httpErr := h.processCreateSecretRequest(req.Context(), userID, payload.Name, payload.Metadata)
	if httpErr != nil {
		httperr.HandleError(w, httpErr)

		return
	}

	api.JSONResponse(w, http.StatusCreated, resp)
}

// processCreateSecretRequest processes create file request.
func (h *Handlers) processCreateSecretRequest(ctx context.Context, userID, secretName string, metadata map[string]string,
) (*CreateSecretResponse, *httperr.HTTPError) {
	secret, err := h.filesvc.CreateSecret(ctx, userID, secretName, metadata)
	if err != nil {
		if errors.Is(err, filesvc.ErrSecretEntryAlreadyExists) {
			return nil, httperr.NewHTTPError(http.StatusConflict, err)
		}

		h.log.Error("failed to create file secret entry", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusBadRequest, err)
	}

	return &CreateSecretResponse{
		&Secret{
			ID:        secret.ID(),
			Name:      secret.Name(),
			Metadata:  secret.Metadata(),
			CreatedAt: secret.CreatedAt(),
			UpdatedAt: secret.UpdatedAt(),
			File: &File{
				Name:     secret.ContentInfo().FileName(),
				Size:     secret.ContentInfo().Size(),
				Checksum: secret.ContentInfo().Checksum(),
			},
		},
	}, nil
}

// UpdateSecret handles update file request.
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
		req.Context(), userID, secretName, payload.File.Name, payload.Metadata)
	if httpErr != nil {
		httperr.HandleError(w, httpErr)

		return
	}

	api.JSONResponse(w, http.StatusAccepted, resp)
}

// processUpdateSecretRequest processes update file request.
func (h *Handlers) processUpdateSecretRequest(ctx context.Context, userID, secretName, fileName string,
	metadata map[string]string) (*UpdateSecretResponse, *httperr.HTTPError) {
	secret, err := h.filesvc.UpdateSecret(ctx, userID, secretName, fileName, metadata)
	if err != nil {
		if errors.Is(err, filesvc.ErrSecretEntryNotFound) {
			return nil, httperr.NewHTTPError(http.StatusNotFound, err)
		}

		h.log.Error("failed to update file entry", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusBadRequest, err)
	}

	return &UpdateSecretResponse{
		&Secret{
			ID:        secret.ID(),
			Name:      secret.Name(),
			Metadata:  secret.Metadata(),
			CreatedAt: secret.CreatedAt(),
			UpdatedAt: secret.UpdatedAt(),
			File: &File{
				Name:     secret.ContentInfo().FileName(),
				Size:     secret.ContentInfo().Size(),
				Checksum: secret.ContentInfo().Checksum(),
			},
		},
	}, nil
}

// ListSecrets handles list file secrets request.
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

// processListSecretsRequest processes list files request.
func (h *Handlers) processListSecretsRequest(ctx context.Context, userID string) (*ListSecretsResponse, *httperr.HTTPError) {
	secrets, err := h.filesvc.ListSecrets(ctx, userID)
	if err != nil {
		h.log.Error("failed to list files", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusBadRequest, err)
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
			File: &File{
				Name:     secret.ContentInfo().FileName(),
				Size:     secret.ContentInfo().Size(),
				Checksum: secret.ContentInfo().Checksum(),
			},
		})
	}

	sort.Slice(resp.Secrets, func(i, j int) bool {
		return resp.Secrets[i].ID < resp.Secrets[j].ID
	})

	return resp, nil
}

// GetSecret handles get file request.
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

// processGetSecretRequest processes get file request.
func (h *Handlers) processGetSecretRequest(ctx context.Context, userID, secretName string) (*Secret, *httperr.HTTPError) {
	secret, err := h.filesvc.GetSecret(ctx, userID, secretName)
	if err != nil {
		if errors.Is(err, filesvc.ErrSecretEntryNotFound) {
			return nil, httperr.NewHTTPError(http.StatusNotFound, err)
		}

		h.log.Error("failed to get file secret entry", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusBadRequest, err)
	}

	return &Secret{
		ID:        secret.ID(),
		Name:      secret.Name(),
		Metadata:  secret.Metadata(),
		CreatedAt: secret.CreatedAt(),
		UpdatedAt: secret.UpdatedAt(),
		File: &File{
			Name:     secret.ContentInfo().FileName(),
			Size:     secret.ContentInfo().Size(),
			Checksum: secret.ContentInfo().Checksum(),
		},
	}, nil
}

// UploadSecret handles upload file request.
func (h *Handlers) UploadSecret(w http.ResponseWriter, req *http.Request) {
	userID := req.Header.Get("X-User-Id")
	if userID == "" {
		httperr.HandleError(w, httperr.ErrUsrIDHeaderEmpty)

		return
	}

	secretName := req.FormValue("secret_name")
	if secretName == "" {
		httperr.HandleError(w, ErrSecretNameEmpty)

		return
	}

	file, fileHeader, err := req.FormFile("file")
	if err != nil {
		httperr.HandleError(w, httperr.NewHTTPError(http.StatusBadRequest, err))

		return
	}
	defer file.Close()

	resp, httpErr := h.processUploadSecretRequest(req.Context(), userID, secretName, file, fileHeader)
	if httpErr != nil {
		httperr.HandleError(w, httpErr)

		return
	}

	api.JSONResponse(w, http.StatusOK, resp)
}

// processUploadSecretRequest processes upload file request.
func (h *Handlers) processUploadSecretRequest(ctx context.Context, userID, secretName string,
	file multipart.File, fileHeader *multipart.FileHeader,
) (*Secret, *httperr.HTTPError) {
	secret, err := h.filesvc.UploadSecret(ctx, userID, secretName, filesvc.UploadSecretRequest{
		FileName: fileHeader.Filename,
		Size:     -1, // Must be set explicitly to '-1'.
		Data:     file,
	})
	if err != nil {
		if errors.Is(err, filesvc.ErrSecretEntryNotFound) {
			return nil, httperr.NewHTTPError(http.StatusNotFound, err)
		}

		h.log.Error("failed to upload file secret content", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusBadRequest, err)
	}

	return &Secret{
		ID:        secret.ID(),
		Name:      secret.Name(),
		Metadata:  secret.Metadata(),
		CreatedAt: secret.CreatedAt(),
		UpdatedAt: secret.UpdatedAt(),
		File: &File{
			Name:     secret.ContentInfo().FileName(),
			Size:     secret.ContentInfo().Size(),
			Checksum: secret.ContentInfo().Checksum(),
		},
	}, nil
}

// DownloadSecret handles download file secret content request.
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

	filename, stream, httpErr := h.processDownloadSecretContentRequest(req.Context(), userID, secretName)
	if httpErr != nil {
		httperr.HandleError(w, httpErr)

		return
	}
	defer stream.Close()

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)

	_, err := io.Copy(w, stream)
	if err != nil {
		httperr.HandleError(w, httperr.NewHTTPError(http.StatusInternalServerError, err))
	}
}

// processDownloadSecretContentRequest processes download file request.
func (h *Handlers) processDownloadSecretContentRequest(ctx context.Context, userID, secretName string,
) (string, io.ReadCloser, *httperr.HTTPError) {
	secret, rd, err := h.filesvc.DownloadSecret(ctx, userID, secretName)
	if err != nil {
		if errors.Is(err, filesvc.ErrSecretEntryNotFound) {
			return "", nil, httperr.NewHTTPError(http.StatusNotFound, err)
		}

		h.log.Error("failed to download file secret content", slog.Any("error", err))

		return "", nil, httperr.NewHTTPError(http.StatusBadRequest, err)
	}

	return secret.Name(), rd, nil
}

// DeleteSecret handles delete file request.
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

// processDeleteSecretRequest processes delete the secret request.
func (h *Handlers) processDeleteSecretRequest(ctx context.Context, userID, secretName string) *httperr.HTTPError {
	err := h.filesvc.DeleteSecret(ctx, userID, secretName)
	if err != nil {
		if errors.Is(err, filesvc.ErrSecretEntryNotFound) {
			return httperr.NewHTTPError(http.StatusNotFound, err)
		}

		h.log.Error("failed to delete file secret entry", slog.Any("error", err))

		return httperr.NewHTTPError(http.StatusBadRequest, err)
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
