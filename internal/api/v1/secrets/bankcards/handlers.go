package bankcards

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
		httperr.HandleError(w, httperr.ErrUsrIDHeaderEmpty)

		return
	}

	var payload CreateCardRequest

	if err := h.readBody(req.Body, &payload); err != nil {
		httperr.HandleError(w, err)

		return
	}

	defer req.Body.Close()

	if err := h.processCreateCardRequest(req.Context(), userID, payload); err != nil {
		httperr.HandleError(w, err)

		return
	}

	api.JSONResponse(w, http.StatusCreated, &CreateCardResponse{Message: "ok"})
}

// processCreateCardRequest processes create card request.
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

	card, err := bankcard.CreateBankcard(payload.ID, userID, payload.Metadata, encData)
	if err != nil {
		h.log.Error("failed to create bank card", slog.Any("error", err))

		return httperr.NewHTTPError(http.StatusBadRequest, err)
	}

	if err := h.storage.AddCard(ctx, card); err != nil {
		if errors.Is(err, cardrepo.ErrCardAlreadyExists) {
			h.log.Error("failed to add bank card to storage", slog.Any("error", err))

			return httperr.NewHTTPError(http.StatusConflict, err)
		}

		h.log.Error("failed to add bank card to storage", slog.Any("error", err))

		return httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	return nil
}

// GetCard handles get card request.
func (h *Handlers) GetCard(w http.ResponseWriter, req *http.Request) {
	userID := req.Header.Get("X-User-Id")
	if userID == "" {
		httperr.HandleError(w, httperr.ErrUsrIDHeaderEmpty)

		return
	}

	cardID := chi.URLParam(req, "cardID")
	if cardID == "" {
		httperr.HandleError(w, ErrCardIDEmpty)

		return
	}

	resp, err := h.processGetCardRequest(req.Context(), userID, cardID)
	if err != nil {
		httperr.HandleError(w, err)

		return
	}

	api.JSONResponse(w, http.StatusOK, resp)
}

// processGetCardRequest processes get card request.
func (h *Handlers) processGetCardRequest(ctx context.Context, userID, cardID string) (*GetCardResponse, *httperr.HTTPError) {
	card, err := h.storage.GetCard(ctx, userID, cardID)
	if err != nil {
		if errors.Is(err, cardrepo.ErrCardNotFound) {
			return nil, httperr.NewHTTPError(http.StatusNotFound, err)
		}

		h.log.Error("failed to get bank card from storage", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	decData, err := card.Data().Decrypt(h.cryptoKey)
	if err != nil {
		h.log.Error("failed to decrypt bank card data", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	// Set decrypted data.
	card.SetData(decData)

	return &GetCardResponse{
		BankCard: BankCard{
			ID:        card.ID(),
			Metadata:  card.Metadata(),
			CreatedAt: card.CreatedAt(),
			UpdatedAt: card.UpdatedAt(),
			Data: &Data{
				Number:   card.Data().Number(),
				Name:     card.Data().Name(),
				CVV:      card.Data().CVV(),
				ExpireAt: card.Data().ExpireAt(),
			},
		},
	}, nil
}

// ListCards handles list cards request.
func (h *Handlers) ListCards(w http.ResponseWriter, req *http.Request) {
	userID := req.Header.Get("X-User-Id")
	if userID == "" {
		httperr.HandleError(w, httperr.ErrUsrIDHeaderEmpty)

		return
	}

	resp, err := h.processListCardsRequest(req.Context(), userID)
	if err != nil {
		httperr.HandleError(w, err)

		return
	}

	api.JSONResponse(w, http.StatusOK, resp)
}

// processListCardsRequest processes list cards request.
func (h *Handlers) processListCardsRequest(ctx context.Context, userID string) (*ListCardsResponse, *httperr.HTTPError) {
	cards, err := h.storage.ListCards(ctx, userID)
	if err != nil {
		h.log.Error("failed to list bank cards from storage", slog.Any("error", err))

		return nil, httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	if cards == nil {
		return &ListCardsResponse{Cards: []*BankCard{}}, nil
	}

	crds := make([]*BankCard, 0, len(cards))

	for _, card := range cards {
		crds = append(crds, &BankCard{
			ID:        card.ID(),
			Metadata:  card.Metadata(),
			CreatedAt: card.CreatedAt(),
			UpdatedAt: card.UpdatedAt(),
		})
	}

	sort.Slice(crds, func(i, j int) bool {
		return crds[i].ID < crds[j].ID
	})

	return &ListCardsResponse{Cards: crds}, nil
}

// UpdateCard handles update card request.
func (h *Handlers) UpdateCard(w http.ResponseWriter, req *http.Request) {
	userID := req.Header.Get("X-User-Id")
	if userID == "" {
		httperr.HandleError(w, httperr.ErrUsrIDHeaderEmpty)

		return
	}

	cardID := chi.URLParam(req, "cardID")
	if cardID == "" {
		httperr.HandleError(w, ErrCardIDEmpty)

		return
	}

	var payload UpdateCardRequest

	if err := h.readBody(req.Body, &payload); err != nil {
		httperr.HandleError(w, err)

		return
	}

	defer req.Body.Close()

	err := h.processUpdateCardRequest(req.Context(), userID, cardID, &payload)
	if err != nil {
		httperr.HandleError(w, err)

		return
	}

	api.JSONResponse(w, http.StatusNoContent, nil)
}

// processUpdateCardRequest processes update card request.
func (h *Handlers) processUpdateCardRequest(ctx context.Context, userID, cardID string, payload *UpdateCardRequest) *httperr.HTTPError {
	// Read bank card data from the payload.
	data, err := bankcard.NewData(payload.Data.Number, payload.Data.Name, payload.Data.CVV, payload.Data.ExpireAt)
	if err != nil {
		h.log.Error("failed to read bank card data", slog.Any("error", err))

		return httperr.NewHTTPError(http.StatusBadRequest, err)
	}

	// Encrypt bank card data from the payload.
	encData, err := data.Encrypt(h.cryptoKey)
	if err != nil {
		h.log.Error("failed to encrypt bank card data", slog.Any("error", err))

		return httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	card, err := bankcard.CreateBankcard(cardID, userID, payload.Metadata, encData)
	if err != nil {
		h.log.Error("failed to create bank card", slog.Any("error", err))

		return httperr.NewHTTPError(http.StatusBadRequest, err)
	}

	if err := h.storage.UpdateCard(ctx, card); err != nil {
		if errors.Is(err, cardrepo.ErrCardNotFound) {
			h.log.Error("failed to get bank card from storage", slog.Any("error", err))

			return httperr.NewHTTPError(http.StatusNotFound, err)
		}

		h.log.Error("failed to update bank card in storage", slog.Any("error", err))

		return httperr.NewHTTPError(http.StatusInternalServerError, err)
	}

	return nil
}

// DeleteCard handles delete card request.
func (h *Handlers) DeleteCard(w http.ResponseWriter, req *http.Request) {
	userID := req.Header.Get("X-User-Id")
	if userID == "" {
		httperr.HandleError(w, httperr.ErrUsrIDHeaderEmpty)

		return
	}

	cardID := chi.URLParam(req, "cardID")
	if cardID == "" {
		httperr.HandleError(w, ErrCardIDEmpty)

		return
	}

	err := h.processDeleteCardRequest(req.Context(), userID, cardID)
	if err != nil {
		httperr.HandleError(w, err)

		return
	}

	api.JSONResponse(w, http.StatusNoContent, nil)
}

// processDeleteCardRequest processes delete card request.
func (h *Handlers) processDeleteCardRequest(ctx context.Context, userID, cardID string) *httperr.HTTPError {
	err := h.storage.DeleteCard(ctx, userID, cardID)
	if err != nil {
		if errors.Is(err, cardrepo.ErrCardNotFound) {
			return httperr.NewHTTPError(http.StatusNotFound, err)
		}

		h.log.Error("failed to delete bank card from storage", slog.Any("error", err))

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
