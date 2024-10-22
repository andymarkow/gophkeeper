package credrepo

import (
	"context"
	"fmt"
	"sync"

	"github.com/andymarkow/gophkeeper/internal/domain/vault/credential"
)

var _ Storage = (*InMemory)(nil)

// InMemory represents in-memory credentials storage.
type InMemory struct {
	// UserID -> SecretName -> Secret.
	secrets map[string]map[string]credential.Secret

	mu sync.RWMutex
}

// NewInMemory creates new in-memory credentials storage.
func NewInMemory() *InMemory {
	return &InMemory{
		secrets: make(map[string]map[string]credential.Secret),
	}
}

// AddSecret adds a new credential secret entry to the storage.
func (s *InMemory) AddSecret(_ context.Context, secret *credential.Secret) (*credential.Secret, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the user id entry exists in the storage.
	secrets, ok := s.secrets[secret.UserID()]
	if !ok {
		// Check if secrets entry is nil.
		if secrets == nil {
			// Initialize secrets entry.
			s.secrets[secret.UserID()] = make(map[string]credential.Secret)
		}

		// UserID does not exist in the storage. Add user login and credential to the storage.
		s.secrets[secret.UserID()][secret.Name()] = *secret

		secr := s.secrets[secret.UserID()][secret.Name()]

		return &secr, nil
	}

	if _, ok := secrets[secret.Name()]; ok {
		// Credential secret entry already exists in the storage.
		return nil, fmt.Errorf("%w: %s", ErrSecretAlreadyExists, secret.Name())
	}

	// Add credential secret entry to the storage.
	s.secrets[secret.UserID()][secret.Name()] = *secret

	secr := s.secrets[secret.UserID()][secret.Name()]

	return &secr, nil
}

// GetSecret returns a credential from the storage.
func (s *InMemory) GetSecret(_ context.Context, userID, secretName string) (*credential.Secret, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if the user id entry exists in the storage.
	secrets, ok := s.secrets[userID]
	if !ok {
		return nil, fmt.Errorf("%w for user id %s: %s", ErrSecretNotFound, userID, secretName)
	}

	// Check if the credential entry exists in the storage.
	if secret, ok := secrets[secretName]; ok {
		secr := secret

		return &secr, nil
	}

	return nil, fmt.Errorf("%w for user id %s: %s", ErrSecretNotFound, userID, secretName)
}

// ListSecrets returns a list of credential secret entries from the storage.
func (s *InMemory) ListSecrets(_ context.Context, userID string) ([]*credential.Secret, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if the user id entry exists in the storage.
	secrets, ok := s.secrets[userID]
	if !ok {
		return []*credential.Secret{}, nil
	}

	secretsList := make([]*credential.Secret, 0, len(secrets))

	for _, secret := range secrets {
		secr := secret
		secretsList = append(secretsList, &secr)
	}

	return secretsList, nil
}

// UpdateSecret updates a credential secret entry in the storage.
func (s *InMemory) UpdateSecret(_ context.Context, secret *credential.Secret) (*credential.Secret, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the user id entry exists in the storage.
	secrets, ok := s.secrets[secret.UserID()]
	if !ok {
		return nil, fmt.Errorf("%w for user id %s: %s", ErrSecretNotFound, secret.UserID(), secret.Name())
	}

	// Check if the credential secret entry exists in the storage.
	if _, ok := secrets[secret.Name()]; !ok {
		return nil, fmt.Errorf("%w for user id %s: %s", ErrSecretNotFound, secret.UserID(), secret.Name())
	}

	// Update credential secret entry in the storage.
	s.secrets[secret.UserID()][secret.Name()] = *secret

	secr := s.secrets[secret.UserID()][secret.Name()]

	return &secr, nil
}

// DeleteSecret deletes a credential secret entry from the storage.
func (s *InMemory) DeleteSecret(_ context.Context, userID string, secretName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the user id entry exists in the storage.
	secrets, ok := s.secrets[userID]
	if !ok {
		return fmt.Errorf("%w for user id %s: %s", ErrSecretNotFound, userID, secretName)
	}

	if _, ok := secrets[secretName]; !ok {
		return fmt.Errorf("%w for user id %s: %s", ErrSecretNotFound, userID, secretName)
	}

	// Delete credential secret entry from the storage.
	delete(s.secrets[userID], secretName)

	return nil
}
