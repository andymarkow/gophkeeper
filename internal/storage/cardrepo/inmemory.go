// Package cardrepo provides bank cards storage implementation.
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
	// UserLogin -> CardNumber -> BankCard
	cards map[string]map[string]*bankcard.BankCard

	mu sync.RWMutex
}

// NewInMemory creates new in-memory bank cards storage.
func NewInMemory() *InMemory {
	return &InMemory{
		cards: make(map[string]map[string]*bankcard.BankCard),
	}
}

// Add adds a new bank card to the storage.
func (s *InMemory) Add(_ context.Context, card *bankcard.BankCard) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the user login entry exists in the storage.
	cards, ok := s.cards[card.UserLogin()]
	if !ok {
		// UserLogin does not exist in the storage. Add user login and bank card to the storage.
		s.cards[card.UserLogin()][card.ID()] = card

		return nil
	}

	if _, ok := cards[card.ID()]; ok {
		// Bank card already exists in the storage.
		return fmt.Errorf("%w: %s", ErrCardAlreadyExists, card.ID())
	}

	// Add bank card to the storage.
	s.cards[card.UserLogin()][card.ID()] = card

	return nil
}

// Get returns a bank card from the storage.
func (s *InMemory) Get(_ context.Context, userLogin, cardID string) (*bankcard.BankCard, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if the user login entry exists in the storage.
	cards, ok := s.cards[userLogin]
	if !ok {
		return nil, fmt.Errorf("%w for user login %s: %s", ErrCardNotFound, userLogin, cardID)
	}

	// Check if the bank card entry exists in the storage.
	if card, ok := cards[cardID]; ok {
		return card, nil
	}

	return nil, fmt.Errorf("%w for user login %s: %s", ErrCardNotFound, userLogin, cardID)
}

// List returns a list of bank cards from the storage.
func (s *InMemory) List(_ context.Context, userLogin string) ([]*bankcard.BankCard, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cardEntries := make([]*bankcard.BankCard, 0)

	// Check if the user login entry exists in the storage.
	cards, ok := s.cards[userLogin]
	if !ok {
		return cardEntries, nil
	}

	for _, card := range cards {
		cardEntries = append(cardEntries, card)
	}

	return cardEntries, nil
}

// Update updates a bank card in the storage.
func (s *InMemory) Update(_ context.Context, card *bankcard.BankCard) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the user login entry exists in the storage.
	cards, ok := s.cards[card.UserLogin()]
	if !ok {
		return fmt.Errorf("%w for user login %s: %s", ErrCardNotFound, card.UserLogin(), card.ID())
	}

	if _, ok := cards[card.ID()]; !ok {
		return fmt.Errorf("%w for user login %s: %s", ErrCardNotFound, card.UserLogin(), card.ID())
	}

	// Update bank card in the storage.
	s.cards[card.UserLogin()][card.ID()] = card

	return nil
}

// Delete deletes a bank card from the storage.
func (s *InMemory) Delete(_ context.Context, userLogin, cardID string) error {
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
