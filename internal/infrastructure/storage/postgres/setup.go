package postgres

import (
	"database/sql"

	"github.com/D1sordxr/image-processor/internal/infrastructure/config"

	"github.com/pressly/goose/v3"
)

func SetupStorage(db *sql.DB, cfg config.Postgres) error {
	if cfg.Migrations {
		goose.SetBaseFS(embedMigrations)
		if err := goose.SetDialect("postgres"); err != nil {
			return err
		}
		if err := goose.Up(db, "migrations"); err != nil {
			return err
		}
	}
	return nil
}
