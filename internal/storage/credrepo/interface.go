// Package credrepo provides credentials storage implementation.
package credrepo

import (
	"context"

	"github.com/andymarkow/gophkeeper/internal/domain/vault/credential"
)

var _ Storage = (*InMemory)(nil)

// Storage represents credentials storage interface.
type Storage interface {
	AddCredential(ctx context.Context, cred *credential.Credential) error
	GetCredential(ctx context.Context, userLogin, credID string) (*credential.Credential, error)
	ListCredentials(ctx context.Context, userLogin string) ([]string, error)
	UpdateCredential(ctx context.Context, cred *credential.Credential) error
	DeleteCredential(ctx context.Context, userLogin string, credID string) error
}
