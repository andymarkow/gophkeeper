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
	"github.com/andymarkow/gophkeeper/internal/domain/vault/file"
	"github.com/andymarkow/gophkeeper/internal/storage/filerepo"
	"github.com/andymarkow/gophkeeper/internal/storage/objrepo"
)

var _ Service = (*FileService)(nil)

// FileService represents the service.
type FileService struct {
	log         *slog.Logger
	dbStorage   filerepo.Storage
	objStorage  objrepo.Storage
	objBasePath string
	cryptoKey   []byte
}

// NewFileService creates a new service.
func NewFileService(filestore filerepo.Storage, objstore objrepo.Storage, opts ...Option) *FileService {
	svc := &FileService{
		log:        slog.New(&slog.JSONHandler{}),
		dbStorage:  filestore,
		objStorage: objstore,
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

// CreateFile creates a new file entry in the file storage.
func (s *FileService) CreateFile(ctx context.Context, userID, secretName string, metadata map[string]string) (*file.Secret, error) {
	// Create new file object entry to store it in the file storage.
	secret, err := file.CreateSecret(secretName, userID, metadata)
	if err != nil {
		return nil, fmt.Errorf("file.CreateSecret: %w", err)
	}

	// Store the file object entry in the file storage.
	secretEntry, err := s.dbStorage.AddSecret(ctx, secret)
	if err != nil {
		if errors.Is(err, filerepo.ErrSecretAlreadyExists) {
			return nil, ErrFileEntryAlreadyExists
		}

		return nil, fmt.Errorf("storage.AddFile: %w", err)
	}

	return secretEntry, nil
}

// UpdateFile updates the file metadata.
func (s *FileService) UpdateFile(ctx context.Context, userID, secretName, fileName string,
	metadata map[string]string) (*file.Secret, error) {
	// Get the file object entry from the file storage.
	secret, err := s.dbStorage.GetSecret(ctx, userID, secretName)
	if err != nil {
		if errors.Is(err, filerepo.ErrSecretNotFound) {
			return nil, ErrFileEntryNotFound
		}

		return nil, fmt.Errorf("storage.GetFile: %w", err)
	}

	if metadata != nil {
		secret.AddMetadata(metadata)
	}

	if fileName != "" {
		secret.ContentInfo().SetFileName(fileName)
	}

	secret.SetUpdatedAt(time.Now())

	// Update the file object entry in the file storage.
	f, err := s.dbStorage.UpdateSecret(ctx, secret)
	if err != nil {
		return nil, fmt.Errorf("storage.UpdateSecret: %w", err)
	}

	return f, nil
}

// UploadFileRequest represents a request for uploading a file.
type UploadFileRequest struct {
	FileName string
	Size     int64
	Data     io.Reader
}

// UploadFile uploads a file content to the object storage.
func (s *FileService) UploadFile(ctx context.Context, userID, secretName string, req UploadFileRequest) (*file.Secret, error) {
	// Get the file object entry from the file storage.
	// This must exists before the file data is uploaded to the object storage.
	secret, err := s.dbStorage.GetSecret(ctx, userID, secretName)
	if err != nil {
		if errors.Is(err, filerepo.ErrSecretNotFound) {
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

	objName := s.getObjName(userID, secret.ID())

	// Put the file data to the object storage.
	info, err := s.objStorage.PutObject(ctx, objName, req.Size, stream.Stream())
	if err != nil {
		return nil, fmt.Errorf("storage.PutObject: %w", err)
	}

	// Calculate the checksum after the Reader has been fully read.
	checksum := hash.Sum32()

	// Create new content info object for the secret.
	contInfo, err := file.NewContentInfo(req.FileName, info.Location(), fmt.Sprintf("%d", checksum), stream.SaltHex(), stream.IVHex(), info.Size())
	if err != nil {
		return nil, fmt.Errorf("file.NewContentInfo: %w", err)
	}

	secret.SetContentInfo(contInfo)
	secret.SetUpdatedAt(time.Now())

	updFile, err := s.dbStorage.UpdateSecret(ctx, secret)
	if err != nil {
		return nil, fmt.Errorf("storage.UpdateSecret: %w", err)
	}

	return updFile, nil
}

// ListFiles returns a list of files for the user.
func (s *FileService) ListFiles(ctx context.Context, userID string) ([]*file.Secret, error) {
	files, err := s.dbStorage.ListSecrets(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("storage.ListSecrets: %w", err)
	}

	return files, nil
}

// GetFile returns a file for the user.
func (s *FileService) GetFile(ctx context.Context, userID, secretName string) (*file.Secret, error) {
	file, err := s.dbStorage.GetSecret(ctx, userID, secretName)
	if err != nil {
		if errors.Is(err, filerepo.ErrSecretNotFound) {
			return nil, ErrFileEntryNotFound
		}

		return nil, fmt.Errorf("storage.GetSecret: %w", err)
	}

	return file, nil
}

// DownloadFile downloads a file from the object storage.
func (s *FileService) DownloadFile(ctx context.Context, userID, secretName string) (*file.Secret, io.ReadCloser, error) {
	secret, err := s.dbStorage.GetSecret(ctx, userID, secretName)
	if err != nil {
		if errors.Is(err, filerepo.ErrSecretNotFound) {
			return nil, nil, ErrFileEntryNotFound
		}

		return nil, nil, fmt.Errorf("storage.GetSecret: %w", err)
	}

	objName := s.getObjName(userID, secret.ID())

	obj, err := s.objStorage.GetObject(ctx, objName)
	if err != nil {
		return nil, nil, fmt.Errorf("storage.GetObject: %w", err)
	}

	salt, err := hex.DecodeString(secret.ContentInfo().Salt())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode salt: %w", err)
	}

	iv, err := hex.DecodeString(secret.ContentInfo().IV())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode init vector: %w", err)
	}

	stream, err := cryptutils.DecryptStream(s.cryptoKey, salt, iv, obj)
	if err != nil {
		if err := obj.Close(); err != nil {
			s.log.Error("failed to close object reader", slog.String("object", objName), slog.Any("error", err))
		}

		return nil, nil, fmt.Errorf("cryptutils.DecryptStream: %w", err)
	}

	return secret, stream, nil
}

// DeleteFile deletes a file entry in the file storage and object storage.
func (s *FileService) DeleteFile(ctx context.Context, userID, secretName string) error {
	secret, err := s.dbStorage.GetSecret(ctx, userID, secretName)
	if err != nil {
		if errors.Is(err, filerepo.ErrSecretNotFound) {
			return ErrFileEntryNotFound
		}

		return fmt.Errorf("storage.GetSecret: %w", err)
	}

	objName := s.getObjName(userID, secret.ID())

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

	err = s.dbStorage.DeleteSecret(ctx, userID, secretName)
	if err != nil {
		if errors.Is(err, filerepo.ErrSecretNotFound) {
			return ErrFileEntryNotFound
		}

		return fmt.Errorf("storage.DeleteSecret: %w", err)
	}

	return nil
}

func (s *FileService) getObjName(userID, secretID string) string {
	objName := fmt.Sprintf("%s/%s", userID, secretID)

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
