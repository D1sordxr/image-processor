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
	return &Connection{
		cfg:     &cfg,
		Storage: nil,
	}
}

func (c *Connection) Run(ctx context.Context) error {
	const op = "minio.Connection.Run"

	client, err := minio.New(c.cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(c.cfg.AccessKey, c.cfg.SecretKey, ""),
		Secure: c.cfg.UseSSL,
		Region: c.cfg.Region,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	c.Storage = client

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	exists, err := client.BucketExists(ctx, c.cfg.BucketName)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if !exists {
		err = client.MakeBucket(ctx, c.cfg.BucketName, minio.MakeBucketOptions{
			Region: c.cfg.Region,
		})
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		err = client.SetBucketPolicy(ctx, c.cfg.BucketName, fmt.Sprintf(policy, c.cfg.BucketName))
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
