// Package cardrepo provides bank cards storage implementation.
package cardrepo

import (
	"context"
	"io"

	"github.com/andymarkow/gophkeeper/internal/domain/vault/bankcard"
)

type Storage interface {
	io.Closer

	AddSecret(ctx context.Context, secret *bankcard.Secret) (*bankcard.Secret, error)
	GetSecret(ctx context.Context, userID string, secretName string) (*bankcard.Secret, error)
	ListSecrets(ctx context.Context, userID string) ([]*bankcard.Secret, error)
	UpdateSecret(ctx context.Context, secret *bankcard.Secret) (*bankcard.Secret, error)
	DeleteSecret(ctx context.Context, userID string, secretName string) error
}
