package textrepo

import (
	"context"
	"fmt"
	"sync"

	"github.com/andymarkow/gophkeeper/internal/domain/vault/text"
)

var _ Storage = (*InMemory)(nil)

// InMemory represents in-memory text data storage.
type InMemory struct {
	// UserID -> SecretName -> Secret.
	secrets map[string]map[string]text.Secret

	mu sync.RWMutex
}

// NewInMemory creates new in-memory text data storage.
func NewInMemory() *InMemory {
	return &InMemory{
		secrets: make(map[string]map[string]text.Secret),
	}
}

// AddSecret adds a new secret entry to the storage.
func (s *InMemory) AddSecret(_ context.Context, secret *text.Secret) (*text.Secret, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the user login entry exists in the storage.
	secrets, ok := s.secrets[secret.UserID()]
	if !ok {
		// Check if files entry is nil.
		if secrets == nil {
			// Initialize files entry.
			s.secrets[secret.UserID()] = make(map[string]text.Secret)
		}

		// UserID does not exist in the storage. Add user login and text data entry to the storage.
		s.secrets[secret.UserID()][secret.Name()] = *secret

		return secret, nil
	}

	if _, ok := secrets[secret.Name()]; ok {
		// Text data entry already exists in the storage.
		return nil, fmt.Errorf("%w: %s", ErrSecretEntryAlreadyExists, secret.Name())
	}

	// Add secret entry to the storage.
	s.secrets[secret.UserID()][secret.Name()] = *secret

	return secret, nil
}

// GetSecret returns a secret entry from the storage.
func (s *InMemory) GetSecret(_ context.Context, userID, secretName string) (*text.Secret, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if the user login entry exists in the storage.
	secrets, ok := s.secrets[userID]
	if !ok {
		return nil, fmt.Errorf("%w for user login %s: %s", ErrSecretEntryNotFound, userID, secretName)
	}

	// Check if the secret entry exists in the storage.
	if secret, ok := secrets[secretName]; ok {
		sec := secret

		return &sec, nil
	}

	return nil, fmt.Errorf("%w for user login %s: %s", ErrSecretEntryNotFound, userID, secretName)
}

// ListSecrets returns a list of secret entries from the storage.
func (s *InMemory) ListSecrets(_ context.Context, userID string) ([]*text.Secret, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	secrets := s.secrets[userID]

	secretsList := make([]*text.Secret, 0, len(secrets))

	for _, secret := range secrets {
		sec := secret
		secretsList = append(secretsList, &sec)
	}

	return secretsList, nil
}

// UpdateSecret updates a secret entry in the storage.
func (s *InMemory) UpdateSecret(_ context.Context, secret *text.Secret) (*text.Secret, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the user login entry exists in the storage.
	secrets, ok := s.secrets[secret.UserID()]
	if !ok {
		return nil, fmt.Errorf("%w for user login %s: %s", ErrSecretEntryNotFound, secret.UserID(), secret.Name())
	}

	// Check if the text data entry exists in the storage.
	if _, ok := secrets[secret.Name()]; !ok {
		return nil, fmt.Errorf("%w for user login %s: %s", ErrSecretEntryNotFound, secret.UserID(), secret.Name())
	}

	// Update secret entry in the storage.
	s.secrets[secret.UserID()][secret.Name()] = *secret

	sec := s.secrets[secret.UserID()][secret.Name()]

	return &sec, nil
}

// DeleteSecret deletes a secret entry from the storage.
func (s *InMemory) DeleteSecret(_ context.Context, userID, secretName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the user login entry exists in the storage.
	secrets, ok := s.secrets[userID]
	if !ok {
		return fmt.Errorf("%w for user login %s: %s", ErrSecretEntryNotFound, userID, secretName)
	}

	// Check if the secret entry exists in the storage.
	if _, ok := secrets[secretName]; !ok {
		return fmt.Errorf("%w for user login %s: %s", ErrSecretEntryNotFound, userID, secretName)
	}

	// Delete secret entry from the storage.
	delete(s.secrets[userID], secretName)

	return nil
}
