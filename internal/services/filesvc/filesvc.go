// Package filesvc provides the file service.
package filesvc

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/andymarkow/gophkeeper/internal/cryptutils"
	"github.com/andymarkow/gophkeeper/internal/domain/vault/fileobj"
	"github.com/andymarkow/gophkeeper/internal/storage/filerepo"
	"github.com/andymarkow/gophkeeper/internal/storage/objrepo"
)

// Service represents the service.
type Service struct {
	log         *slog.Logger
	fileStorage filerepo.Storage
	objStorage  objrepo.Storage
	cryptoKey   string
}

// NewService creates a new service.
func NewService(filestore filerepo.Storage, objstore objrepo.Storage, opts ...Option) *Service {
	svc := &Service{
		log:         slog.New(&slog.JSONHandler{}),
		fileStorage: filestore,
		objStorage:  objstore,
	}

	for _, opt := range opts {
		opt(svc)
	}

	return svc
}

// Option is a functional option for the service.
type Option func(*Service)

// WithLogger sets the logger for the service.
func WithLogger(log *slog.Logger) Option {
	return func(s *Service) {
		s.log = log
	}
}

// WithCryptoKey sets the crypto key for the service.
func WithCryptoKey(key string) Option {
	return func(s *Service) {
		s.cryptoKey = key
	}
}

func (s *Service) CreateFile(ctx context.Context, file *fileobj.File, objName string, data io.Reader) (any, error) {
	// Create encrypted file data stream.
	stream, err := cryptutils.EncryptStream(s.cryptoKey, data)
	if err != nil {
		return nil, fmt.Errorf("cryptutils.EncryptStream: %w", err)
	}

	// Put the file data to the object storage.
	info, err := s.objStorage.PutObject(ctx, objName, -1, stream)
	if err != nil {
		return nil, fmt.Errorf("storage.PutObject: %w", err)
	}

	// Create new file object entry to store it in the file storage.
	repoFile, err := fileobj.NewFile(
		file.ID(), file.UserID(), file.Name(), file.Metadata(),
		time.Now(), time.Now(), info.Location(), info.CRC32C())
	if err != nil {
		return nil, fmt.Errorf("fileobj.NewFile: %w", err)
	}

	// Store the file object entry in the file storage.
	err = s.fileStorage.AddFile(ctx, repoFile)
	if err != nil {
		return nil, fmt.Errorf("storage.AddFile: %w", err)
	}

	return info, nil
}
