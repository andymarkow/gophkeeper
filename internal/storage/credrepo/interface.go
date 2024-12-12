// Package credrepo provides credentials storage implementation.
package credrepo

import (
	"context"
	"io"

	"github.com/andymarkow/gophkeeper/internal/domain/vault/credential"
)

// Storage represents credentials storage interface.
type Storage interface {
	io.Closer

	AddSecret(ctx context.Context, secret *credential.Secret) (*credential.Secret, error)
	GetSecret(ctx context.Context, userID, secretName string) (*credential.Secret, error)
	ListSecrets(ctx context.Context, userID string) ([]*credential.Secret, error)
	UpdateSecret(ctx context.Context, secret *credential.Secret) (*credential.Secret, error)
	DeleteSecret(ctx context.Context, userID string, secretName string) error
}
