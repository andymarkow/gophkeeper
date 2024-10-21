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
	secrets map[string]map[string]bankcard.Secret

	mu sync.RWMutex
}

// NewInMemory creates new in-memory bank cards storage.
func NewInMemory() *InMemory {
	return &InMemory{
		secrets: make(map[string]map[string]bankcard.Secret),
	}
}

// AddSecret adds a new bank card secret entry to the storage.
func (s *InMemory) AddSecret(_ context.Context, secret *bankcard.Secret) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the user login entry exists in the storage.
	secrets, ok := s.secrets[secret.UserID()]
	if !ok {
		// Check if cards entry is nil.
		if secrets == nil {
			// Initialize cards entry.
			s.secrets[secret.UserID()] = make(map[string]bankcard.Secret)
		}

		// UserID does not exist in the storage. Add user login and bank card to the storage.
		s.secrets[secret.UserID()][secret.Name()] = *secret

		return nil
	}

	if _, ok := secrets[secret.Name()]; ok {
		// Bank card already exists in the storage.
		return fmt.Errorf("%w: %s", ErrSecretAlreadyExists, secret.Name())
	}

	// Add bank card to the storage.
	s.secrets[secret.UserID()][secret.Name()] = *secret

	return nil
}

// GetSecret returns a bank card secret entry from the storage.
func (s *InMemory) GetSecret(_ context.Context, userID, secretName string) (*bankcard.Secret, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if the user login entry exists in the storage.
	secrets, ok := s.secrets[userID]
	if !ok {
		return nil, fmt.Errorf("%w for user login %s: %s", ErrSecretNotFound, userID, secretName)
	}

	// Check if the bank card secret entry exists in the storage.
	if card, ok := secrets[secretName]; ok {
		crd := card

		return &crd, nil
	}

	return nil, fmt.Errorf("%w for user login %s: %s", ErrSecretNotFound, userID, secretName)
}

// ListSecrets returns a list of bank card secret entries from the storage.
func (s *InMemory) ListSecrets(_ context.Context, userID string) ([]*bankcard.Secret, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if the user login entry exists in the storage.
	secrets, ok := s.secrets[userID]
	if !ok {
		return []*bankcard.Secret{}, nil
	}

	secretsList := make([]*bankcard.Secret, 0, len(secrets))

	for _, card := range secrets {
		crd := card
		secretsList = append(secretsList, &crd)
	}

	return secretsList, nil
}

// UpdateSecret updates a bank card secret entry in the storage.
func (s *InMemory) UpdateSecret(_ context.Context, secret *bankcard.Secret) (*bankcard.Secret, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the user login entry exists in the storage.
	secrets, ok := s.secrets[secret.UserID()]
	if !ok {
		return nil, fmt.Errorf("%w for user login %s: %s", ErrSecretNotFound, secret.UserID(), secret.Name())
	}

	if _, ok := secrets[secret.Name()]; !ok {
		return nil, fmt.Errorf("%w for user login %s: %s", ErrSecretNotFound, secret.UserID(), secret.Name())
	}

	// Update bank card secret entry in the storage.
	s.secrets[secret.UserID()][secret.Name()] = *secret

	sec := s.secrets[secret.UserID()][secret.Name()]

	return &sec, nil
}

// DeleteSecret deletes a bank card secret entry from the storage.
func (s *InMemory) DeleteSecret(_ context.Context, userID, secretName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the user login entry exists in the storage.
	secrets, ok := s.secrets[userID]
	if !ok {
		return fmt.Errorf("%w for user login %s: %s", ErrSecretNotFound, userID, secretName)
	}

	if _, ok := secrets[secretName]; !ok {
		return fmt.Errorf("%w for user login %s: %s", ErrSecretNotFound, userID, secretName)
	}

	// Delete bank card secret entry from the storage.
	delete(s.secrets[userID], secretName)

	return nil
}
