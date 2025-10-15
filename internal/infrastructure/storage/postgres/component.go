package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/D1sordxr/image-processor/internal/domain/app/port"
	"github.com/D1sordxr/image-processor/internal/domain/core/image/options"
	"github.com/D1sordxr/image-processor/internal/infrastructure/config"
	"github.com/wb-go/wbf/dbpg"
)

func NewRunFunc(log port.Logger, db *dbpg.DB, cfg config.Postgres) port.RunFunc {
	return func(ctx context.Context) error {
		const op = "postgres.Run"

		if err := SetupStorage(db.Master, cfg); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		healthTicker := time.NewTicker(options.HealthTickerInterval)
		defer healthTicker.Stop()

		for {
			select {
			case <-healthTicker.C:
				if err := db.Master.Ping(); err != nil {
					log.Error("Database ping failed", "error", err)
					return fmt.Errorf("%s: %w", op, err)
				}
			case <-ctx.Done():
				log.Info("Database component stopping")
				return nil
			}
		}
	}
}

func NewShutdownFunc(db *dbpg.DB) port.ShutdownFunc {
	return func(ctx context.Context) error {
		return db.Master.Close()
	}
}
