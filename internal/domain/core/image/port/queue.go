package port

import (
	"context"

	"github.com/D1sordxr/image-processor/internal/domain/core/image/model"
)

type Queue interface {
	PublishImageTask(ctx context.Context, task *model.ProcessingImage) error
}

type Consumer interface {
	StartProcessing(
		ctx context.Context,
		processor func(context.Context, *model.ProcessingImage) error,
	) error
}
