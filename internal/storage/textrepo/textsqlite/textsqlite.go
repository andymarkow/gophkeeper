// Package textsqlite provides SQLite storage implementation for text secrets.
package textsqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/andymarkow/gophkeeper/internal/domain/vault/text"
	"github.com/andymarkow/gophkeeper/internal/storage/textrepo"
)

var _ textrepo.Storage = (*Storage)(nil)

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

// AddSecret adds a text secret entry to the storage.
func (s *Storage) AddSecret(ctx context.Context, secret *text.Secret) (*text.Secret, error) {
	metadata, err := secret.MetadataJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	query := `INSERT INTO vault_texts
		(id, name, user_id, created_at, updated_at, version, metadata, salt, iv, location, checksum)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err = s.db.ExecContext(ctx, query,
		secret.ID(), secret.Name(), secret.UserID(), secret.CreatedAt(), secret.UpdatedAt(), secret.Version(), metadata,
		secret.ContentInfo().Salt(), secret.ContentInfo().IV(), secret.ContentInfo().Location(), secret.ContentInfo().Checksum())
	if err != nil {
		return nil, fmt.Errorf("db.ExecContext: %w", err)
	}

	secr, err := s.GetSecret(ctx, secret.UserID(), secret.Name())
	if err != nil {
		return nil, fmt.Errorf("storage.GetSecret: %w", err)
	}

	return secr, nil
}

// GetSecret returns a text secret entry from the storage.
func (s *Storage) GetSecret(ctx context.Context, userID, name string) (*text.Secret, error) {
	query := `SELECT id, name, user_id, created_at, updated_at, version, metadata, salt, iv, location, checksum
		FROM vault_texts
		WHERE user_id = $1 AND name = $2`

	row := s.db.QueryRowContext(ctx, query, userID, name)

	var dbSecret textrepo.Secret

	err := row.Scan(&dbSecret.ID, &dbSecret.Name, &dbSecret.UserID, &dbSecret.CreatedAt, &dbSecret.UpdatedAt,
		&dbSecret.Version, &dbSecret.Metadata, &dbSecret.Salt, &dbSecret.IV, &dbSecret.Location, &dbSecret.Checksum)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, textrepo.ErrSecretNotFound
		}

		return nil, fmt.Errorf("row.Scan: %w", err)
	}

	var metadata map[string]string

	err = json.Unmarshal([]byte(dbSecret.Metadata), &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	info := text.NewContentInfo(dbSecret.Salt, dbSecret.IV, dbSecret.Location, dbSecret.Checksum)

	secret, err := text.NewSecret(dbSecret.ID, dbSecret.Name, dbSecret.UserID, metadata,
		dbSecret.CreatedAt, dbSecret.UpdatedAt, dbSecret.Version, info)
	if err != nil {
		return nil, fmt.Errorf("failed to create text secret: %w", err)
	}

	return secret, nil
}

// ListSecrets returns a list of text secret entries from the storage.
func (s *Storage) ListSecrets(ctx context.Context, userID string) ([]*text.Secret, error) {
	query := `SELECT id, name, user_id, created_at, updated_at, version, metadata, salt, iv, location, checksum
		FROM vault_texts
		WHERE user_id = $1`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("db.QueryContext: %w", err)
	}
	defer rows.Close()

	var secrets []*text.Secret

	for rows.Next() {
		var dbSecret textrepo.Secret

		err := rows.Scan(&dbSecret.ID, &dbSecret.Name, &dbSecret.UserID, &dbSecret.CreatedAt, &dbSecret.UpdatedAt,
			&dbSecret.Version, &dbSecret.Metadata, &dbSecret.Salt, &dbSecret.IV, &dbSecret.Location, &dbSecret.Checksum)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		var metadata map[string]string

		err = json.Unmarshal([]byte(dbSecret.Metadata), &metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		info := text.NewContentInfo(dbSecret.Salt, dbSecret.IV, dbSecret.Location, dbSecret.Checksum)

		secret, err := text.NewSecret(dbSecret.ID, dbSecret.Name, dbSecret.UserID, metadata,
			dbSecret.CreatedAt, dbSecret.UpdatedAt, dbSecret.Version, info)
		if err != nil {
			return nil, fmt.Errorf("failed to create text secret: %w", err)
		}

		secrets = append(secrets, secret)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return secrets, nil
}

// UpdateSecret updates a text secret entry in the storage.
func (s *Storage) UpdateSecret(ctx context.Context, secret *text.Secret) (*text.Secret, error) {
	metadata, err := json.Marshal(secret.Metadata())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `UPDATE vault_texts
		SET updated_at = $1, version = $2, metadata = $3, salt = $4, iv = $5, location = $6, checksum = $7
		WHERE user_id = $8 AND name = $9`

	_, err = s.db.ExecContext(ctx, query, secret.UpdatedAt(), secret.Version(), metadata,
		secret.ContentInfo().Salt(), secret.ContentInfo().IV(), secret.ContentInfo().Location(),
		secret.ContentInfo().Checksum(), secret.UserID(), secret.Name())
	if err != nil {
		return nil, fmt.Errorf("db.ExecContext: %w", err)
	}

	secr, err := s.GetSecret(ctx, secret.UserID(), secret.Name())
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return secr, nil
}

// DeleteSecret deletes a text secret entry from the storage.
func (s *Storage) DeleteSecret(ctx context.Context, userID, name string) error {
	query := `DELETE FROM vault_texts WHERE user_id = $1 AND name = $2`

	result, err := s.db.ExecContext(ctx, query, userID, name)
	if err != nil {
		return fmt.Errorf("db.ExecContext: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("result.RowsAffected: %w", err)
	}

	if rowsAffected == 0 {
		return textrepo.ErrSecretNotFound
	}

	return nil
}
