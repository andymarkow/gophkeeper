package objrepo

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinioClient represents a minio client.
type MinioClient struct {
	client *minio.Client
	bucket string
}

// MinioClientOpts represents minio client options.
type MinioClientOpts struct {
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
}

// NewMinioClient creates a new minio client.
func NewMinioClient(endpoint, bucket string, opts *MinioClientOpts) (*MinioClient, error) {
	if opts == nil {
		opts = defaultOpts()
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(opts.AccessKeyID, opts.SecretAccessKey, ""),
		Secure: opts.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio.New: %w", err)
	}

	return &MinioClient{client: client, bucket: bucket}, nil
}

// defaultOpts returns the default options.
func defaultOpts() *MinioClientOpts {
	return &MinioClientOpts{
		AccessKeyID:     "",
		SecretAccessKey: "",
		UseSSL:          false,
	}
}

// InitBucket creates the bucket if it doesn't exist.
func (c *MinioClient) InitBucket(ctx context.Context) error {
	found, err := c.client.BucketExists(ctx, c.bucket)
	if err != nil {
		return fmt.Errorf("client.BucketExists: %w", err)
	}

	if found {
		return nil
	}

	err = c.client.MakeBucket(ctx, c.bucket, minio.MakeBucketOptions{})
	if err != nil {
		return fmt.Errorf("client.MakeBucket: %w", err)
	}

	return nil
}

// GetObject gets an object from the bucket.
func (c *MinioClient) GetObject(ctx context.Context, objName string) (*minio.Object, error) {
	info, err := c.client.GetObject(ctx, c.bucket, objName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("client.GetObject: %w", err)
	}

	return info, nil
}

// PutObject puts an object into the bucket.
func (c *MinioClient) PutObject(ctx context.Context, objName string, objSize int64, rd io.Reader) (*minio.UploadInfo, error) {
	info, err := c.client.PutObject(ctx, c.bucket, objName, rd, objSize, minio.PutObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("client.PutObject: %w", err)
	}

	return &info, nil
}

// RemoveObject removes an object from the bucket.
func (c *MinioClient) RemoveObject(ctx context.Context, objName string) error {
	err := c.client.RemoveObject(ctx, c.bucket, objName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("client.RemoveObject: %w", err)
	}

	return nil
}
