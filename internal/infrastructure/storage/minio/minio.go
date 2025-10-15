package minio

import (
	"context"
	"fmt"
	"time"

	"github.com/D1sordxr/image-processor/internal/infrastructure/config"
	"github.com/minio/minio-go/v7"

	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Storage struct {
	Client     *minio.Client
	bucketName string
	region     string
}

func New(cfg config.Minio) (*Storage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	exists, err := client.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{
			Region: cfg.Region,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}

		err = client.SetBucketPolicy(ctx, cfg.BucketName, fmt.Sprintf(policy, cfg.BucketName))
		if err != nil {
			return nil, fmt.Errorf("failed to set bucket policy: %w", err)
		}
	}

	return &Storage{
		Client:     client,
		bucketName: cfg.BucketName,
		region:     cfg.Region,
	}, nil
}
