package s3repo

import (
	"bytes"
	"context"
	"fmt"
	"github.com/D1sordxr/image-processor/internal/domain/core/image/model"
	"github.com/D1sordxr/image-processor/internal/domain/core/image/options"
	"github.com/D1sordxr/image-processor/internal/domain/core/image/vo"
	minioRoot "github.com/D1sordxr/image-processor/internal/infrastructure/storage/minio"
	"github.com/minio/minio-go/v7"
	"io"
	"mime"
	"path/filepath"
)

type S3Repository struct {
	storage    *minioRoot.Storage
	bucketName string
}

func New(storage *minioRoot.Storage, bucketName string) *S3Repository {
	return &S3Repository{
		storage:    storage,
		bucketName: bucketName,
	}
}

func (s3 *S3Repository) Save(ctx context.Context, data []byte, filename string) (*model.FileInfo, error) {
	ext := filepath.Ext(filename)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	reader := bytes.NewReader(data)
	info, err := s3.storage.Client.PutObject(ctx, s3.bucketName, filename, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: mimeType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload to MinIO: %w", err)
	}

	return &model.FileInfo{
		Path:     filename,
		Size:     info.Size,
		MimeType: mimeType,
		ETag:     info.ETag,
	}, nil
}

func (s3 *S3Repository) SaveOriginal(ctx context.Context, data []byte, filename string) (*model.FileInfo, error) {
	ext := filepath.Ext(filename)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	reader := bytes.NewReader(data)
	info, err := s3.storage.Client.PutObject(
		ctx,
		s3.bucketName,
		vo.NewFilenameOriginal(filename).String(),
		reader,
		int64(len(data)),
		minio.PutObjectOptions{
			ContentType: mimeType,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to MinIO: %w", err)
	}

	return &model.FileInfo{
		Path:     filename,
		Size:     info.Size,
		MimeType: mimeType,
		ETag:     info.ETag,
	}, nil
}

func (s3 *S3Repository) Get(ctx context.Context, filename string) ([]byte, error) {
	object, err := s3.storage.Client.GetObject(ctx, s3.bucketName, filename, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	defer func() { _ = object.Close() }()

	data, err := io.ReadAll(object)
	if err != nil {
		return nil, fmt.Errorf("failed to read object data: %w", err)
	}

	return data, nil
}

func (s3 *S3Repository) GetOriginal(ctx context.Context, filename string) ([]byte, error) {
	object, err := s3.storage.Client.GetObject(
		ctx,
		s3.bucketName,
		vo.NewFilenameOriginal(filename).String(),
		minio.GetObjectOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	defer func() { _ = object.Close() }()

	data, err := io.ReadAll(object)
	if err != nil {
		return nil, fmt.Errorf("failed to read object data: %w", err)
	}

	return data, nil
}

func (s3 *S3Repository) Delete(ctx context.Context, filename string) error {
	err := s3.storage.Client.RemoveObject(ctx, s3.bucketName, filename, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}

func (s3 *S3Repository) DeleteOriginal(ctx context.Context, filename string) error {
	if err := s3.storage.Client.RemoveObject(
		ctx,
		s3.bucketName,
		vo.NewFilenameOriginal(filename).String(),
		minio.RemoveObjectOptions{},
	); err != nil {
		return fmt.Errorf("failed to delete original object: %w", err)
	}
	return nil
}

func (s3 *S3Repository) Exists(ctx context.Context, filename string) (bool, error) {
	_, err := s3.storage.Client.StatObject(ctx, s3.bucketName, filename, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check object existence: %w", err)
	}
	return true, nil
}

func (s3 *S3Repository) GetURL(ctx context.Context, filename string) (string, error) {
	url, err := s3.storage.Client.PresignedGetObject(ctx, s3.bucketName, filename, options.S3PresignedDuration, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return url.String(), nil
}

func (s3 *S3Repository) GetFileInfo(ctx context.Context, filename string) (*model.FileInfo, error) {
	info, err := s3.storage.Client.StatObject(ctx, s3.bucketName, filename, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object info: %w", err)
	}

	mimeType := info.ContentType
	if mimeType == "" {
		ext := filepath.Ext(filename)
		mimeType = mime.TypeByExtension(ext)
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
	}

	return &model.FileInfo{
		Path:     filename,
		Size:     info.Size,
		MimeType: mimeType,
		ETag:     info.ETag,
	}, nil
}

func (s3 *S3Repository) ListFiles(ctx context.Context, prefix string) ([]model.FileInfo, error) {
	var files []model.FileInfo

	objectCh := s3.storage.Client.ListObjects(ctx, s3.bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", object.Err)
		}

		files = append(files, model.FileInfo{
			Path: object.Key,
			Size: object.Size,
			ETag: object.ETag,
		})
	}

	return files, nil
}

func (s3 *S3Repository) CreateFolder(ctx context.Context, path string) error {
	folderPath := path + "/"
	_, err := s3.storage.Client.PutObject(ctx, s3.bucketName, folderPath, bytes.NewReader([]byte{}), 0, minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to create folder: %w", err)
	}
	return nil
}
