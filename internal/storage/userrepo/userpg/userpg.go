// Package userpg provides PostgreSQL storage implementation for user model.
package userpg

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pressly/goose/v3"

	// Postgres driver.
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/andymarkow/gophkeeper/internal/domain/user"
	"github.com/andymarkow/gophkeeper/internal/pgutils"
	"github.com/andymarkow/gophkeeper/internal/storage/userrepo"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// Storage implements storage.
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

// AddUser adds a new user to the storage.
func (s *Storage) AddUser(ctx context.Context, usr *user.User) error {
	err := pgutils.WithRetry(func() error {
		if _, err := s.db.ExecContext(ctx,
			`INSERT INTO users (id, login, password) VALUES ($1, $2, $3)`,
			usr.ID(), usr.Login(), usr.Password(),
		); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
				return userrepo.ErrUsrAlreadyExists
			}

			return fmt.Errorf("db.ExecContext: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

// GetUser returns a user from the storage.
func (s *Storage) GetUser(ctx context.Context, login string) (*user.User, error) {
	var usrID, usrLogin, usrPassword string

	err := pgutils.WithRetry(func() error {
		if err := s.db.QueryRowContext(ctx,
			`SELECT id, login, password FROM users WHERE login = $1`,
			login,
		).Scan(&usrID, &usrLogin, &usrPassword); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return userrepo.ErrUsrNotFound
			}

			return fmt.Errorf("db.QueryRowContext: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	usr, err := user.NewUser(usrID, usrLogin, usrPassword)
	if err != nil {
		return nil, fmt.Errorf("user.New: %w", err)
	}

	return usr, nil
}
