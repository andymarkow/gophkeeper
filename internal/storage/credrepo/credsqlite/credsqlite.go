// Package credsqlite implements SQLite storage.
package credsqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/mattn/go-sqlite3"

	"github.com/andymarkow/gophkeeper/internal/domain/vault/credential"
	"github.com/andymarkow/gophkeeper/internal/storage/cardrepo"
	"github.com/andymarkow/gophkeeper/internal/storage/credrepo"
)

var _ credrepo.Storage = (*Storage)(nil)

// Storage implements SQLite storage.
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

// AddSecret adds a credential secret entry to the storage.
func (s *Storage) AddSecret(ctx context.Context, secret *credential.Secret) (*credential.Secret, error) {
	metadata, err := secret.MetadataJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	data, err := secret.DataJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	query := `INSERT INTO vault_credentials
			(id, name, user_id, created_at, updated_at, version, metadata, data)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	if _, err := s.db.ExecContext(ctx, query,
		secret.ID(), secret.Name(), secret.UserID(), secret.CreatedAt(), secret.UpdatedAt(), secret.Version(), metadata, data); err != nil {
		var sqliteErr sqlite3.Error

		if errors.As(err, &sqliteErr) && sqliteErr.Code == sqlite3.ErrNo(sqlite3.ErrConstraintUnique) {
			return nil, cardrepo.ErrSecretAlreadyExists
		}

		return nil, fmt.Errorf("db.ExecContext: %w", err)
	}

	secr, err := s.GetSecret(ctx, secret.UserID(), secret.Name())
	if err != nil {
		return nil, fmt.Errorf("storage.GetSecret: %w", err)
	}

	return secr, nil
}

// GetSecret gets a credential secret entry from the storage.
func (s *Storage) GetSecret(ctx context.Context, userID, name string) (*credential.Secret, error) {
	query := `SELECT id, name, user_id, created_at, updated_at, version, metadata, data
			FROM vault_credentials
			WHERE user_id = $1 AND name = $2`

	var dbSecret cardrepo.Secret

	row := s.db.QueryRowContext(ctx, query, userID, name)

	err := row.Scan(&dbSecret.ID, &dbSecret.Name, &dbSecret.UserID, &dbSecret.CreatedAt,
		&dbSecret.UpdatedAt, &dbSecret.Version, &dbSecret.Metadata, &dbSecret.Data)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, cardrepo.ErrSecretNotFound
		}

		return nil, fmt.Errorf("db.QueryRowContext: %w", err)
	}

	var metadata map[string]string

	err = json.Unmarshal([]byte(dbSecret.Metadata), &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	data, err := credential.UnmarshalData([]byte(dbSecret.Data))
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	secret, err := credential.NewSecret(dbSecret.ID, dbSecret.Name, dbSecret.UserID, metadata,
		dbSecret.CreatedAt, dbSecret.UpdatedAt, dbSecret.Version, data)
	if err != nil {
		return nil, fmt.Errorf("failed to create secret: %w", err)
	}

	return secret, nil
}

// ListSecrets returns a list of credential secret entries from the storage.
func (s *Storage) ListSecrets(ctx context.Context, userID string) ([]*credential.Secret, error) {
	query := `SELECT id, name, user_id, created_at, updated_at, version, metadata
			FROM vault_credentials
			WHERE user_id = $1`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("db.QueryContext: %w", err)
	}
	defer rows.Close()

	var secrets []*credential.Secret

	for rows.Next() {
		var dbSecret credrepo.Secret

		if err := rows.Scan(&dbSecret.ID, &dbSecret.Name, &dbSecret.UserID, &dbSecret.CreatedAt,
			&dbSecret.UpdatedAt, &dbSecret.Version, &dbSecret.Metadata); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		var metadata map[string]string

		err = json.Unmarshal([]byte(dbSecret.Metadata), &metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		secret, err := credential.NewSecret(dbSecret.ID, dbSecret.Name, dbSecret.UserID, metadata,
			dbSecret.CreatedAt, dbSecret.UpdatedAt, dbSecret.Version, credential.NewEmptyData())
		if err != nil {
			return nil, fmt.Errorf("failed to create secret: %w", err)
		}

		secrets = append(secrets, secret)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return secrets, nil
}

// UpdateSecret updates a credential secret entry in the storage.
func (s *Storage) UpdateSecret(ctx context.Context, secret *credential.Secret) (*credential.Secret, error) {
	metadata, err := secret.MetadataJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	data, err := secret.DataJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	query := `UPDATE vault_credentials
			SET updated_at = $1, version = $2, metadata = $3, data = $4
			WHERE user_id = $5 AND name = $6`

	if _, err := s.db.ExecContext(context.Background(), query, secret.UpdatedAt(), secret.Version(),
		metadata, data, secret.UserID(), secret.Name()); err != nil {
		return nil, fmt.Errorf("db.ExecContext: %w", err)
	}

	secr, err := s.GetSecret(ctx, secret.UserID(), secret.Name())
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return secr, nil
}

// DeleteSecret deletes a credential secret entry from the storage.
func (s *Storage) DeleteSecret(ctx context.Context, userID, secretName string) error {
	query := `DELETE FROM vault_credentials
			WHERE user_id = $1 AND name = $2`

	result, err := s.db.ExecContext(ctx, query, userID, secretName)
	if err != nil {
		return fmt.Errorf("db.ExecContext: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("result.RowsAffected: %w", err)
	}

	if rowsAffected == 0 {
		return credrepo.ErrSecretNotFound
	}

	return nil
}
