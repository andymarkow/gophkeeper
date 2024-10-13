package bankcards

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/andymarkow/gophkeeper/internal/api/v1/response"
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

// CreateCard handles create card request.
func (h *Handlers) CreateCard(w http.ResponseWriter, req *http.Request) {
	userID := req.Header.Get("X-User-Id")
	if userID == "" {
		httperr.HandleError(w, ErrUsrIDHeaderEmpty)

		return
	}

	var payload CreateCardRequest

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

	if err := h.processCreateCardRequest(req.Context(), userID, payload); err != nil {
		httperr.HandleError(w, err)

		return
	}

	response.JSONResponse(w, http.StatusCreated, &CreateCardResponse{Message: "ok"})
}

func (h *Handlers) processCreateCardRequest(ctx context.Context, userID string, payload CreateCardRequest) *httperr.HTTPError {
	data, err := bankcard.CreateData(payload.Data.Number, payload.Data.Name, payload.Data.CVV, payload.Data.ExpireAt)
	if err != nil {
		h.log.Error("failed to create bank card data", slog.Any("error", err))

		return httperr.NewHTTPError(http.StatusBadRequest, err)
	}

	encData, err := data.Encrypt(h.cryptoKey)
	if err != nil {
		h.log.Error("failed to encrypt bank card data", slog.Any("error", err))

		return httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	card, err := bankcard.CreateBankCard(payload.ID, userID, payload.Metadata, encData)
	if err != nil {
		h.log.Error("failed to create bank card", slog.Any("error", err))

		return httperr.NewHTTPError(http.StatusBadRequest, err)
	}

	if err := h.storage.AddCard(ctx, card); err != nil {
		if errors.Is(err, cardrepo.ErrCardAlreadyExists) {
			h.log.Error("failed to add bank card to storage", slog.Any("error", err))

			return httperr.NewHTTPError(http.StatusBadRequest, err)
		}

		h.log.Error("failed to add bank card to storage", slog.Any("error", err))

		return httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	return nil
}
