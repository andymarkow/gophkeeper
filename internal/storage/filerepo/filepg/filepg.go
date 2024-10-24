// Package filepg provides PostgreSQL storage implementation for file secrets.
package filepg

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

	"github.com/andymarkow/gophkeeper/internal/domain/vault/file"
	"github.com/andymarkow/gophkeeper/internal/pgutils"
	"github.com/andymarkow/gophkeeper/internal/storage/filerepo"
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

// AddSecret adds a file secret entry to the storage.
func (s *Storage) AddSecret(ctx context.Context, secret *file.Secret) (*file.Secret, error) {
	err := pgutils.WithRetry(func() error {
		query := `INSERT INTO vault_files
			(id, name, user_id, created_at, updated_at, metadata, salt, iv, filename, location, checksum, size)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

		_, err := s.db.ExecContext(ctx, query,
			secret.ID(), secret.Name(), secret.UserID(), secret.CreatedAt(), secret.UpdatedAt(), secret.Metadata(),
			secret.ContentInfo().Salt(), secret.ContentInfo().IV(), secret.ContentInfo().FileName(),
			secret.ContentInfo().Location(), secret.ContentInfo().Checksum(), secret.ContentInfo().Size())
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
				return filerepo.ErrSecretAlreadyExists
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

// GetSecret returns a file secret entry from the storage.
func (s *Storage) GetSecret(ctx context.Context, userID, name string) (*file.Secret, error) {
	var dbSecret filerepo.Secret

	err := pgutils.WithRetry(func() error {
		query := `SELECT id, name, user_id, created_at, updated_at, metadata, salt, iv, filename, location, checksum, size
			FROM vault_files
			WHERE user_id = $1 AND name = $2`

		row := s.db.QueryRowContext(ctx, query, userID, name)

		err := row.Scan(&dbSecret.ID, &dbSecret.Name, &dbSecret.UserID, &dbSecret.CreatedAt,
			&dbSecret.UpdatedAt, &dbSecret.Metadata, &dbSecret.Salt, &dbSecret.IV, &dbSecret.FileName,
			&dbSecret.Location, &dbSecret.Checksum, &dbSecret.Size)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return filerepo.ErrSecretNotFound
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

	info, err := file.NewContentInfo(dbSecret.FileName, dbSecret.Location,
		dbSecret.Checksum, dbSecret.Salt, dbSecret.IV, dbSecret.Size)
	if err != nil {
		return nil, fmt.Errorf("failed to create file content info: %w", err)
	}

	secret, err := file.NewSecret(dbSecret.ID, dbSecret.Name, dbSecret.UserID, metadata,
		dbSecret.CreatedAt, dbSecret.UpdatedAt, info)
	if err != nil {
		return nil, fmt.Errorf("failed to create file secret: %w", err)
	}

	return secret, nil
}

// ListSecrets returns a list of file secret entries from the storage.
func (s *Storage) ListSecrets(ctx context.Context, userID string) ([]*file.Secret, error) {
	var dbSecrets []filerepo.Secret

	err := pgutils.WithRetry(func() error {
		query := `SELECT id, name, user_id, created_at, updated_at, metadata, filename, location, checksum, size
			FROM vault_files WHERE user_id = $1`

		rows, err := s.db.QueryContext(ctx, query, userID)
		if err != nil {
			return fmt.Errorf("db.QueryContext: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var dbSecret filerepo.Secret

			err := rows.Scan(&dbSecret.ID, &dbSecret.Name, &dbSecret.UserID, &dbSecret.CreatedAt,
				&dbSecret.UpdatedAt, &dbSecret.Metadata, &dbSecret.FileName, &dbSecret.Location,
				&dbSecret.Checksum, &dbSecret.Size)
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

	secrets := make([]*file.Secret, 0, len(dbSecrets))

	for _, dbSecret := range dbSecrets {
		var metadata map[string]string

		err = json.Unmarshal([]byte(dbSecret.Metadata), &metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		info, err := file.NewContentInfo(dbSecret.FileName, dbSecret.Location,
			dbSecret.Checksum, "", "", dbSecret.Size)
		if err != nil {
			return nil, fmt.Errorf("failed to create file content info: %w", err)
		}

		secret, err := file.NewSecret(dbSecret.ID, dbSecret.Name, dbSecret.UserID, metadata,
			dbSecret.CreatedAt, dbSecret.UpdatedAt, info)
		if err != nil {
			return nil, fmt.Errorf("failed to create file secret: %w", err)
		}

		secrets = append(secrets, secret)
	}

	return secrets, nil
}

// UpdateSecret updates a file secret entry in the storage.
func (s *Storage) UpdateSecret(ctx context.Context, secret *file.Secret) (*file.Secret, error) {
	err := pgutils.WithRetry(func() error {
		query := `UPDATE vault_files
			SET updated_at = $1, metadata = $2, salt = $3, iv = $4, filename = $5, location = $6, checksum = $7, size = $8
			WHERE user_id = $9 AND name = $10`

		_, err := s.db.ExecContext(ctx, query, secret.UpdatedAt(), secret.Metadata(),
			secret.ContentInfo().Salt(), secret.ContentInfo().IV(), secret.ContentInfo().FileName(),
			secret.ContentInfo().Location(), secret.ContentInfo().Checksum(), secret.ContentInfo().Size(),
			secret.UserID(), secret.Name())
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

// DeleteSecret deletes a file secret entry from the storage.
func (s *Storage) DeleteSecret(ctx context.Context, userID, secretName string) error {
	err := pgutils.WithRetry(func() error {
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
	})
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}
