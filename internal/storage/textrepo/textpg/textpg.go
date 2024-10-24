// Package textpg provides PostgreSQL storage implementation for text secrets.
package textpg

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	// Postgres driver.
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/andymarkow/gophkeeper/internal/domain/vault/text"
	"github.com/andymarkow/gophkeeper/internal/pgutils"
	"github.com/andymarkow/gophkeeper/internal/storage/textrepo"
)

// Storage implements PostgreSQL storage.
type Storage struct {
	db  *sql.DB
	log *slog.Logger
}

// config is the configuration for PostgreSQL storage.
type config struct {
	log             *slog.Logger
	maxOpenConns    int
	maxIdleConns    int
	connMaxIdleTime time.Duration
	connMaxLifetime time.Duration
}

// NewStorage creates a new PostgreSQL storage instance with the given connection string.
func NewStorage(connStr string, opts ...Option) (*Storage, error) {
	cfg := &config{
		log:             slog.New(&slog.JSONHandler{}),
		maxOpenConns:    10,
		maxIdleConns:    5,
		connMaxIdleTime: 180 * time.Second,
		connMaxLifetime: 3600 * time.Second,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	db.SetMaxOpenConns(cfg.maxOpenConns)
	db.SetMaxIdleConns(cfg.maxIdleConns)
	db.SetConnMaxIdleTime(cfg.connMaxIdleTime)
	db.SetConnMaxLifetime(cfg.connMaxLifetime)

	return &Storage{db: db, log: cfg.log}, nil
}

// Option is a PostgreSQL storage option.
type Option func(s *config)

// WithLogger sets the logger.
func WithLogger(logger *slog.Logger) Option {
	return func(c *config) {
		c.log = logger
	}
}

// WithMaxOpenConns sets the maximum number of open connections to the database.
func WithMaxOpenConns(conns int) Option {
	return func(c *config) {
		c.maxOpenConns = conns
	}
}

// WithMaxIdleConns sets the maximum number of idle connections in the connection pool.
func WithMaxIdleConns(conns int) Option {
	return func(c *config) {
		c.maxIdleConns = conns
	}
}

// WithConnMaxIdleTime sets the maximum amount of time a connection may be reused.
func WithConnMaxIdleTime(idleTime time.Duration) Option {
	return func(c *config) {
		c.connMaxIdleTime = idleTime
	}
}

// WithConnMaxLifetime sets the maximum amount of time a connection may be reused.
func WithConnMaxLifetime(lifetime time.Duration) Option {
	return func(c *config) {
		c.connMaxLifetime = lifetime
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
	err := pgutils.WithRetry(func() error {
		if err := s.db.PingContext(ctx); err != nil {
			return fmt.Errorf("db.PingContext: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

// AddSecret adds a text secret entry to the storage.
func (s *Storage) AddSecret(ctx context.Context, secret *text.Secret) (*text.Secret, error) {
	err := pgutils.WithRetry(func() error {
		query := `INSERT INTO vault_texts
			(id, name, user_id, created_at, updated_at, metadata, salt, iv, location, checksum)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

		_, err := s.db.ExecContext(ctx, query,
			secret.ID(), secret.Name(), secret.UserID(), secret.CreatedAt(), secret.UpdatedAt(), secret.Metadata(),
			secret.ContentInfo().Salt(), secret.ContentInfo().IV(), secret.ContentInfo().Location(), secret.ContentInfo().Checksum())
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
				return textrepo.ErrSecretAlreadyExists
			}

			return fmt.Errorf("db.ExecContext: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	secr, err := s.GetSecret(ctx, secret.UserID(), secret.Name())
	if err != nil {
		return nil, fmt.Errorf("storage.GetSecret: %w", err)
	}

	return secr, nil
}

// GetSecret returns a text secret entry from the storage.
func (s *Storage) GetSecret(ctx context.Context, userID, name string) (*text.Secret, error) {
	var dbSecret textrepo.Secret

	err := pgutils.WithRetry(func() error {
		query := `SELECT id, name, user_id, created_at, updated_at, metadata, salt, iv, location, checksum
			FROM vault_texts
			WHERE user_id = $1 AND name = $2`

		row := s.db.QueryRowContext(ctx, query, userID, name)

		err := row.Scan(&dbSecret.ID, &dbSecret.Name, &dbSecret.UserID, &dbSecret.CreatedAt, &dbSecret.UpdatedAt,
			&dbSecret.Metadata, &dbSecret.Salt, &dbSecret.IV, &dbSecret.Location, &dbSecret.Checksum)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return textrepo.ErrSecretNotFound
			}

			return fmt.Errorf("row.Scan: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	var metadata map[string]string

	err = json.Unmarshal([]byte(dbSecret.Metadata), &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	info := text.NewContentInfo(dbSecret.Salt, dbSecret.IV, dbSecret.Location, dbSecret.Checksum)

	secret, err := text.NewSecret(dbSecret.ID, dbSecret.Name, dbSecret.UserID, metadata,
		dbSecret.CreatedAt, dbSecret.UpdatedAt, info)
	if err != nil {
		return nil, fmt.Errorf("failed to create text secret: %w", err)
	}

	return secret, nil
}

// ListSecrets returns a list of text secret entries from the storage.
func (s *Storage) ListSecrets(ctx context.Context, userID string) ([]*text.Secret, error) {
	var dbSecrets []textrepo.Secret

	err := pgutils.WithRetry(func() error {
		query := `SELECT id, name, user_id, created_at, updated_at, metadata, location, checksum
			FROM vault_texts WHERE user_id = $1`

		rows, err := s.db.QueryContext(ctx, query, userID)
		if err != nil {
			return fmt.Errorf("db.QueryContext: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var dbSecret textrepo.Secret

			err := rows.Scan(&dbSecret.ID, &dbSecret.Name, &dbSecret.UserID, &dbSecret.CreatedAt,
				&dbSecret.UpdatedAt, &dbSecret.Metadata, &dbSecret.Location, &dbSecret.Checksum)
			if err != nil {
				return fmt.Errorf("rows.Scan: %w", err)
			}

			dbSecrets = append(dbSecrets, dbSecret)
		}

		if err := rows.Err(); err != nil {
			return fmt.Errorf("rows.Err: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	secrets := make([]*text.Secret, 0, len(dbSecrets))

	for _, dbSecret := range dbSecrets {
		var metadata map[string]string

		err = json.Unmarshal([]byte(dbSecret.Metadata), &metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		info := text.NewContentInfo("", "", dbSecret.Location, dbSecret.Checksum)

		secret, err := text.NewSecret(dbSecret.ID, dbSecret.Name, dbSecret.UserID, metadata,
			dbSecret.CreatedAt, dbSecret.UpdatedAt, info)
		if err != nil {
			return nil, fmt.Errorf("failed to create text secret: %w", err)
		}

		secrets = append(secrets, secret)
	}

	return secrets, nil
}

// UpdateSecret updates a text secret entry in the storage.
func (s *Storage) UpdateSecret(ctx context.Context, secret *text.Secret) (*text.Secret, error) {
	err := pgutils.WithRetry(func() error {
		query := `UPDATE vault_texts
			SET updated_at = $1, metadata = $2, salt = $3, iv = $4, location = $5, checksum = $6
			WHERE user_id = $7 AND name = $8`

		_, err := s.db.ExecContext(ctx, query, secret.UpdatedAt(), secret.Metadata(),
			secret.ContentInfo().Salt(), secret.ContentInfo().IV(), secret.ContentInfo().Location(),
			secret.ContentInfo().Checksum(), secret.UserID(), secret.Name())
		if err != nil {
			return fmt.Errorf("db.ExecContext: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	secr, err := s.GetSecret(ctx, secret.UserID(), secret.Name())
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return secr, nil
}

// DeleteSecret deletes a text secret entry from the storage.
func (s *Storage) DeleteSecret(ctx context.Context, userID, secretName string) error {
	err := pgutils.WithRetry(func() error {
		query := `DELETE FROM vault_texts WHERE user_id = $1 AND name = $2`

		result, err := s.db.ExecContext(ctx, query, userID, secretName)
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
	})
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}
