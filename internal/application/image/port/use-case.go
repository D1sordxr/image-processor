package port

import (
	"context"
	"github.com/D1sordxr/image-processor/internal/application/image/input"
	"github.com/D1sordxr/image-processor/internal/application/image/output"
	"github.com/D1sordxr/image-processor/internal/domain/core/image/model"
)

type UseCase interface {
	UploadImage(ctx context.Context, in input.UploadImageInput) (*output.UploadImageOutput, error)
	ProcessImage(ctx context.Context, image *model.ProcessingImage) error
	GetImage(ctx context.Context, in input.GetImageInput) (*output.GetImageOutput, error)
	GetImageStatus(ctx context.Context, in input.GetImageStatusInput) (*output.GetImageStatusOutput, error)
	DeleteImage(ctx context.Context, in input.DeleteImageInput) (*output.DeleteImageOutput, error)
	ProcessImageSync(ctx context.Context, in input.ProcessImageSyncInput) (*output.ProcessImageSyncOutput, error)
}
