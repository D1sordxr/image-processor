package postgres

import (
	"embed"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS
