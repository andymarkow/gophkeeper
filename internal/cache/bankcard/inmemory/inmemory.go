// Package inmemory provides in-memory bank cards cache storage.
package inmemory

import (
	"context"
	"fmt"
	"sync"

	cache "github.com/andymarkow/gophkeeper/internal/cache/bankcard"
)

// Storage represents in-memory cache storage.
type Storage struct {
	// UserID -> SecretName -> Item.
	items map[string]map[string]cache.Item

	mu sync.RWMutex
}

// NewStorage creates new in-memory cache storage.
func NewStorage() *Storage {
	return &Storage{
		items: make(map[string]map[string]cache.Item),
	}
}

// Close closes the storage.
func (s *Storage) Close() error {
	return nil
}

// CreateItem creates a new item in the cache storage.
func (s *Storage) CreateItem(_ context.Context, item *cache.Item) (*cache.Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	items, ok := s.items[item.Secret().UserID()]
	if !ok {
		if items == nil {
			s.items[item.Secret().UserID()] = make(map[string]cache.Item)
		}

		s.items[item.Secret().UserID()][item.Secret().Name()] = *item

		it := s.items[item.Secret().UserID()][item.Secret().Name()]

		return &it, nil
	}

	if _, ok := items[item.Secret().Name()]; ok {
		return nil, fmt.Errorf("%w: %s", cache.ErrItemAlreadyExists, item.Secret().Name())
	}

	s.items[item.Secret().UserID()][item.Secret().Name()] = *item

	it := s.items[item.Secret().UserID()][item.Secret().Name()]

	return &it, nil
}

// GetItem returns an item from the cache storage.
func (s *Storage) GetItem(_ context.Context, userID, secretName string) (*cache.Item, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items, ok := s.items[userID]
	if !ok {
		return nil, fmt.Errorf("%w for user id %s: %s", cache.ErrItemNotFound, userID, secretName)
	}

	if item, ok := items[secretName]; ok {
		it := item

		return &it, nil
	}

	return nil, fmt.Errorf("%w for user id %s: %s", cache.ErrItemNotFound, userID, secretName)
}

// ListItems returns a list of items from the cache storage.
func (s *Storage) ListItems(_ context.Context, userID string) ([]*cache.Item, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items, ok := s.items[userID]
	if !ok {
		return []*cache.Item{}, nil
	}

	itemsList := make([]*cache.Item, 0, len(items))

	for _, card := range items {
		crd := card
		itemsList = append(itemsList, &crd)
	}

	return itemsList, nil
}

// UpdateItem updates an item in the cache storage.
func (s *Storage) UpdateItem(_ context.Context, item *cache.Item) (*cache.Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	items, ok := s.items[item.Secret().UserID()]
	if !ok {
		return nil, fmt.Errorf("%w for user id %s: %s", cache.ErrItemNotFound, item.Secret().UserID(), item.Secret().Name())
	}

	if _, ok := items[item.Secret().Name()]; !ok {
		return nil, fmt.Errorf("%w for user id %s: %s", cache.ErrItemNotFound, item.Secret().UserID(), item.Secret().Name())
	}

	s.items[item.Secret().UserID()][item.Secret().Name()] = *item

	it := s.items[item.Secret().UserID()][item.Secret().Name()]

	return &it, nil
}

// DeleteItem deletes an item from the cache storage.
func (s *Storage) DeleteItem(_ context.Context, userID, secretName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	items, ok := s.items[userID]
	if !ok {
		return fmt.Errorf("%w for user id %s: %s", cache.ErrItemNotFound, userID, secretName)
	}

	if _, ok := items[secretName]; !ok {
		return fmt.Errorf("%w for user id %s: %s", cache.ErrItemNotFound, userID, secretName)
	}

	delete(s.items[userID], secretName)

	return nil
}

func (s *Storage) SetItem(_ context.Context, item *cache.Item) (*cache.Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	items, ok := s.items[item.Secret().UserID()]
	if !ok {
		if items == nil {
			s.items[item.Secret().UserID()] = make(map[string]cache.Item)
		}

		s.items[item.Secret().UserID()][item.Secret().Name()] = *item

		it := s.items[item.Secret().UserID()][item.Secret().Name()]

		return &it, nil
	}

	s.items[item.Secret().UserID()][item.Secret().Name()] = *item

	it := s.items[item.Secret().UserID()][item.Secret().Name()]

	return &it, nil
}
