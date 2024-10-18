// Package filerepo provides file storage implementation.
package filerepo

import (
	"context"

	"github.com/andymarkow/gophkeeper/internal/domain/vault/fileobj"
)

var _ Storage = (*InMemory)(nil)

// Storage represents files storage interface.
type Storage interface {
	AddFile(ctx context.Context, file *fileobj.File) (*fileobj.File, error)
	GetFile(ctx context.Context, userLogin, fileID string) (*fileobj.File, error)
	ListFiles(ctx context.Context, userLogin string) ([]*fileobj.File, error)
	UpdateFile(ctx context.Context, file *fileobj.File) error
	DeleteFile(ctx context.Context, userLogin string, fileID string) error
}
