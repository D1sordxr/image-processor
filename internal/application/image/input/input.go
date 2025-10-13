package input

import (
	"github.com/D1sordxr/image-processor/internal/domain/core/image/model"
)

type UploadImageInput struct {
	ImageData []byte
	Filename  string
	Options   model.ProcessingOptions
	// CallbackURL string
}

type GetImageInput struct {
	ImageID string
}

type GetImageStatusInput struct {
	ImageID string
}

type DeleteImageInput struct {
	ImageID string
}

type ProcessImageSyncInput struct {
	ImageID string
}
