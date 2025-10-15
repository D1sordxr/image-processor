package port

import (
	"context"

	"github.com/D1sordxr/image-processor/internal/domain/core/image/model"
	"github.com/D1sordxr/image-processor/internal/domain/core/image/options"
	"github.com/google/uuid"
)

type Repository interface {
	Save(ctx context.Context, params options.ImageCreateParams) (*model.ImageMetadata, error)
	SaveProcessed(ctx context.Context, p options.ProcessedImageCreateParams) error
	UpdateStatus(ctx context.Context, p options.ImageUpdateParams) error
	Get(ctx context.Context, imageID uuid.UUID) (*model.ImageMetadata, error)
	GetWithProcessedData(ctx context.Context, imageID uuid.UUID) (*model.ImageMetadata, error)
	Delete(ctx context.Context, imageID uuid.UUID) error
	DeleteProcessed(ctx context.Context, imageID uuid.UUID) error
}
