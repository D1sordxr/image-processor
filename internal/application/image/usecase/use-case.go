package usecase

import (
	"context"
	"fmt"
	"github.com/D1sordxr/image-processor/internal/application/image/input"
	"github.com/D1sordxr/image-processor/internal/application/image/output"
	appPorts "github.com/D1sordxr/image-processor/internal/domain/app/port"
	"github.com/D1sordxr/image-processor/internal/domain/core/image/model"
	"github.com/D1sordxr/image-processor/internal/domain/core/image/options"
	"github.com/D1sordxr/image-processor/internal/domain/core/image/port"
	"github.com/D1sordxr/image-processor/internal/domain/core/image/vo"
	sharedVO "github.com/D1sordxr/image-processor/internal/domain/core/shared/vo"
	"github.com/D1sordxr/image-processor/pkg/logger"
	"github.com/google/uuid"
	"time"
)

type UseCase struct {
	log       appPorts.Logger
	txManager appPorts.TxManager
	repo      port.Repository
	s3        port.S3Repository
	queue     port.Queue
	baseURL   sharedVO.BaseURL
}

func NewUseCase(
	log appPorts.Logger,
	txManager appPorts.TxManager,
	repo port.Repository,
	s3 port.S3Repository,
	queue port.Queue,
	baseURL sharedVO.BaseURL,

) *UseCase {
	return &UseCase{
		log:       log,
		txManager: txManager,
		repo:      repo,
		s3:        s3,
		queue:     queue,
		baseURL:   baseURL,
	}
}

func (uc *UseCase) UploadImage(ctx context.Context, in input.UploadImageInput) (*output.UploadImageOutput, error) {
	const op = "image.UseCase.UploadImage"
	logFields := logger.WithFields("operation", op)

	uc.log.Info("Uploading new image...", logFields(
		"filename", in.Filename,
	)...)

	imageID := uuid.New()
	filename := vo.NewFilenameOriginal(in.Filename)
	resultURL := vo.NewResultUrl(uc.baseURL, imageID.String())

	fileInfo, err := uc.s3.SaveOriginal(ctx, in.ImageData, imageID.String())
	if err != nil {
		uc.log.Error("Failed to upload image", logFields("error", err)...)
		return nil, fmt.Errorf("%s: failed to save original image: %w", op, err)
	}

	var imageMetadata *model.ImageMetadata
	txErr := uc.txManager.WithTransaction(ctx, nil, func(ctx context.Context) error {
		var innerErr error

		imageMetadata, innerErr = uc.repo.SaveImageMetadata(ctx, options.ImageCreateParams{
			ID:           imageID,
			OriginalName: in.Filename,
			FileName:     filename,
			Status:       vo.StatusProcessing,
			ResultURL:    resultURL,
			Size:         fileInfo.Size,
			Format:       fileInfo.MimeType,
			UploadedAt:   time.Now(),
		})
		if innerErr != nil {
			uc.log.Error("Failed to save image metadata", logFields("error", innerErr)...)
			return fmt.Errorf("save image metadata: %w", innerErr)
		}

		if innerErr = uc.queue.PublishImageTask(ctx, &model.ProcessingImage{
			ImageID: imageID.String(),
			Options: in.Options,
		}); innerErr != nil {
			uc.log.Error("Failed to publish image task", logFields("error", innerErr)...)
			return fmt.Errorf("publish image task: %w", innerErr)
		}

		return nil
	})

	if txErr != nil {
		uc.log.Error("Failed to perform transaction", logFields("error", txErr)...)

		if delErr := uc.s3.DeleteOriginal(ctx, imageID.String()); delErr != nil {
			uc.log.Error("Failed to delete original image after transaction failure",
				logFields("error", delErr, "image_id", imageID.String())...)
		} else {
			uc.log.Info("Successfully cleaned up original image after transaction failure",
				logFields("image_id", imageID.String())...)
		}

		return nil, fmt.Errorf("%s: %w", op, txErr)
	}

	uc.log.Info("Successfully uploaded image", logFields(
		"image_id", imageID.String(),
		"status", imageMetadata.Status.String(),
	)...)

	return &output.UploadImageOutput{
		ImageID:   imageMetadata.ID.String(),
		Status:    imageMetadata.Status.String(),
		ResultURL: imageMetadata.ResultURL.String(),
	}, nil
}

func (uc *UseCase) GetImage(ctx context.Context, in input.GetImageInput) (*output.GetImageOutput, error) {
	//TODO implement me
	panic("implement me")
}

func (uc *UseCase) GetImageStatus(ctx context.Context, in input.GetImageStatusInput) (*output.GetImageStatusOutput, error) {
	//TODO implement me
	panic("implement me")
}

func (uc *UseCase) DeleteImage(ctx context.Context, in input.DeleteImageInput) (*output.DeleteImageOutput, error) {
	//TODO implement me
	panic("implement me")
}

func (uc *UseCase) ProcessImageSync(ctx context.Context, in input.ProcessImageSyncInput) (*output.ProcessImageSyncOutput, error) {
	//TODO implement me
	panic("implement me")
}
