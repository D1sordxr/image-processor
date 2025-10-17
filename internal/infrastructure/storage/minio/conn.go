package minio

import (
	"context"
	"fmt"
	"time"

	"github.com/D1sordxr/image-processor/internal/domain/core/image/options"
	"github.com/D1sordxr/image-processor/internal/infrastructure/ticker"

	"github.com/D1sordxr/image-processor/internal/infrastructure/config"
	"github.com/minio/minio-go/v7"

	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Connection struct {
	cfg          *config.Minio
	Storage      *minio.Client
	healthCancel context.CancelFunc
}

func New(cfg config.Minio) *Connection {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		panic("failed to connect to minio")
	}

	return &Connection{
		cfg:     &cfg,
		Storage: client,
	}
}

func (c *Connection) Run(ctx context.Context) error {
	const op = "minio.Connection.Run"

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	exists, err := c.Storage.BucketExists(ctx, c.cfg.BucketName)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if !exists {
		err = c.Storage.MakeBucket(ctx, c.cfg.BucketName, minio.MakeBucketOptions{
			Region: c.cfg.Region,
		})
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		err = c.Storage.SetBucketPolicy(ctx, c.cfg.BucketName, fmt.Sprintf(policy, c.cfg.BucketName))
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	healthCancel, err := c.Storage.HealthCheck(options.HealthTickerInterval)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	c.healthCancel = healthCancel

	healthTicker := ticker.NewHealth()
	defer healthTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-healthTicker.C:
			if ok := c.Storage.IsOnline(); !ok {
				return fmt.Errorf("%s: storage is not online", op)
			}
		}
	}
}

func (c *Connection) Shutdown(_ context.Context) error {
	c.healthCancel()
	return nil
}
