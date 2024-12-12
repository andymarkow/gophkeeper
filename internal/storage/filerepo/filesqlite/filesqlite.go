// Package filesqlite provides SQLite storage implementation for file secrets.
package filesqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/andymarkow/gophkeeper/internal/domain/vault/file"
	"github.com/andymarkow/gophkeeper/internal/storage/filerepo"
)

var _ filerepo.Storage = (*Storage)(nil)

type Storage struct {
	log *slog.Logger
	db  *sql.DB
}

// NewStorage creates a new SQLite storage instance with the given connection string.
func NewStorage(dbPath string, opts ...Option) (*Storage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	store := &Storage{log: slog.New(&slog.JSONHandler{}), db: db}

	for _, opt := range opts {
		opt(store)
	}

	return store, nil
}

// Option is a storage option.
type Option func(*Storage)

// WithLogger is a storage option that sets logger.
func WithLogger(logger *slog.Logger) Option {
	return func(s *Storage) {
		s.log = logger
	}
}

// Close closes the underlying database connection.
func (s *Storage) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("db.Close: %w", err)
	}

	return nil
}

// Ping pings the underlying database connection.
func (s *Storage) Ping(ctx context.Context) error {
	if err := s.db.PingContext(ctx); err != nil {
		return fmt.Errorf("db.PingContext: %w", err)
	}

	return nil
}

// AddSecret adds a file secret entry to the storage.
func (s *Storage) AddSecret(ctx context.Context, secret *file.Secret) (*file.Secret, error) {
	query := `INSERT INTO vault_files
			(id, name, user_id, created_at, updated_at, version, metadata, salt, iv, filename, location, checksum, size)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	metadata, err := secret.MetadataJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	_, err = s.db.ExecContext(ctx, query,
		secret.ID(), secret.Name(), secret.UserID(), secret.CreatedAt(), secret.UpdatedAt(), secret.Version(), metadata,
		secret.ContentInfo().Salt(), secret.ContentInfo().IV(), secret.ContentInfo().FileName(),
		secret.ContentInfo().Location(), secret.ContentInfo().Checksum(), secret.ContentInfo().Size())
	if err != nil {
		return nil, fmt.Errorf("db.ExecContext: %w", err)
	}

	secr, err := s.GetSecret(ctx, secret.UserID(), secret.Name())
	if err != nil {
		return nil, fmt.Errorf("storage.GetSecret: %w", err)
	}

	return secr, nil
}

// GetSecret gets a file secret entry from the storage.
func (s *Storage) GetSecret(ctx context.Context, userID, name string) (*file.Secret, error) {
	query := `SELECT id, name, user_id, created_at, updated_at, version, metadata, salt, iv, filename, location, checksum, size
			FROM vault_files
			WHERE user_id = $1 AND name = $2`

	row := s.db.QueryRowContext(ctx, query, userID, name)

	var dbSecret filerepo.Secret

	err := row.Scan(&dbSecret.ID, &dbSecret.Name, &dbSecret.UserID, &dbSecret.CreatedAt, &dbSecret.UpdatedAt,
		&dbSecret.Version, &dbSecret.Metadata, &dbSecret.Salt, &dbSecret.IV, &dbSecret.FileName,
		&dbSecret.Location, &dbSecret.Checksum, &dbSecret.Size)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, filerepo.ErrSecretNotFound
		}

		return nil, fmt.Errorf("db.QueryRowContext: %w", err)
	}

	var metadata map[string]string

	err = json.Unmarshal([]byte(dbSecret.Metadata), &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	contentInfo := file.NewContentInfo(dbSecret.Salt, dbSecret.IV, dbSecret.FileName, dbSecret.Location, dbSecret.Checksum, dbSecret.Size)

	secret, err := file.NewSecret(dbSecret.ID, dbSecret.Name, dbSecret.UserID, metadata,
		dbSecret.CreatedAt, dbSecret.UpdatedAt, dbSecret.Version, contentInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to create file secret: %w", err)
	}

	return secret, nil
}

// ListSecrets returns a list of file secret entries from the storage.
func (s *Storage) ListSecrets(ctx context.Context, userID string) ([]*file.Secret, error) {
	query := `SELECT id, name, user_id, created_at, updated_at, version, metadata, salt, iv, filename, location, checksum, size
			FROM vault_files WHERE user_id = $1`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("db.QueryContext: %w", err)
	}
	defer rows.Close()

	var secrets []*file.Secret

	for rows.Next() {
		var dbSecret filerepo.Secret

		err := rows.Scan(&dbSecret.ID, &dbSecret.Name, &dbSecret.UserID, &dbSecret.CreatedAt, &dbSecret.UpdatedAt,
			&dbSecret.Version, &dbSecret.Metadata, &dbSecret.Salt, &dbSecret.IV, &dbSecret.FileName,
			&dbSecret.Location, &dbSecret.Checksum, &dbSecret.Size)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		var metadata map[string]string

		err = json.Unmarshal([]byte(dbSecret.Metadata), &metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		contentInfo := file.NewContentInfo(dbSecret.Salt, dbSecret.IV, dbSecret.FileName, dbSecret.Location, dbSecret.Checksum, dbSecret.Size)

		secret, err := file.NewSecret(dbSecret.ID, dbSecret.Name, dbSecret.UserID, metadata,
			dbSecret.CreatedAt, dbSecret.UpdatedAt, dbSecret.Version, contentInfo)
		if err != nil {
			return nil, fmt.Errorf("failed to create file secret: %w", err)
		}

		secrets = append(secrets, secret)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return secrets, nil
}

// UpdateSecret updates a file secret entry in the storage.
func (s *Storage) UpdateSecret(ctx context.Context, secret *file.Secret) (*file.Secret, error) {
	query := `UPDATE vault_files
		SET updated_at = $1, version = $2, metadata = $3, salt = $4, iv = $5, filename = $6, location = $7, checksum = $8, size = $9
		WHERE user_id = $10 AND name = $11`

	_, err := s.db.ExecContext(ctx, query, secret.UpdatedAt, secret.Version, secret.Metadata,
		secret.ContentInfo().Salt(), secret.ContentInfo().IV(), secret.ContentInfo().FileName(),
		secret.ContentInfo().Location(), secret.ContentInfo().Checksum(), secret.ContentInfo().Size(), secret.UserID(), secret.Name())
	if err != nil {
		return nil, fmt.Errorf("db.ExecContext: %w", err)
	}

	secr, err := s.GetSecret(ctx, secret.UserID(), secret.Name())
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return secr, nil
}

// DeleteSecret removes a file secret entry from the storage.
func (s *Storage) DeleteSecret(ctx context.Context, userID, secretName string) error {
	query := `DELETE FROM vault_files WHERE user_id = $1 AND name = $2`

	result, err := s.db.ExecContext(ctx, query, userID, secretName)
	if err != nil {
		return fmt.Errorf("db.ExecContext: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("result.RowsAffected: %w", err)
	}

	if rowsAffected == 0 {
		return filerepo.ErrSecretNotFound
	}

	return nil
}
