// Package sqlitemigrate provides SQLite database migrations.
package sqlitemigrate

import (
	"context"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// Run runs migrations.
func Run(ctx context.Context, dbPath string) error {
	db, err := goose.OpenDBWithDriver(string(goose.DialectSQLite3), dbPath)
	if err != nil {
		return fmt.Errorf("goose.OpenDBWithDriver: %w", err)
	}
	defer db.Close()

	goose.SetBaseFS(embedMigrations)

	if err := goose.UpContext(ctx, db, "migrations"); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}
