package cardrepo

import (
	"context"
	"fmt"
	"sync"

	"github.com/andymarkow/gophkeeper/internal/domain/vault/bankcard"
)

var _ Storage = (*InMemory)(nil)

// InMemory represents in-memory bank cards storage.
type InMemory struct {
	// UserID -> SecretName -> Secret.
	cards map[string]map[string]bankcard.Bankcard

	mu sync.RWMutex
}

// NewInMemory creates new in-memory bank cards storage.
func NewInMemory() *InMemory {
	return &InMemory{
		cards: make(map[string]map[string]bankcard.Bankcard),
	}
}

// AddCard adds a new bank card to the storage.
func (s *InMemory) AddCard(_ context.Context, card *bankcard.Bankcard) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the user login entry exists in the storage.
	cards, ok := s.cards[card.UserID()]
	if !ok {
		// Check if cards entry is nil.
		if cards == nil {
			// Initialize cards entry.
			s.cards[card.UserID()] = make(map[string]bankcard.Bankcard)
		}

		// UserID does not exist in the storage. Add user login and bank card to the storage.
		s.cards[card.UserID()][card.ID()] = *card

		return nil
	}

	if _, ok := cards[card.ID()]; ok {
		// Bank card already exists in the storage.
		return fmt.Errorf("%w: %s", ErrCardAlreadyExists, card.ID())
	}

	// Add bank card to the storage.
	s.cards[card.UserID()][card.ID()] = *card

	return nil
}

// GetCard returns a bank card from the storage.
func (s *InMemory) GetCard(_ context.Context, userLogin, cardID string) (*bankcard.Bankcard, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if the user login entry exists in the storage.
	cards, ok := s.cards[userLogin]
	if !ok {
		return nil, fmt.Errorf("%w for user login %s: %s", ErrCardNotFound, userLogin, cardID)
	}

	// Check if the bank card entry exists in the storage.
	if card, ok := cards[cardID]; ok {
		crd := card

		return &crd, nil
	}

	return nil, fmt.Errorf("%w for user login %s: %s", ErrCardNotFound, userLogin, cardID)
}

// ListCards returns a list of bank card IDs from the storage.
func (s *InMemory) ListCards(_ context.Context, userLogin string) ([]*bankcard.Bankcard, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if the user login entry exists in the storage.
	cards, ok := s.cards[userLogin]
	if !ok {
		return []*bankcard.Bankcard{}, nil
	}

	cardsList := make([]*bankcard.Bankcard, 0, len(cards))

	for _, card := range cards {
		crd := card
		cardsList = append(cardsList, &crd)
	}

	return cardsList, nil
}

// UpdateCard updates a bank card in the storage.
func (s *InMemory) UpdateCard(_ context.Context, card *bankcard.Bankcard) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the user login entry exists in the storage.
	cards, ok := s.cards[card.UserID()]
	if !ok {
		return fmt.Errorf("%w for user login %s: %s", ErrCardNotFound, card.UserID(), card.ID())
	}

	if _, ok := cards[card.ID()]; !ok {
		return fmt.Errorf("%w for user login %s: %s", ErrCardNotFound, card.UserID(), card.ID())
	}

	// Update bank card in the storage.
	s.cards[card.UserID()][card.ID()] = *card

	return nil
}

// DeleteCard deletes a bank card from the storage.
func (s *InMemory) DeleteCard(_ context.Context, userLogin, cardID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the user login entry exists in the storage.
	cards, ok := s.cards[userLogin]
	if !ok {
		return fmt.Errorf("%w for user login %s: %s", ErrCardNotFound, userLogin, cardID)
	}

	if _, ok := cards[cardID]; !ok {
		return fmt.Errorf("%w for user login %s: %s", ErrCardNotFound, userLogin, cardID)
	}

	// Delete bank card from the storage.
	delete(s.cards[userLogin], cardID)

	return nil
}
