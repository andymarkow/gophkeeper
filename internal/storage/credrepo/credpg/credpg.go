// Package credpg provides PostgreSQL storage implementation for credentials.
package credpg

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

	"github.com/andymarkow/gophkeeper/internal/domain/vault/credential"
	"github.com/andymarkow/gophkeeper/internal/pgutils"
	"github.com/andymarkow/gophkeeper/internal/storage/credrepo"
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

	err = pgutils.WithRetry(func() error {
		query := `INSERT INTO vault_credentials
			(id, name, user_id, created_at, updated_at, metadata, data)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`

		if _, err := s.db.ExecContext(ctx, query,
			secret.ID(), secret.Name(), secret.UserID(), secret.CreatedAt(), secret.UpdatedAt(), metadata, data); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
				return credrepo.ErrSecretAlreadyExists
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

// GetSecret returns a credential secret entry from the storage.
func (s *Storage) GetSecret(ctx context.Context, userID, name string) (*credential.Secret, error) {
	var dbSecret credrepo.Secret

	err := pgutils.WithRetry(func() error {
		query := `SELECT id, name, user_id, created_at, updated_at, metadata, data
			FROM vault_credentials WHERE user_id = $1 AND name = $2`

		row := s.db.QueryRowContext(ctx, query, userID, name)

		err := row.Scan(&dbSecret.ID, &dbSecret.Name, &dbSecret.UserID, &dbSecret.CreatedAt,
			&dbSecret.UpdatedAt, &dbSecret.Metadata, &dbSecret.Data)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return credrepo.ErrSecretNotFound
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

	data, err := credential.UnmarshalData([]byte(dbSecret.Data))
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	secret, err := credential.NewSecret(dbSecret.ID, dbSecret.Name, dbSecret.UserID, metadata,
		dbSecret.CreatedAt, dbSecret.UpdatedAt, data)
	if err != nil {
		return nil, fmt.Errorf("failed to create bank card secret: %w", err)
	}

	return secret, nil
}

// ListSecrets returns a list of credential secret entries from the storage.
func (s *Storage) ListSecrets(ctx context.Context, userID string) ([]*credential.Secret, error) {
	var dbSecrets []credrepo.Secret

	err := pgutils.WithRetry(func() error {
		query := `SELECT id, name, user_id, created_at, updated_at, metadata
			FROM vault_credentials WHERE user_id = $1`

		rows, err := s.db.QueryContext(ctx, query, userID)
		if err != nil {
			return fmt.Errorf("db.QueryContext: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var dbSecret credrepo.Secret

			if err := rows.Scan(&dbSecret.ID, &dbSecret.Name, &dbSecret.UserID, &dbSecret.CreatedAt,
				&dbSecret.UpdatedAt, &dbSecret.Metadata); err != nil {
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

	secrets := make([]*credential.Secret, 0, len(dbSecrets))

	for _, dbSecret := range dbSecrets {
		var metadata map[string]string

		err = json.Unmarshal([]byte(dbSecret.Metadata), &metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		secret, err := credential.NewSecret(dbSecret.ID, dbSecret.Name, dbSecret.UserID, metadata,
			dbSecret.CreatedAt, dbSecret.UpdatedAt, credential.NewEmptyData())
		if err != nil {
			return nil, fmt.Errorf("failed to create credential secret: %w", err)
		}

		secrets = append(secrets, secret)
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

	err = pgutils.WithRetry(func() error {
		query := `UPDATE vault_credentials
			SET metadata = $1, data = $2, updated_at = $3
			WHERE user_id = $4 AND name = $5`

		_, err := s.db.ExecContext(ctx, query, metadata, data, secret.UpdatedAt(), secret.UserID(), secret.Name())
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

// DeleteSecret deletes a credential secret entry from the storage.
func (s *Storage) DeleteSecret(ctx context.Context, userID, secretName string) error {
	err := pgutils.WithRetry(func() error {
		query := `DELETE FROM vault_credentials WHERE user_id = $1 AND name = $2`

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
	})
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}
