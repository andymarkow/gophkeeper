package textsvc

import (
	"context"
	"io"

	"github.com/andymarkow/gophkeeper/internal/domain/vault/text"
)

// Service represents the text data service.
type Service interface {
	// io.Closer

	CreateSecret(ctx context.Context, userID, secretName string, metadata map[string]string) (*text.Secret, error)
	ListSecrets(ctx context.Context, userID string) ([]*text.Secret, error)
	GetSecret(ctx context.Context, userID, secretName string) (*text.Secret, error)
	UpdateSecret(ctx context.Context, userID, secretName string, metadata map[string]string) (*text.Secret, error)
	DeleteSecret(ctx context.Context, userID, secretName string) error
	UploadSecret(ctx context.Context, userID, secretName string, data io.Reader) (*text.Secret, error)
	DownloadSecret(ctx context.Context, userID, secretName string) (*text.Secret, io.ReadCloser, error)
}
