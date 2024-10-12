package credrepo

import (
	"context"

	"github.com/andymarkow/gophkeeper/internal/domain/vault/credential"
)

var _ Storage = (*InMemory)(nil)

// Storage represents credentials storage interface.
type Storage interface {
	Add(ctx context.Context, cred *credential.Credential) error
	Get(ctx context.Context, userLogin, credID string) (*credential.Credential, error)
	List(ctx context.Context, userLogin string) ([]*credential.Credential, error)
	Update(ctx context.Context, cred *credential.Credential) error
	Delete(ctx context.Context, userLogin string, credID string) error
}
