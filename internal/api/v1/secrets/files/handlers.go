package files

import (
	"context"
	"encoding/json"
	"log/slog"
	"mime/multipart"
	"net/http"

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

	resourceID := req.FormValue("resource_id")
	if resourceID == "" {
		httperr.HandleError(w, ErrResIDEmpty)

		return
	}

	checksum := req.FormValue("checksum")
	if checksum == "" {
		httperr.HandleError(w, ErrChecksumEmpty)

		return
	}

	var metadata map[string]string

	metaStr := req.FormValue("metadata")
	if metaStr != "" {
		err := json.Unmarshal([]byte(metaStr), &metadata)
		if err != nil {
			httperr.HandleError(w, httperr.NewHTTPError(http.StatusBadRequest, err))

			return
		}
	}

	file, fileHeader, err := req.FormFile("file")
	if err != nil {
		httperr.HandleError(w, httperr.NewHTTPError(http.StatusBadRequest, err))

		return
	}
	defer file.Close()

	resp, httpErr := h.processCreateFileRequest(req.Context(), userID, resourceID, checksum, metadata, file, fileHeader)
	if httpErr != nil {
		httperr.HandleError(w, httpErr)

		return
	}

	api.JSONResponse(w, http.StatusCreated, resp)
}

func (h *Handlers) processCreateFileRequest(ctx context.Context, userID, resourceID, checksum string,
	metadata map[string]string, file multipart.File, fileHeader *multipart.FileHeader,
) (*CreateFileResponse, *httperr.HTTPError) {
	f, err := h.filesvc.CreateFile(ctx, userID, resourceID, filesvc.CreateFileRequest{
		Name:     fileHeader.Filename,
		Checksum: checksum,
		Size:     -1, // Must be set explicitly to '-1'.
		Metadata: metadata,
		Data:     file,
	})
	if err != nil {
		h.log.Error("failed to create file", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusBadRequest, err)
	}

	return &CreateFileResponse{
		&File{
			ID:        f.ID(),
			Name:      f.Name(),
			Checksum:  f.Checksum(),
			Location:  f.Location().String(),
			Size:      f.Size(),
			Metadata:  f.Metadata(),
			CreatedAt: f.CreatedAt(),
			UpdatedAt: f.UpdatedAt(),
		},
	}, nil
}

// func (h *Handlers) readBody(body io.ReadCloser, v any) *httperr.HTTPError {
// 	err := json.NewDecoder(body).Decode(v)
// 	if err != nil {
// 		if errors.Is(err, io.EOF) {
// 			h.log.Error("json.NewDecoder.Decode", slog.Any("error", err))

// 			return httperr.ErrReqPayloadEmpty
// 		}

// 		h.log.Error("json.NewDecoder.Decode", slog.Any("error", err))

// 		return httperr.NewHTTPError(http.StatusBadRequest, err)
// 	}

// 	return nil
// }
