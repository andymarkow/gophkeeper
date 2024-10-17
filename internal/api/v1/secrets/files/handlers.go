package files

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/andymarkow/gophkeeper/internal/httperr"
	"github.com/andymarkow/gophkeeper/internal/services/filesvc"
)

// Handlers represents bankcards API handlers.
type Handlers struct {
	log     *slog.Logger
	filesvc *filesvc.Service
}

// NewHandlers returns a new Handlers instance.
func NewHandlers(svc *filesvc.Service, opts ...HandlersOpt) *Handlers {
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

	file, fileHeader, err := req.FormFile("file")
	if err != nil {
		httperr.HandleError(w, httperr.NewHTTPError(http.StatusBadRequest, err))

		return
	}
	defer file.Close()

	fmt.Println("fileHeader.Filename", fileHeader.Filename)
	fmt.Println("fileHeader.Size", fileHeader.Size)
	fmt.Println("fileHeader.Header", fileHeader.Header)
}

// func (h *Handlers) processCreateFileRequest(ctx context.Context, userID string, payload CreateFileRequest) error {
// 	return h.filesvc.CreateFile(ctx, userID, payload)
// }
