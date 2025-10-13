package output

import "github.com/D1sordxr/image-processor/internal/domain/core/image/model"

type UploadImageOutput struct {
	ImageID   string `json:"image_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	ResultURL string `json:"result_url"`
}

type GetImageOutput struct {
	ImageData []byte               `json:"image_data,omitempty"`
	Metadata  *model.ImageMetadata `json:"metadata"`
	ImageURL  string               `json:"image_url"`
}

type GetImageStatusOutput struct {
	Status string `json:"status"`
}

type DeleteImageOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ProcessImageSyncOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
