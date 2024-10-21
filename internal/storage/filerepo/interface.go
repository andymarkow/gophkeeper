// Package filerepo provides file storage implementation.
package filerepo

import (
	"context"

	"github.com/andymarkow/gophkeeper/internal/domain/vault/file"
)

// Storage represents files storage interface.
type Storage interface {
	AddSecret(ctx context.Context, secret *file.Secret) (*file.Secret, error)
	GetSecret(ctx context.Context, userID, secretName string) (*file.Secret, error)
	ListSecrets(ctx context.Context, userID string) ([]*file.Secret, error)
	UpdateSecret(ctx context.Context, secret *file.Secret) (*file.Secret, error)
	DeleteSecret(ctx context.Context, userID string, secretName string) error
}
