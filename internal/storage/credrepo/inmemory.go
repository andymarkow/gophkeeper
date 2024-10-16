package credrepo

import (
	"context"
	"fmt"
	"sync"

	"github.com/andymarkow/gophkeeper/internal/domain/vault/credential"
)

// InMemory represents in-memory credentials storage.
type InMemory struct {
	// UserID -> Login -> Credential.
	creds map[string]map[string]credential.Credential

	mu sync.RWMutex
}

// NewInMemory creates new in-memory credentials storage.
func NewInMemory() *InMemory {
	return &InMemory{
		creds: make(map[string]map[string]credential.Credential),
	}
}

// AddCredential adds a new credential to the storage.
func (s *InMemory) AddCredential(_ context.Context, cred *credential.Credential) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the user login entry exists in the storage.
	creds, ok := s.creds[cred.UserID()]
	if !ok {
		// Check if creds entry is nil.
		if creds == nil {
			// Initialize creds entry.
			s.creds[cred.UserID()] = make(map[string]credential.Credential)
		}

		// UserID does not exist in the storage. Add user login and credential to the storage.
		s.creds[cred.UserID()][cred.ID()] = *cred

		return nil
	}

	if _, ok := creds[cred.ID()]; ok {
		// Credential already exists in the storage.
		return fmt.Errorf("%w: %s", ErrCredAlreadyExists, cred.ID())
	}

	// Add credential to the storage.
	s.creds[cred.UserID()][cred.ID()] = *cred

	return nil
}

// GetCredential returns a credential from the storage.
func (s *InMemory) GetCredential(_ context.Context, userLogin, credID string) (*credential.Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if the user login entry exists in the storage.
	creds, ok := s.creds[userLogin]
	if !ok {
		return nil, fmt.Errorf("%w for user login %s: %s", ErrCredNotFound, userLogin, credID)
	}

	// Check if the credential entry exists in the storage.
	if cred, ok := creds[credID]; ok {
		crd := cred

		return &crd, nil
	}

	return nil, fmt.Errorf("%w for user login %s: %s", ErrCredNotFound, userLogin, credID)
}

func (s *InMemory) ListCredentials(_ context.Context, userLogin string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if the user login entry exists in the storage.
	creds, ok := s.creds[userLogin]
	if !ok {
		return []string{}, nil
	}

	credIDs := make([]string, 0, len(creds))

	for credID := range creds {
		credIDs = append(credIDs, credID)
	}

	return credIDs, nil
}

// UpdateCredential updates a credential in the storage.
func (s *InMemory) UpdateCredential(_ context.Context, cred *credential.Credential) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the user login entry exists in the storage.
	creds, ok := s.creds[cred.UserID()]
	if !ok {
		return fmt.Errorf("%w for user login %s: %s", ErrCredNotFound, cred.UserID(), cred.ID())
	}

	// Check if the credential entry exists in the storage.
	if _, ok := creds[cred.ID()]; !ok {
		return fmt.Errorf("%w for user login %s: %s", ErrCredNotFound, cred.UserID(), cred.ID())
	}

	// Update credential in the storage.
	s.creds[cred.UserID()][cred.ID()] = *cred

	return nil
}

// DeleteCredential removes a credential from the storage.
func (s *InMemory) DeleteCredential(_ context.Context, userLogin string, credID string) error {
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
