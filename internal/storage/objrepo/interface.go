// Package objrepo provides the minio client.
package objrepo

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
)

type Storage interface {
	GetObject(ctx context.Context, objName string) (*minio.Object, error)
	PutObject(ctx context.Context, objName string, objSize int64, rd io.Reader) (*minio.UploadInfo, error)
	RemoveObject(ctx context.Context, objName string) error
}
