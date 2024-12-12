// Package fileinmem provides in-memory file storage implementation.
package fileinmem

import (
	"context"
	"fmt"
	"sync"

	"github.com/andymarkow/gophkeeper/internal/domain/vault/file"
	"github.com/andymarkow/gophkeeper/internal/storage/filerepo"
)

var _ filerepo.Storage = (*InMemory)(nil)

// InMemory represents in-memory files storage.
type InMemory struct {
	secrets map[string]map[string]file.Secret

	mu sync.RWMutex
}

// NewInMemory creates new in-memory files storage.
func NewInMemory() *InMemory {
	return &InMemory{
		secrets: make(map[string]map[string]file.Secret),
	}
}

// Close closes the storage.
func (s *InMemory) Close() error {
	return nil
}

// AddSecret adds a new secret entry to the storage.
func (s *InMemory) AddSecret(_ context.Context, secret *file.Secret) (*file.Secret, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the user login entry exists in the storage.
	secrets, ok := s.secrets[secret.UserID()]
	if !ok {
		// Check if secrets entry is nil.
		if secrets == nil {
			// Initialize files entry.
			s.secrets[secret.UserID()] = make(map[string]file.Secret)
		}

		// UserID does not exist in the storage. Add user login and file to the storage.
		s.secrets[secret.UserID()][secret.Name()] = *secret

		secr := s.secrets[secret.UserID()][secret.Name()]

		return &secr, nil
	}

	if _, ok := secrets[secret.Name()]; ok {
		// Secret already exists in the storage.
		return nil, fmt.Errorf("%w: %s", filerepo.ErrSecretAlreadyExists, secret.Name())
	}

	// Add secret to the storage.
	s.secrets[secret.UserID()][secret.Name()] = *secret

	secr := s.secrets[secret.UserID()][secret.Name()]

	return &secr, nil
}

// GetSecret returns a secret entry from the storage.
func (s *InMemory) GetSecret(_ context.Context, userID, secretName string) (*file.Secret, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if the user id entry exists in the storage.
	secrets, ok := s.secrets[userID]
	if !ok {
		return nil, fmt.Errorf("%w for user id %s and secret id %s", filerepo.ErrSecretNotFound, userID, secretName)
	}

	// Check if the secret entry exists in the storage.
	if secret, ok := secrets[secretName]; ok {
		s := secret

		return &s, nil
	}

	return nil, fmt.Errorf("%w for user id %s: %s", filerepo.ErrSecretNotFound, userID, secretName)
}

// ListSecrets returns a list of secret entries from the storage.
func (s *InMemory) ListSecrets(_ context.Context, userID string) ([]*file.Secret, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	secrets := s.secrets[userID]

	sl := make([]*file.Secret, 0, len(secrets))

	for _, secret := range secrets {
		s := secret
		sl = append(sl, &s)
	}

	return sl, nil
}

// UpdateSecret updates a secret entry in the storage.
func (s *InMemory) UpdateSecret(_ context.Context, secret *file.Secret) (*file.Secret, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the user id entry exists in the storage.
	secrets, ok := s.secrets[secret.UserID()]
	if !ok {
		return nil, fmt.Errorf("%w for user id %s: %s", filerepo.ErrSecretNotFound, secret.UserID(), secret.Name())
	}

	// Check if the secret entry exists in the storage.
	if _, ok := secrets[secret.Name()]; !ok {
		return nil, fmt.Errorf("%w for user id %s: %s", filerepo.ErrSecretNotFound, secret.UserID(), secret.Name())
	}

	// Update secret entry in the storage.
	s.secrets[secret.UserID()][secret.Name()] = *secret

	secrt := s.secrets[secret.UserID()][secret.Name()]

	return &secrt, nil
}

// DeleteSecret deletes a secret entry from the storage.
func (s *InMemory) DeleteSecret(_ context.Context, userID, secretName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the user id entry exists in the storage.
	secrets, ok := s.secrets[userID]
	if !ok {
		return fmt.Errorf("%w for user id %s: %s", filerepo.ErrSecretNotFound, userID, secretName)
	}

	// Check if the secret entry exists in the storage.
	if _, ok := secrets[secretName]; !ok {
		return fmt.Errorf("%w for user id %s: %s", filerepo.ErrSecretNotFound, userID, secretName)
	}

	// Delete secret entry from the storage.
	delete(s.secrets[userID], secretName)

	return nil
}
