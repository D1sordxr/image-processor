package port

import (
	"context"
	"github.com/D1sordxr/image-processor/internal/domain/core/image/model"
	"github.com/D1sordxr/image-processor/internal/domain/core/image/options"
)

type Repository interface {
	SaveImageMetadata(ctx context.Context, params options.ImageCreateParams) (*model.ImageMetadata, error)
}
