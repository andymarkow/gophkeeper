// Package objrepo provides the minio client.
package objrepo

import (
	"context"
	"io"
)

type Storage interface {
	GetObjectInfo(ctx context.Context, objName string) (*ObjectInfo, error)
	GetObject(ctx context.Context, objName string) (io.ReadSeekCloser, error)
	PutObject(ctx context.Context, objName string, objSize int64, rd io.Reader) (*ObjectInfo, error)
	RemoveObject(ctx context.Context, objName string) error
}
