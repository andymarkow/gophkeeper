// Package credrepo provides credentials storage implementation.
package credrepo

import (
	"context"
	"fmt"
	"sync"

	"github.com/andymarkow/gophkeeper/internal/domain/vault/credential"
)

// InMemory represents in-memory credentials storage.
type InMemory struct {
	// UserLogin -> Login -> Credential.
	creds map[string]map[string]*credential.Credential

	mu sync.RWMutex
}

// NewInMemory creates new in-memory credentials storage.
func NewInMemory() *InMemory {
	return &InMemory{
		creds: make(map[string]map[string]*credential.Credential),
	}
}

// Add adds a new credential to the storage.
func (s *InMemory) Add(_ context.Context, cred *credential.Credential) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the user login entry exists in the storage.
	creds, ok := s.creds[cred.UserLogin()]
	if !ok {
		// UserLogin does not exist in the storage. Add user login and credential to the storage.
		s.creds[cred.UserLogin()][cred.ID()] = cred

		return nil
	}

	if _, ok := creds[cred.ID()]; ok {
		// Credential already exists in the storage.
		return fmt.Errorf("%w: %s", ErrCredAlreadyExists, cred.ID())
	}

	// Add credential to the storage.
	s.creds[cred.UserLogin()][cred.ID()] = cred

	return nil
}

// Get returns a credential from the storage.
func (s *InMemory) Get(_ context.Context, userLogin, credID string) (*credential.Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if the user login entry exists in the storage.
	creds, ok := s.creds[userLogin]
	if !ok {
		return nil, fmt.Errorf("%w for user login %s: %s", ErrCredNotFound, userLogin, credID)
	}

	// Check if the credential entry exists in the storage.
	if cred, ok := creds[credID]; ok {
		return cred, nil
	}

	return nil, fmt.Errorf("%w for user login %s: %s", ErrCredNotFound, userLogin, credID)
}

func (s *InMemory) List(_ context.Context, userLogin string) ([]*credential.Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	credEntries := make([]*credential.Credential, 0)

	creds, ok := s.creds[userLogin]
	if !ok {
		return credEntries, nil
	}

	for _, cred := range creds {
		credEntries = append(credEntries, cred)
	}

	return credEntries, nil
}

// Update updates a credential in the storage.
func (s *InMemory) Update(_ context.Context, cred *credential.Credential) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the user login entry exists in the storage.
	creds, ok := s.creds[cred.UserLogin()]
	if !ok {
		return fmt.Errorf("%w for user login %s: %s", ErrCredNotFound, cred.UserLogin(), cred.ID())
	}

	if _, ok := creds[cred.ID()]; !ok {
		return fmt.Errorf("%w for user login %s: %s", ErrCredNotFound, cred.UserLogin(), cred.ID())
	}

	// Update credential in the storage.
	s.creds[cred.UserLogin()][cred.ID()] = cred

	return nil
}

// Delete removes a credential from the storage.
func (s *InMemory) Delete(_ context.Context, userLogin string, credID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the user login entry exists in the storage.
	creds, ok := s.creds[userLogin]
	if !ok {
		return fmt.Errorf("%w for user login %s: %s", ErrCredNotFound, userLogin, credID)
	}

	if _, ok := creds[credID]; !ok {
		return fmt.Errorf("%w for user login %s: %s", ErrCredNotFound, userLogin, credID)
	}

	// Delete credential from the storage.
	delete(s.creds[userLogin], credID)

	return nil
}
