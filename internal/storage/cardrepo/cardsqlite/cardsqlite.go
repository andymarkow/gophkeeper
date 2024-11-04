// Package cardsqlite provides SQLite storage implementation for bank cards.
package cardsqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	// SQLite driver.
	"github.com/mattn/go-sqlite3"

	"github.com/andymarkow/gophkeeper/internal/domain/vault/bankcard"
	"github.com/andymarkow/gophkeeper/internal/storage/cardrepo"
)

// Storage implements bank card storage.
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

// AddSecret adds a bank card secret entry to the storage.
func (s *Storage) AddSecret(ctx context.Context, secret *bankcard.Secret) (*bankcard.Secret, error) {
	metadata, err := secret.MetadataJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	data, err := secret.DataJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	query := `INSERT INTO vault_bankcards
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

// GetSecret gets a bank card secret entry from the storage.
func (s *Storage) GetSecret(ctx context.Context, userID, secretName string) (*bankcard.Secret, error) {
	query := `SELECT id, name, user_id, created_at, updated_at, version, metadata, data
			FROM vault_bankcards
			WHERE user_id = $1 AND name = $2`

	var dbSecret cardrepo.Secret

	row := s.db.QueryRowContext(ctx, query, userID, secretName)

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

	data, err := bankcard.UnmarshalData([]byte(dbSecret.Data))
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	secret, err := bankcard.NewSecret(dbSecret.ID, dbSecret.Name, dbSecret.UserID, metadata,
		dbSecret.CreatedAt, dbSecret.UpdatedAt, dbSecret.Version, data)
	if err != nil {
		return nil, fmt.Errorf("failed to create bank card secret: %w", err)
	}

	return secret, nil
}
