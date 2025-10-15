package port

import (
	"context"

	"github.com/D1sordxr/image-processor/internal/application/image/input"
	"github.com/D1sordxr/image-processor/internal/application/image/output"
	"github.com/D1sordxr/image-processor/internal/domain/core/image/model"
)

type UseCase interface {
	Upload(ctx context.Context, in input.UploadImageInput) (*output.UploadImageOutput, error)
	Process(ctx context.Context, image *model.ProcessingImage) error
	Get(ctx context.Context, in input.GetImageInput) (*output.GetImageOutput, error)
	GetStatus(ctx context.Context, in input.GetImageStatusInput) (*output.GetImageStatusOutput, error)
	Delete(ctx context.Context, in input.DeleteImageInput) (*output.DeleteImageOutput, error)
	ProcessSync(ctx context.Context, in input.ProcessImageSyncInput) (*output.ProcessImageSyncOutput, error)
}
