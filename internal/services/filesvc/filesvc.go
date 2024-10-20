// Package filesvc provides the file service.
package filesvc

import (
	"context"
	"encoding/hex"
	"errors"
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
	cryptoKey   []byte
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
func WithCryptoKey(key []byte) Option {
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

// CreateFileRequest represents a request for creating a file.
type CreateFileRequest struct {
	Name     string
	Size     int64
	Metadata map[string]string
	Data     io.Reader
}

// CreateFile creates a new file entry in the file storage.
func (s *FileService) CreateFile(ctx context.Context, userID, fileID string, metadata map[string]string) (*fileobj.File, error) {
	// Create new file object entry to store it in the file storage.
	repoFile, err := fileobj.CreateEmptyFileEntry(fileID, userID, metadata)
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

// UpdateFile updates the file metadata.
func (s *FileService) UpdateFile(ctx context.Context, userID, fileID, fileName string, metadata map[string]string) (*fileobj.File, error) {
	// Get the file object entry from the file storage.
	file, err := s.fileStorage.GetFile(ctx, userID, fileID)
	if err != nil {
		if errors.Is(err, filerepo.ErrFileNotFound) {
			return nil, ErrFileEntryNotFound
		}

		return nil, fmt.Errorf("storage.GetFile: %w", err)
	}

	if fileName != "" {
		file.SetName(fileName)
	}

	if metadata != nil {
		file.AddMetadata(metadata)
	}

	file.SetUpdatedAt(time.Now())

	// Update the file object entry in the file storage.
	f, err := s.fileStorage.UpdateFile(ctx, file)
	if err != nil {
		return nil, fmt.Errorf("storage.UpdateFile: %w", err)
	}

	return f, nil
}

// UploadFileRequest represents a request for uploading a file.
type UploadFileRequest struct {
	Name string
	Size int64
	Data io.Reader
}

// UploadFile uploads a file content to the object storage.
func (s *FileService) UploadFile(ctx context.Context, userID, objID string, req UploadFileRequest) (*fileobj.File, error) {
	// Get the file object entry from the file storage.
	// This must exists before the file data is uploaded to the object storage.
	f, err := s.fileStorage.GetFile(ctx, userID, objID)
	if err != nil {
		if errors.Is(err, filerepo.ErrFileNotFound) {
			return nil, ErrFileEntryNotFound
		}

		return nil, fmt.Errorf("storage.GetFile: %w", err)
	}

	rd, hash := cryptutils.CalcStreamHash(req.Data)

	// Create encrypted file data stream.
	stream, err := cryptutils.EncryptStream(s.cryptoKey, rd)
	if err != nil {
		return nil, fmt.Errorf("cryptutils.EncryptStream: %w", err)
	}

	objName := s.getObjName(userID, objID)

	s.log.Debug("uploading file to the object storage", slog.String("object", objName))

	// Put the file data to the object storage.
	info, err := s.objStorage.PutObject(ctx, objName, req.Size, stream.Stream())
	if err != nil {
		return nil, fmt.Errorf("storage.PutObject: %w", err)
	}

	// Calculate the checksum after the Reader has been fully read.
	checksum := hash.Sum32()

	s.log.Debug("uploaded file to the object storage", slog.String("object", objName))

	fileEntry, err := fileobj.NewFile(f.ID(), f.UserID(), req.Name, info.Location(), fmt.Sprintf("%d", checksum),
		stream.SaltHex(), stream.IVHex(), info.Size(), f.Metadata(), f.CreatedAt(), time.Now())
	if err != nil {
		return nil, fmt.Errorf("fileobj.NewFile: %w", err)
	}

	s.log.Debug("updating file entry in the file storage", slog.String("object", objName))

	updFile, err := s.fileStorage.UpdateFile(ctx, fileEntry)
	if err != nil {
		return nil, fmt.Errorf("storage.UpdateFile: %w", err)
	}

	s.log.Debug("updated file entry in the file storage", slog.String("object", objName))

	return updFile, nil
}

// ListFiles returns a list of files for the user.
func (s *FileService) ListFiles(ctx context.Context, userID string) ([]*fileobj.File, error) {
	files, err := s.fileStorage.ListFiles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("storage.ListFiles: %w", err)
	}

	return files, nil
}

// GetFile returns a file for the user.
func (s *FileService) GetFile(ctx context.Context, userID, objID string) (*fileobj.File, error) {
	file, err := s.fileStorage.GetFile(ctx, userID, objID)
	if err != nil {
		if errors.Is(err, filerepo.ErrFileNotFound) {
			return nil, ErrFileEntryNotFound
		}

		return nil, fmt.Errorf("storage.GetFile: %w", err)
	}

	return file, nil
}

// DownloadFile downloads a file from the object storage.
func (s *FileService) DownloadFile(ctx context.Context, userID, objID string) (*fileobj.File, io.ReadCloser, error) {
	file, err := s.fileStorage.GetFile(ctx, userID, objID)
	if err != nil {
		if errors.Is(err, filerepo.ErrFileNotFound) {
			return nil, nil, ErrFileEntryNotFound
		}

		return nil, nil, fmt.Errorf("storage.GetFile: %w", err)
	}

	objName := s.getObjName(userID, objID)

	obj, err := s.objStorage.GetObject(ctx, objName)
	if err != nil {
		return nil, nil, fmt.Errorf("storage.GetObject: %w", err)
	}

	salt, err := hex.DecodeString(file.Salt())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode salt: %w", err)
	}

	iv, err := hex.DecodeString(file.IV())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode iv: %w", err)
	}

	stream, err := cryptutils.DecryptStream(s.cryptoKey, salt, iv, obj)
	if err != nil {
		if err := obj.Close(); err != nil {
			s.log.Error("failed to close object reader", slog.String("object", objName), slog.Any("error", err))
		}

		return nil, nil, fmt.Errorf("cryptutils.DecryptStream: %w", err)
	}

	return file, stream, nil
}

// DeleteFile deletes a file entry in the file storage and object storage.
func (s *FileService) DeleteFile(ctx context.Context, userID, objID string) error {
	objName := s.getObjName(userID, objID)

	_, found, err := s.statObject(ctx, objName)
	if err != nil {
		return fmt.Errorf("failed to stat object: %w", err)
	}

	if !found {
		return ErrFileObjectNotFound
	}

	err = s.objStorage.RemoveObject(ctx, objName)
	if err != nil {
		return fmt.Errorf("storage.RemoveObject: %w", err)
	}

	err = s.fileStorage.DeleteFile(ctx, userID, objID)
	if err != nil {
		if errors.Is(err, filerepo.ErrFileNotFound) {
			return ErrFileEntryNotFound
		}

		return fmt.Errorf("storage.RemoveFile: %w", err)
	}

	return nil
}

func (s *FileService) getObjName(userID, objID string) string {
	objName := fmt.Sprintf("%s/%s", userID, objID)

	if s.objBasePath != "" {
		objName = fmt.Sprintf("%s/%s", s.objBasePath, objName)
	}

	return objName
}

func (s *FileService) statObject(ctx context.Context, objName string) (*objrepo.ObjectInfo, bool, error) {
	info, err := s.objStorage.GetObjectInfo(ctx, objName)
	if err != nil {
		if errors.Is(err, objrepo.ErrObjNotExist) {
			return nil, false, nil
		}

		return nil, false, fmt.Errorf("storage.GetObjectInfo: %w", err)
	}

	return info, true, nil
}
