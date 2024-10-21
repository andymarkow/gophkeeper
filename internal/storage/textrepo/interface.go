// Package textrepo provides file storage implementation.
package textrepo

import (
	"context"

	"github.com/andymarkow/gophkeeper/internal/domain/vault/text"
)

// Storage represents text data storage interface.
type Storage interface {
	AddSecret(ctx context.Context, secret *text.Secret) (*text.Secret, error)
	GetSecret(ctx context.Context, userID, secretName string) (*text.Secret, error)
	ListSecrets(ctx context.Context, userID string) ([]*text.Secret, error)
	UpdateSecret(ctx context.Context, secret *text.Secret) (*text.Secret, error)
	DeleteSecret(ctx context.Context, userID string, secretName string) error
}
