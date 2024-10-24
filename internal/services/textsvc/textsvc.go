// Package textsvc provides the file service.
package textsvc

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/pkg/errors"

	"github.com/andymarkow/gophkeeper/internal/cryptutils"
	"github.com/andymarkow/gophkeeper/internal/domain/vault/text"
	"github.com/andymarkow/gophkeeper/internal/storage/objrepo"
	"github.com/andymarkow/gophkeeper/internal/storage/textrepo"
)

var _ Service = (*SecretService)(nil)

// SecretService represents the secret data service.
type SecretService struct {
	log         *slog.Logger
	dbStorage   textrepo.Storage
	objStorage  objrepo.Storage
	objBasePath string
	cryptoKey   []byte
}

// NewSecretService creates a new service.
func NewSecretService(textstore textrepo.Storage, objstore objrepo.Storage, opts ...Option) *SecretService {
	svc := &SecretService{
		log:        slog.New(&slog.JSONHandler{}),
		dbStorage:  textstore,
		objStorage: objstore,
	}

	for _, opt := range opts {
		opt(svc)
	}

	return svc
}

// Option is a functional option for the text data service.
type Option func(*SecretService)

// WithLogger sets the logger for the text data service.
func WithLogger(log *slog.Logger) Option {
	return func(s *SecretService) {
		s.log = log
	}
}

// WithObjectBasePath sets the objects base path for the service.
func WithObjectBasePath(path string) Option {
	return func(s *SecretService) {
		s.objBasePath = path
	}
}

// WithCryptoKey sets the crypto key for the service.
func WithCryptoKey(key []byte) Option {
	return func(s *SecretService) {
		s.cryptoKey = key
	}
}

// Close closes the service.
func (s *SecretService) Close() error {
	if err := s.dbStorage.Close(); err != nil {
		return fmt.Errorf("storage.Close: %w", err)
	}

	return nil
}

// CreateSecret creates a new secret entry.
func (s *SecretService) CreateSecret(ctx context.Context, userID, secretName string, metadata map[string]string) (*text.Secret, error) {
	// Create new secret entry.
	text, err := text.CreateSecret(secretName, userID, metadata)
	if err != nil {
		return nil, fmt.Errorf("text.CreateSecret: %w", err)
	}

	// Store the text data entry in the text storage.
	txt, err := s.dbStorage.AddSecret(ctx, text)
	if err != nil {
		if errors.Is(err, textrepo.ErrSecretAlreadyExists) {
			return nil, fmt.Errorf("%w: %s", ErrSecretEntryAlreadyExists, secretName)
		}

		return nil, fmt.Errorf("storage.AddSecret: %w", err)
	}

	return txt, nil
}

// ListSecrets returns a list of secrets for the user.
func (s *SecretService) ListSecrets(ctx context.Context, userID string) ([]*text.Secret, error) {
	secrets, err := s.dbStorage.ListSecrets(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("storage.ListSecrets: %w", err)
	}

	return secrets, nil
}

// GetSecret returns a file for the user.
func (s *SecretService) GetSecret(ctx context.Context, userID, secretName string) (*text.Secret, error) {
	secret, err := s.dbStorage.GetSecret(ctx, userID, secretName)
	if err != nil {
		if errors.Is(err, textrepo.ErrSecretNotFound) {
			return nil, ErrSecretEntryNotFound
		}

		return nil, fmt.Errorf("storage.GetSecret: %w", err)
	}

	return secret, nil
}

// UpdateSecret updates the secret entry.
func (s *SecretService) UpdateSecret(ctx context.Context, userID, secretName string,
	metadata map[string]string) (*text.Secret, error) {
	// Get the secret entry from the DB storage.
	secret, err := s.dbStorage.GetSecret(ctx, userID, secretName)
	if err != nil {
		if errors.Is(err, textrepo.ErrSecretNotFound) {
			return nil, ErrSecretEntryNotFound
		}

		return nil, fmt.Errorf("storage.GetFile: %w", err)
	}

	if metadata != nil {
		secret.AddMetadata(metadata)
	}

	secret.SetUpdatedAt(time.Now())

	// Update the secret entry in the DB storage.
	secr, err := s.dbStorage.UpdateSecret(ctx, secret)
	if err != nil {
		return nil, fmt.Errorf("storage.UpdateSecret: %w", err)
	}

	return secr, nil
}

// DeleteSecret deletes a secret entry from the DB storage and the object storage.
func (s *SecretService) DeleteSecret(ctx context.Context, userID, secretName string) error {
	secret, err := s.dbStorage.GetSecret(ctx, userID, secretName)
	if err != nil {
		if errors.Is(err, textrepo.ErrSecretNotFound) {
			return ErrSecretEntryNotFound
		}

		return fmt.Errorf("storage.GetSecret: %w", err)
	}

	objName := s.getObjName(userID, secret.ID())

	err = s.objStorage.RemoveObject(ctx, objName)
	if err != nil {
		return fmt.Errorf("storage.RemoveObject: %w", err)
	}

	err = s.dbStorage.DeleteSecret(ctx, userID, secretName)
	if err != nil {
		if errors.Is(err, textrepo.ErrSecretNotFound) {
			return ErrSecretEntryNotFound
		}

		return fmt.Errorf("storage.DeleteSecret: %w", err)
	}

	return nil
}

// UploadSecret uploads a secret data to the object storage.
func (s *SecretService) UploadSecret(ctx context.Context, userID, secretName string, data io.Reader) (*text.Secret, error) {
	// Get the secret entry from the DB storage.
	// Secret entry must exists before the secret data is uploaded to the object storage.
	secret, err := s.dbStorage.GetSecret(ctx, userID, secretName)
	if err != nil {
		if errors.Is(err, textrepo.ErrSecretNotFound) {
			return nil, ErrSecretEntryNotFound
		}

		return nil, fmt.Errorf("storage.GetFile: %w", err)
	}

	rd, hash := cryptutils.CalcStreamHash(data)

	// Create encrypted data stream.
	stream, err := cryptutils.EncryptStream(s.cryptoKey, rd)
	if err != nil {
		return nil, fmt.Errorf("cryptutils.EncryptStream: %w", err)
	}

	objName := s.getObjName(userID, secret.ID())

	// Put the secret data to the object storage.
	info, err := s.objStorage.PutObject(ctx, objName, -1, stream.Stream())
	if err != nil {
		return nil, fmt.Errorf("storage.PutObject: %w", err)
	}

	// Calculate the checksum after the Reader has been fully read.
	checksum := hash.Sum32()

	// Create new content info object for the secret.
	contInfo := text.NewContentInfo(stream.SaltHex(), stream.IVHex(), info.Location(), fmt.Sprintf("%d", checksum))

	secret.SetContentInfo(contInfo)
	secret.SetUpdatedAt(time.Now())

	updFile, err := s.dbStorage.UpdateSecret(ctx, secret)
	if err != nil {
		return nil, fmt.Errorf("storage.UpdateSecret: %w", err)
	}

	return updFile, nil
}

// DownloadSecret downloads a secret data from the object storage.
func (s *SecretService) DownloadSecret(ctx context.Context, userID, secretName string) (*text.Secret, io.ReadCloser, error) {
	secret, err := s.dbStorage.GetSecret(ctx, userID, secretName)
	if err != nil {
		if errors.Is(err, textrepo.ErrSecretNotFound) {
			return nil, nil, ErrSecretEntryNotFound
		}

		return nil, nil, fmt.Errorf("storage.GetSecret: %w", err)
	}

	objName := s.getObjName(userID, secret.ID())

	if _, err := s.objStorage.GetObjectInfo(ctx, objName); err != nil {
		if errors.Is(err, objrepo.ErrObjNotExist) {
			return nil, nil, ErrSecretObjectNotFound
		}

		return nil, nil, fmt.Errorf("storage.GetObjectInfo: %w", err)
	}

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

// getObjName returns the object name for the given user and secret ID.
func (s *SecretService) getObjName(userID, secretID string) string {
	objName := fmt.Sprintf("%s/%s", userID, secretID)

	if s.objBasePath != "" {
		objName = fmt.Sprintf("%s/%s", s.objBasePath, objName)
	}

	return objName
}
