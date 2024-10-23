// Package cardpg provides bank cards PostgreSQL storage implementation.
package cardpg

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pressly/goose/v3"

	"github.com/andymarkow/gophkeeper/internal/domain/vault/bankcard"
	"github.com/andymarkow/gophkeeper/internal/pgutils"
	"github.com/andymarkow/gophkeeper/internal/storage/cardrepo"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// Storage implements bank card storage.
type Storage struct {
	db *sql.DB
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

	return &Storage{
		db: db,
	}, nil
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

// Bootstrap runs migrations.
func (s *Storage) Bootstrap(ctx context.Context) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect(string(goose.DialectPostgres)); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	if err := goose.UpContext(ctx, s.db, "migrations"); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
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

	err = pgutils.WithRetry(func() error {
		query := `INSERT INTO bankcards
			(id, name, user_id, created_at, updated_at, metadata, data)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`

		if _, err := s.db.ExecContext(ctx, query,
			secret.ID(), secret.Name(), secret.UserID(), secret.CreatedAt(), secret.UpdatedAt(), metadata, data); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
				return cardrepo.ErrSecretAlreadyExists
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

// GetSecret returns a bank card secret entry from the storage.
func (s *Storage) GetSecret(ctx context.Context, userID, secretName string) (*bankcard.Secret, error) {
	var dbSecret cardrepo.Secret

	err := pgutils.WithRetry(func() error {
		query := `SELECT id, name, user_id, created_at, updated_at, metadata, data
			FROM bankcards WHERE user_id = $1 AND name = $2`

		row := s.db.QueryRowContext(ctx, query, userID, secretName)

		err := row.Scan(&dbSecret.ID, &dbSecret.Name, &dbSecret.UserID, &dbSecret.CreatedAt,
			&dbSecret.UpdatedAt, &dbSecret.Metadata, &dbSecret.Data)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return cardrepo.ErrSecretNotFound
			}

			return fmt.Errorf("db.QueryRowContext: %w", err)
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

	data, err := bankcard.UnmarshalData([]byte(dbSecret.Data))
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	secret, err := bankcard.NewSecret(dbSecret.ID, dbSecret.Name, dbSecret.UserID, metadata,
		dbSecret.CreatedAt, dbSecret.UpdatedAt, data)
	if err != nil {
		return nil, fmt.Errorf("failed to create bank card secret: %w", err)
	}

	return secret, nil
}

// ListSecrets returns a list of bank card secret entries from the storage.
func (s *Storage) ListSecrets(ctx context.Context, userID string) ([]*bankcard.Secret, error) {
	var dbSecrets []cardrepo.Secret

	err := pgutils.WithRetry(func() error {
		query := `SELECT id, name, user_id, created_at, updated_at, metadata
			FROM bankcards WHERE user_id = $1`

		rows, err := s.db.QueryContext(ctx, query, userID)
		if err != nil {
			return fmt.Errorf("db.QueryContext: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var dbSecret cardrepo.Secret

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

	secrets := make([]*bankcard.Secret, 0, len(dbSecrets))

	for _, dbSecret := range dbSecrets {
		var metadata map[string]string

		err = json.Unmarshal([]byte(dbSecret.Metadata), &metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		secret, err := bankcard.NewSecret(dbSecret.ID, dbSecret.Name, dbSecret.UserID, metadata,
			dbSecret.CreatedAt, dbSecret.UpdatedAt, bankcard.NewEmptyData())
		if err != nil {
			return nil, fmt.Errorf("failed to create bank card secret: %w", err)
		}

		secrets = append(secrets, secret)
	}

	return secrets, nil
}

// UpdateSecret updates a bank card secret entry in the storage.
func (s *Storage) UpdateSecret(ctx context.Context, secret *bankcard.Secret) (*bankcard.Secret, error) {
	metadata, err := secret.MetadataJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	data, err := secret.DataJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	err = pgutils.WithRetry(func() error {
		query := `UPDATE bankcards
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

// DeleteSecret deletes a bank card secret entry from the storage.
func (s *Storage) DeleteSecret(ctx context.Context, userID, secretName string) error {
	err := pgutils.WithRetry(func() error {
		query := `DELETE FROM bankcards WHERE user_id = $1 AND name = $2`

		result, err := s.db.ExecContext(ctx, query, userID, secretName)
		if err != nil {
			return fmt.Errorf("db.ExecContext: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("result.RowsAffected: %w", err)
		}

		if rowsAffected == 0 {
			return cardrepo.ErrSecretNotFound
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}
