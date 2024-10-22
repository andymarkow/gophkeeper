// Package cardrepo provides bank cards storage implementation.
package cardrepo

import (
	"context"

	"github.com/andymarkow/gophkeeper/internal/domain/vault/bankcard"
)

type Storage interface {
	AddSecret(ctx context.Context, secret *bankcard.Secret) (*bankcard.Secret, error)
	GetSecret(ctx context.Context, userID string, secretID string) (*bankcard.Secret, error)
	ListSecrets(ctx context.Context, userID string) ([]*bankcard.Secret, error)
	UpdateSecret(ctx context.Context, secret *bankcard.Secret) (*bankcard.Secret, error)
	DeleteSecret(ctx context.Context, userID string, secretID string) error
}
