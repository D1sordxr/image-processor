package repo

import (
	"context"
	"fmt"
	"github.com/D1sordxr/image-processor/internal/domain/core/image/model"
	"github.com/D1sordxr/image-processor/internal/domain/core/image/options"
	"github.com/D1sordxr/image-processor/internal/infrastructure/storage/postgres/executor"
	"github.com/D1sordxr/image-processor/internal/infrastructure/storage/postgres/repositories/image/converters"
	"github.com/D1sordxr/image-processor/internal/infrastructure/storage/postgres/repositories/image/gen"
)

type Repository struct {
	executor *executor.Executor
	queries  *gen.Queries
}

func New(executor *executor.Executor) *Repository {
	return &Repository{
		executor: executor,
		queries:  gen.New(),
	}
}

func (r *Repository) SaveImageMetadata(
	ctx context.Context,
	p options.ImageCreateParams,
) (*model.ImageMetadata, error) {
	const op = "image.Repository.SaveImageMetadata"

	rawImage, err := r.queries.CreateImage(
		ctx,
		r.executor.GetExecutor(ctx),
		converters.ToCreateImageParams(p),
	)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	image := converters.ToDomainImage(rawImage)
	return &image, nil
}
