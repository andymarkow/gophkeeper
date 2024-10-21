package filesvc

import (
	"context"
	"io"

	"github.com/andymarkow/gophkeeper/internal/domain/vault/file"
)

type Service interface {
	CreateFile(ctx context.Context, userID, secretName string, metadata map[string]string) (*file.Secret, error)
	UpdateFile(ctx context.Context, userID, secretName, fileName string, metadata map[string]string) (*file.Secret, error)
	ListFiles(ctx context.Context, userID string) ([]*file.Secret, error)
	GetFile(ctx context.Context, userID, secretName string) (*file.Secret, error)
	UploadFile(ctx context.Context, userID, secretName string, req UploadFileRequest) (*file.Secret, error)
	DownloadFile(ctx context.Context, userID, secretName string) (*file.Secret, io.ReadCloser, error)
	DeleteFile(ctx context.Context, userID, secretName string) error
}
