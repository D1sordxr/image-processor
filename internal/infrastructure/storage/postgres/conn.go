package postgres

import (
	"context"
	"fmt"

	"github.com/D1sordxr/image-processor/internal/infrastructure/config"
	"github.com/D1sordxr/image-processor/internal/infrastructure/ticker"
	"github.com/wb-go/wbf/dbpg"
)

type Connection struct {
	cfg     *config.Postgres
	Storage *dbpg.DB
}

func New(cfg config.Postgres) *Connection {
	return &Connection{
		cfg:     &cfg,
		Storage: nil,
	}
}

func (c *Connection) Run(ctx context.Context) error {
	const op = "postgres.Connection.Run"

	db, err := dbpg.New(c.cfg.ConnectionString(), nil, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	c.Storage = db

	if err = SetupStorage(c.Storage.Master, c.cfg); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	healthTicker := ticker.NewHealth()
	defer healthTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-healthTicker.C:
			if err = c.Storage.Master.Ping(); err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
		}
	}

}

func (c *Connection) Shutdown(_ context.Context) error {
	const op = "postgres.Connection.Shutdown"
	if err := c.Storage.Master.Close(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
