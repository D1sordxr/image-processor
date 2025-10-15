package repo

import (
	"context"
	"fmt"

	"github.com/D1sordxr/image-processor/internal/domain/core/image/model"
	"github.com/D1sordxr/image-processor/internal/domain/core/image/options"
	"github.com/D1sordxr/image-processor/internal/infrastructure/storage/postgres/executor"
	"github.com/D1sordxr/image-processor/internal/infrastructure/storage/postgres/repositories/image/converters"
	"github.com/D1sordxr/image-processor/internal/infrastructure/storage/postgres/repositories/image/gen"
	"github.com/google/uuid"
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

func (r *Repository) Save(
	ctx context.Context,
	p options.ImageCreateParams,
) (*model.ImageMetadata, error) {
	const op = "image.Repository.Save"

	rawImage, err := r.queries.CreateImage(
		ctx,
		r.executor.GetExecutor(ctx),
		converters.ToCreateImageParams(p),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	image := converters.ToDomainImage(rawImage)
	return &image, nil
}

func (r *Repository) SaveProcessed(
	ctx context.Context,
	p options.ProcessedImageCreateParams,
) error {
	const op = "image.Repository.SaveProcessed"

	if _, err := r.queries.CreateProcessedImage(
		ctx,
		r.executor.GetExecutor(ctx),
		converters.ToCreateProcessedImageParams(p),
	); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *Repository) UpdateStatus(
	ctx context.Context,
	p options.ImageUpdateParams,
) error {
	const op = "image.Repository.UpdateStatus"

	if _, err := r.queries.UpdateImageStatus(
		ctx,
		r.executor.GetExecutor(ctx),
		converters.ToUpdateImageStatusParams(p),
	); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *Repository) Get(
	ctx context.Context,
	imageID uuid.UUID,
) (*model.ImageMetadata, error) {
	const op = "image.Repository.Get"

	metadataRaw, err := r.queries.GetImageByID(
		ctx,
		r.executor.GetExecutor(ctx),
		imageID,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	metadata := converters.ToDomainImage(metadataRaw)
	return &metadata, nil
}

func (r *Repository) GetWithProcessedData(
	ctx context.Context,
	imageID uuid.UUID,
) (*model.ImageMetadata, error) {
	const op = "image.Repository.GetWithProcessedData"

	metadataRaw, err := r.queries.GetImageWithProcessedData(
		ctx,
		r.executor.GetExecutor(ctx),
		imageID,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	metadata := converters.ToDomainImageWithProcessedData(metadataRaw)
	return &metadata, nil
}

func (r *Repository) Delete(
	ctx context.Context,
	imageID uuid.UUID,
) error {
	const op = "image.Repository.Delete"

	if err := r.queries.DeleteImage(
		ctx,
		r.executor.GetExecutor(ctx),
		imageID,
	); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *Repository) DeleteProcessed(
	ctx context.Context,
	imageID uuid.UUID,
) error {
	const op = "image.Repository.DeleteProcessed"

	if err := r.queries.DeleteProcessedImage(
		ctx,
		r.executor.GetExecutor(ctx),
		imageID,
	); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
