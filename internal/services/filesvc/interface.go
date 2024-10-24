package filesvc

import (
	"context"
	"io"

	"github.com/andymarkow/gophkeeper/internal/domain/vault/file"
)

type Service interface {
	io.Closer

	CreateSecret(ctx context.Context, userID, secretName string, metadata map[string]string) (*file.Secret, error)
	UpdateSecret(ctx context.Context, userID, secretName, fileName string, metadata map[string]string) (*file.Secret, error)
	ListSecrets(ctx context.Context, userID string) ([]*file.Secret, error)
	GetSecret(ctx context.Context, userID, secretName string) (*file.Secret, error)
	UploadSecret(ctx context.Context, userID, secretName string, req UploadSecretRequest) (*file.Secret, error)
	DownloadSecret(ctx context.Context, userID, secretName string) (*file.Secret, io.ReadCloser, error)
	DeleteSecret(ctx context.Context, userID, secretName string) error
}
