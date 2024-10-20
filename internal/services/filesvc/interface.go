package filesvc

import (
	"context"
	"io"

	"github.com/andymarkow/gophkeeper/internal/domain/vault/fileobj"
)

type Service interface {
	CreateFile(ctx context.Context, userID, objID string, metadata map[string]string) (*fileobj.File, error)
	UpdateFile(ctx context.Context, userID, objID, fileName string, metadata map[string]string) (*fileobj.File, error)
	ListFiles(ctx context.Context, userID string) ([]*fileobj.File, error)
	GetFile(ctx context.Context, userID, objID string) (*fileobj.File, error)
	UploadFile(ctx context.Context, userID, objID string, req UploadFileRequest) (*fileobj.File, error)
	DownloadFile(ctx context.Context, userID, objID string) (*fileobj.File, io.ReadCloser, error)
	DeleteFile(ctx context.Context, userID, objID string) error
}
