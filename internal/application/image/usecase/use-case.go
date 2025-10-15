package usecase

import (
	"context"
	"fmt"
	"time"

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
)

type UseCase struct {
	log       appPorts.Logger
	txManager appPorts.TxManager
	repo      port.Repository
	s3        port.S3Repository
	queue     port.Queue
	processor port.ImageProcessor
	baseURL   sharedVO.BaseURL
}

func New(
	log appPorts.Logger,
	txManager appPorts.TxManager,
	repo port.Repository,
	s3 port.S3Repository,
	queue port.Queue,
	processor port.ImageProcessor,
	baseURL sharedVO.BaseURL,

) *UseCase {
	return &UseCase{
		log:       log,
		txManager: txManager,
		repo:      repo,
		s3:        s3,
		queue:     queue,
		processor: processor,
		baseURL:   baseURL,
	}
}

func (uc *UseCase) Upload(ctx context.Context, in input.UploadImageInput) (*output.UploadImageOutput, error) {
	const op = "image.UseCase.Upload"
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
		imageMetadata, innerErr = uc.repo.Save(ctx, options.ImageCreateParams{
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
			ImageID:   imageID.String(),
			Options:   in.Options,
			Timestamp: time.Now(),
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

func (uc *UseCase) Process(ctx context.Context, image *model.ProcessingImage) error {
	const op = "image.UseCase.Process"
	logFields := logger.WithFields("operation", op, "image_id", image.ImageID)

	uc.log.Info("Processing image", logFields()...)

	imageUUID, err := uuid.Parse(image.ImageID)
	if err != nil {
		uc.log.Error("Failed to parse image UUID", logFields("error", err)...)
		return fmt.Errorf("%s: %w", op, err)
	}

	data, err := uc.s3.GetOriginal(ctx, imageUUID.String())
	if err != nil {
		uc.log.Error("Failed to get original image from S3", logFields("error", err)...)
		return fmt.Errorf("%s: get original: %w", op, err)
	}

	result, err := uc.processor.ProcessImage(data, image.Options)
	if err != nil {
		uc.log.Error("Failed to process image", logFields("error", err)...)
		return fmt.Errorf("%s: process image: %w", op, err)
	}

	if _, err = uc.s3.Save(ctx, result.ProcessedData, imageUUID.String()); err != nil {
		uc.log.Error("Failed to save processed image", logFields("error", err)...)
		return fmt.Errorf("%s: save processed: %w", op, err)
	}

	var txErr error
	defer func() {
		if txErr != nil {
			if innerErr := uc.repo.UpdateStatus(ctx, options.ImageUpdateParams{
				ImageID: imageUUID,
				Status:  vo.StatusFailed,
			}); innerErr != nil {
				uc.log.Error("Failed to update image status to failed", logFields("error", innerErr)...)
			}
		}
	}()

	txErr = uc.txManager.WithTransaction(ctx, nil, func(ctx context.Context) error {
		if err = uc.repo.UpdateStatus(ctx, options.ImageUpdateParams{
			ImageID: imageUUID,
			Status:  vo.StatusCompleted,
		}); err != nil {
			return fmt.Errorf("update status: %w", err)
		}

		if err = uc.repo.SaveProcessed(ctx, options.ProcessedImageCreateParams{
			ImageID:     imageUUID,
			Width:       result.Width,
			Height:      result.Height,
			ProcessedAt: time.Now(),
		}); err != nil {
			return fmt.Errorf("save processed data: %w", err)
		}

		return nil
	})

	if txErr != nil {
		uc.log.Error("Failed to update image metadata", logFields("error", txErr)...)
		return fmt.Errorf("%s: update metadata: %w", op, txErr)
	}

	uc.log.Info("Successfully processed image", logFields()...)
	return nil
}

func (uc *UseCase) Get(ctx context.Context, in input.GetImageInput) (*output.GetImageOutput, error) {
	const op = "image.UseCase.Get"
	logFields := logger.WithFields("operation", op, "image_id", in.ImageID)

	uc.log.Info("Attempting to get image", logFields()...)

	imageID, err := uuid.Parse(in.ImageID)
	if err != nil {
		uc.log.Error("Failed to parse image UUID", logFields("error", err)...)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	metadata, err := uc.repo.GetWithProcessedData(ctx, imageID)
	if err != nil {
		uc.log.Error("Image not found", logFields("error", err)...)
		return nil, fmt.Errorf("%s: image not found: %w", op, err)
	}

	if metadata.Status != vo.StatusCompleted {
		uc.log.Info("Image not yet processed", logFields("status", metadata.Status)...)
		return &output.GetImageOutput{ImageData: nil, Metadata: metadata}, nil
	}

	processedData, err := uc.s3.Get(ctx, imageID.String())
	if err != nil {
		uc.log.Error("Failed to get processed image from S3", logFields("error", err)...)
		return nil, fmt.Errorf("%s: get processed image: %w", op, err)
	}

	uc.log.Info("Successfully retrieved processed image", logFields()...)
	return &output.GetImageOutput{ImageData: processedData, Metadata: metadata}, nil
}

func (uc *UseCase) GetStatus(ctx context.Context, in input.GetImageStatusInput) (*output.GetImageStatusOutput, error) {
	const op = "image.UseCase.GetStatus"
	logFields := logger.WithFields("operation", op, "image_id", in.ImageID)

	uc.log.Info("Attempting to get image status", logFields()...)
	imageID, err := uuid.Parse(in.ImageID)
	if err != nil {
		uc.log.Error("Failed to parse image UUID", logFields("error", err)...)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	metadata, err := uc.repo.Get(ctx, imageID)
	if err != nil {
		uc.log.Error("Failed to get image metadata", logFields("error", err)...)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	uc.log.Info("Successfully get image status", logFields()...)

	return &output.GetImageStatusOutput{Status: metadata.Status.String()}, nil
}

func (uc *UseCase) Delete(ctx context.Context, in input.DeleteImageInput) (*output.DeleteImageOutput, error) {
	const op = "image.UseCase.Delete"
	logFields := logger.WithFields("operation", op, "image_id", in.ImageID)

	uc.log.Info("Attempting to delete image", logFields()...)

	imageID, err := uuid.Parse(in.ImageID)
	if err != nil {
		uc.log.Error("Failed to parse image UUID", logFields("error", err)...)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	image, err := uc.repo.Get(ctx, imageID)
	if err != nil {
		uc.log.Error("Image not found", logFields("error", err)...)
		return nil, fmt.Errorf("%s: image not found: %w", op, err)
	}

	if err = uc.s3.Delete(ctx, imageID.String()); err != nil {
		uc.log.Error("Failed to delete image from S3", logFields("error", err)...)
		return nil, fmt.Errorf("%s: delete from s3: %w", op, err)
	}

	var txErr error
	defer func() {
		if txErr != nil {
			if innerErr := uc.repo.UpdateStatus(ctx, options.ImageUpdateParams{
				ImageID: image.ID,
				Status:  vo.StatusFailed,
			}); innerErr != nil {
				uc.log.Error("Failed to update image status to failed", logFields("error", innerErr)...)
			}
		}
	}()

	if txErr = uc.txManager.WithTransaction(ctx, nil, func(ctx context.Context) error {
		if err = uc.repo.Delete(ctx, image.ID); err != nil {
			return fmt.Errorf("delete image metadata: %w", err)
		}

		if image.Status == vo.StatusCompleted {
			if err = uc.repo.DeleteProcessed(ctx, image.ID); err != nil {
				return fmt.Errorf("delete processed data: %w", err)
			}
		}

		return nil
	}); txErr != nil {
		uc.log.Error("Failed to delete image metadata", logFields("error", txErr)...)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	uc.log.Info("Successfully deleted image", logFields()...)

	return &output.DeleteImageOutput{
		Success: true,
		Message: "Successfully deleted image",
	}, nil
}

func (uc *UseCase) ProcessSync(_ context.Context, _ input.ProcessImageSyncInput) (*output.ProcessImageSyncOutput, error) {
	const op = "image.UseCase.ProcessSync"

	return nil, fmt.Errorf("%s: not implemented", op)
}
