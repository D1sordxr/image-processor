package port

import (
	"context"
	"github.com/D1sordxr/image-processor/internal/domain/core/image/model"
	"github.com/D1sordxr/image-processor/internal/domain/core/image/options"
)

type Repository interface {
	Save(ctx context.Context, params options.ImageCreateParams) (*model.ImageMetadata, error)
	SaveProcessed(ctx context.Context, p options.ProcessedImageCreateParams) error
	UpdateStatus(ctx context.Context, p options.ImageUpdateParams) error
}
