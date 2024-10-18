package filesvc

import (
	"context"

	"github.com/andymarkow/gophkeeper/internal/domain/vault/fileobj"
)

type Service interface {
	CreateFile(ctx context.Context, userID, objID string, req CreateFileRequest) (*fileobj.File, error)
}
