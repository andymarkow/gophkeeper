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

// Handlers represents bankcards API handlers.
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

func (h *Handlers) CreateFile(w http.ResponseWriter, req *http.Request) {
	userID := req.Header.Get("X-User-Id")
	if userID == "" {
		httperr.HandleError(w, httperr.ErrUsrIDHeaderEmpty)

		return
	}

	var payload CreateFileRequest
	if err := h.readBody(req.Body, &payload); err != nil {
		httperr.HandleError(w, err)

		return
	}
	defer req.Body.Close()

	resp, httpErr := h.processCreateFileRequest(req.Context(), userID, payload.ID, payload.Metadata)
	if httpErr != nil {
		httperr.HandleError(w, httpErr)

		return
	}

	api.JSONResponse(w, http.StatusCreated, resp)
}

func (h *Handlers) processCreateFileRequest(ctx context.Context, userID, fileID string, metadata map[string]string,
) (*CreateFileResponse, *httperr.HTTPError) {
	f, err := h.filesvc.CreateFile(ctx, userID, fileID, metadata)
	if err != nil {
		h.log.Error("failed to create file entry", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusBadRequest, err)
	}

	return &CreateFileResponse{
		&File{
			ID:        f.ID(),
			Name:      f.Name(),
			Checksum:  f.Checksum(),
			Size:      f.Size(),
			Metadata:  f.Metadata(),
			CreatedAt: f.CreatedAt(),
			UpdatedAt: f.UpdatedAt(),
		},
	}, nil
}

func (h *Handlers) UpdateFile(w http.ResponseWriter, req *http.Request) {
	userID := req.Header.Get("X-User-Id")
	if userID == "" {
		httperr.HandleError(w, httperr.ErrUsrIDHeaderEmpty)

		return
	}

	fileID := chi.URLParam(req, "fileID")
	if fileID == "" {
		httperr.HandleError(w, ErrFileIDEmpty)

		return
	}

	var payload UpdateFileRequest
	if err := h.readBody(req.Body, &payload); err != nil {
		httperr.HandleError(w, err)

		return
	}
	defer req.Body.Close()

	resp, httpErr := h.processUpdateFileRequest(req.Context(), userID, fileID, payload.Name, payload.Metadata)
	if httpErr != nil {
		httperr.HandleError(w, httpErr)

		return
	}

	api.JSONResponse(w, http.StatusAccepted, resp)
}

func (h *Handlers) processUpdateFileRequest(ctx context.Context, userID, fileID, fileName string, metadata map[string]string,
) (*UpdateFileResponse, *httperr.HTTPError) {
	f, err := h.filesvc.UpdateFile(ctx, userID, fileID, fileName, metadata)
	if err != nil {
		h.log.Error("failed to update file entry", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusBadRequest, err)
	}

	return &UpdateFileResponse{
		&File{
			ID:        f.ID(),
			Name:      f.Name(),
			Checksum:  f.Checksum(),
			Size:      f.Size(),
			Metadata:  f.Metadata(),
			CreatedAt: f.CreatedAt(),
			UpdatedAt: f.UpdatedAt(),
		},
	}, nil
}

func (h *Handlers) ListFiles(w http.ResponseWriter, req *http.Request) {
	userID := req.Header.Get("X-User-Id")
	if userID == "" {
		httperr.HandleError(w, httperr.ErrUsrIDHeaderEmpty)

		return
	}

	resp, err := h.processListFilesRequest(req.Context(), userID)
	if err != nil {
		httperr.HandleError(w, err)

		return
	}

	api.JSONResponse(w, http.StatusOK, resp)
}

func (h *Handlers) processListFilesRequest(ctx context.Context, userID string) (*ListFilesResponse, *httperr.HTTPError) {
	files, err := h.filesvc.ListFiles(ctx, userID)
	if err != nil {
		h.log.Error("failed to list files", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusBadRequest, err)
	}

	resp := &ListFilesResponse{
		Files: make([]*File, 0, len(files)),
	}

	for _, f := range files {
		resp.Files = append(resp.Files, &File{
			ID:        f.ID(),
			Name:      f.Name(),
			Checksum:  f.Checksum(),
			Size:      f.Size(),
			Metadata:  f.Metadata(),
			CreatedAt: f.CreatedAt(),
			UpdatedAt: f.UpdatedAt(),
		})
	}

	sort.Slice(resp.Files, func(i, j int) bool {
		return resp.Files[i].ID < resp.Files[j].ID
	})

	return resp, nil
}

func (h *Handlers) GetFile(w http.ResponseWriter, req *http.Request) {
	userID := req.Header.Get("X-User-Id")
	if userID == "" {
		httperr.HandleError(w, httperr.ErrUsrIDHeaderEmpty)

		return
	}

	fileID := chi.URLParam(req, "fileID")
	if fileID == "" {
		httperr.HandleError(w, ErrFileIDEmpty)

		return
	}

	resp, err := h.processGetFileRequest(req.Context(), userID, fileID)
	if err != nil {
		httperr.HandleError(w, err)

		return
	}

	api.JSONResponse(w, http.StatusOK, resp)
}

func (h *Handlers) processGetFileRequest(ctx context.Context, userID, fileID string) (*File, *httperr.HTTPError) {
	f, err := h.filesvc.GetFile(ctx, userID, fileID)
	if err != nil {
		if errors.Is(err, filesvc.ErrFileEntryNotFound) {
			return nil, httperr.NewHTTPError(http.StatusNotFound, err)
		}

		h.log.Error("failed to get file", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusBadRequest, err)
	}

	return &File{
		ID:        f.ID(),
		Name:      f.Name(),
		Checksum:  f.Checksum(),
		Size:      f.Size(),
		Metadata:  f.Metadata(),
		CreatedAt: f.CreatedAt(),
		UpdatedAt: f.UpdatedAt(),
	}, nil
}

func (h *Handlers) UploadFile(w http.ResponseWriter, req *http.Request) {
	userID := req.Header.Get("X-User-Id")
	if userID == "" {
		httperr.HandleError(w, httperr.ErrUsrIDHeaderEmpty)

		return
	}

	fileID := req.FormValue("file_id")
	if fileID == "" {
		httperr.HandleError(w, ErrFileIDEmpty)

		return
	}

	file, fileHeader, err := req.FormFile("file")
	if err != nil {
		httperr.HandleError(w, httperr.NewHTTPError(http.StatusBadRequest, err))

		return
	}
	defer file.Close()

	resp, httpErr := h.processUploadFileRequest(req.Context(), userID, fileID, file, fileHeader)
	if httpErr != nil {
		httperr.HandleError(w, httpErr)

		return
	}

	api.JSONResponse(w, http.StatusOK, resp)
}

func (h *Handlers) processUploadFileRequest(ctx context.Context, userID, fileID string,
	file multipart.File, fileHeader *multipart.FileHeader,
) (*UploadFileResponse, *httperr.HTTPError) {
	f, err := h.filesvc.UploadFile(ctx, userID, fileID, filesvc.UploadFileRequest{
		Name: fileHeader.Filename,
		Size: -1, // Must be set explicitly to '-1'.
		Data: file,
	})
	if err != nil {
		h.log.Error("failed to upload file", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusBadRequest, err)
	}

	return &UploadFileResponse{
		&File{
			ID:        f.ID(),
			Name:      f.Name(),
			Checksum:  f.Checksum(),
			Size:      f.Size(),
			Metadata:  f.Metadata(),
			CreatedAt: f.CreatedAt(),
			UpdatedAt: f.UpdatedAt(),
		},
	}, nil
}

func (h *Handlers) DownloadFile(w http.ResponseWriter, req *http.Request) {
	userID := req.Header.Get("X-User-Id")
	if userID == "" {
		httperr.HandleError(w, httperr.ErrUsrIDHeaderEmpty)

		return
	}

	fileID := chi.URLParam(req, "fileID")
	if fileID == "" {
		httperr.HandleError(w, ErrFileIDEmpty)

		return
	}

	filename, stream, httpErr := h.processDownloadFileRequest(req.Context(), userID, fileID)
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

func (h *Handlers) processDownloadFileRequest(ctx context.Context, userID, fileID string) (string, io.ReadCloser, *httperr.HTTPError) {
	f, rd, err := h.filesvc.DownloadFile(ctx, userID, fileID)
	if err != nil {
		if errors.Is(err, filesvc.ErrFileObjectNotFound) {
			return "", nil, httperr.NewHTTPError(http.StatusNotFound, err)
		}

		h.log.Error("failed to download file", slog.Any("error", err))

		return "", nil, httperr.NewHTTPError(http.StatusBadRequest, err)
	}

	return f.Name(), rd, nil
}

func (h *Handlers) DeleteFile(w http.ResponseWriter, req *http.Request) {
	userID := req.Header.Get("X-User-Id")
	if userID == "" {
		httperr.HandleError(w, httperr.ErrUsrIDHeaderEmpty)

		return
	}

	fileID := chi.URLParam(req, "fileID")
	if fileID == "" {
		httperr.HandleError(w, ErrFileIDEmpty)

		return
	}

	err := h.processDeleteFileRequest(req.Context(), userID, fileID)
	if err != nil {
		httperr.HandleError(w, err)

		return
	}

	api.JSONResponse(w, http.StatusNoContent, nil)
}

func (h *Handlers) processDeleteFileRequest(ctx context.Context, userID, fileID string) *httperr.HTTPError {
	err := h.filesvc.DeleteFile(ctx, userID, fileID)
	if err != nil {
		if errors.Is(err, filesvc.ErrFileEntryNotFound) {
			return httperr.NewHTTPError(http.StatusNotFound, err)
		}

		h.log.Error("failed to delete file", slog.Any("error", err))

		return httperr.NewHTTPError(http.StatusBadRequest, err)
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
