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

var _ Service = (*FileService)(nil)

// FileService represents the service.
type FileService struct {
	log         *slog.Logger
	fileStorage filerepo.Storage
	objStorage  objrepo.Storage
	objBasePath string
	cryptoKey   string
}

// NewFileService creates a new service.
func NewFileService(filestore filerepo.Storage, objstore objrepo.Storage, opts ...Option) *FileService {
	svc := &FileService{
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
type Option func(*FileService)

// WithLogger sets the logger for the service.
func WithLogger(log *slog.Logger) Option {
	return func(s *FileService) {
		s.log = log
	}
}

// WithCryptoKey sets the crypto key for the service.
func WithCryptoKey(key string) Option {
	return func(s *FileService) {
		s.cryptoKey = key
	}
}

// WithObjectBasePath sets the object base path for the service.
func WithObjectBasePath(path string) Option {
	return func(s *FileService) {
		s.objBasePath = path
	}
}

type CreateFileRequest struct {
	Name     string
	Checksum string
	Size     int64
	Metadata map[string]string
	Data     io.Reader
}

func (s *FileService) CreateFile(ctx context.Context, userID, objID string, req CreateFileRequest) (*fileobj.File, error) {
	// TODO: Check file checksum before uploading.
	_ = req.Checksum

	// Create encrypted file data stream.
	stream, err := cryptutils.EncryptStream(s.cryptoKey, req.Data)
	if err != nil {
		return nil, fmt.Errorf("cryptutils.EncryptStream: %w", err)
	}

	objName := s.getObjName(userID, objID)

	// Put the file data to the object storage.
	info, err := s.objStorage.PutObject(ctx, objName, req.Size, stream)
	if err != nil {
		return nil, fmt.Errorf("storage.PutObject: %w", err)
	}

	// Create new file object entry to store it in the file storage.
	repoFile, err := fileobj.NewFile(
		objID, userID, req.Name, info.CRC32C(), info.Size(),
		req.Metadata, time.Now(), time.Now(), info.Location())
	if err != nil {
		return nil, fmt.Errorf("fileobj.NewFile: %w", err)
	}

	// Store the file object entry in the file storage.
	file, err := s.fileStorage.AddFile(ctx, repoFile)
	if err != nil {
		return nil, fmt.Errorf("storage.AddFile: %w", err)
	}

	return file, nil
}

func (s *FileService) getObjName(userID, objID string) string {
	objName := fmt.Sprintf("%s/%s", userID, objID)

	if s.objBasePath != "" {
		objName = fmt.Sprintf("%s/%s", s.objBasePath, objName)
	}

	return objName
}
