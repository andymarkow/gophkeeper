// Package userinmem provides in-memory storage implementation for user model.
package userinmem

import (
	"context"
	"fmt"
	"sync"

	"github.com/andymarkow/gophkeeper/internal/domain/user"
	"github.com/andymarkow/gophkeeper/internal/storage/userrepo"
)

var _ userrepo.Storage = (*InMemory)(nil)

// InMemory represents in-memory user storage.
type InMemory struct {
	users map[string]*user.User

	mu sync.RWMutex
}

// NewInMemory creates new in-memory user storage.
func NewInMemory() *InMemory {
	return &InMemory{
		users: make(map[string]*user.User),
	}
}

// Close closes the storage.
func (s *InMemory) Close() error {
	return nil
}

// AddUser adds a new user to the storage.
func (s *InMemory) AddUser(_ context.Context, usr *user.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.users[usr.Login()]; ok {
		return fmt.Errorf("%w: %s", userrepo.ErrUsrAlreadyExists, usr.Login())
	}

	s.users[usr.Login()] = usr

	return nil
}

// GetUser returns a user from the storage.
func (s *InMemory) GetUser(_ context.Context, login string) (*user.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if usr, ok := s.users[login]; ok {
		return usr, nil
	}

	return nil, fmt.Errorf("%w: %s", userrepo.ErrUsrNotFound, login)
}
