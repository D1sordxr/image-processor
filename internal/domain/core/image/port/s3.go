package port

import (
	"context"
	"github.com/D1sordxr/image-processor/internal/domain/core/image/model"
)

type S3Repository interface {
	Save(ctx context.Context, data []byte, filename string) (*model.FileInfo, error)
	SaveOriginal(ctx context.Context, data []byte, filename string) (*model.FileInfo, error)
	Get(ctx context.Context, filename string) ([]byte, error)
	Delete(ctx context.Context, filename string) error
	DeleteOriginal(ctx context.Context, filename string) error
	Exists(ctx context.Context, filename string) (bool, error)
	GetURL(ctx context.Context, filename string) (string, error)
	GetFileInfo(ctx context.Context, filename string) (*model.FileInfo, error)
	ListFiles(ctx context.Context, prefix string) ([]model.FileInfo, error)
	CreateFolder(ctx context.Context, path string) error
}
