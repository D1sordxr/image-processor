package dto

import (
	"github.com/D1sordxr/image-processor/internal/domain/core/image/model"
	"github.com/D1sordxr/image-processor/internal/domain/core/shared/validator"
)

type UploadRequest struct {
	Width         int    `form:"width" validate:"min=0"`
	Height        int    `form:"height" validate:"min=0"`
	Quality       int    `form:"quality" validate:"min=1,max=100"`
	Format        string `form:"format" validate:"oneof=jpeg jpg png gif"`
	WatermarkText string `form:"watermark"`
	Thumbnail     bool   `form:"thumbnail"`
}

func (r *UploadRequest) Validate() error {
	return validator.ValidateStruct(r)
}

func (r *UploadRequest) ToProcessingOptions() model.ProcessingOptions {
	return model.ProcessingOptions{
		Width:         r.Width,
		Height:        r.Height,
		Quality:       r.Quality,
		Format:        r.Format,
		WatermarkText: r.WatermarkText,
		Thumbnail:     r.Thumbnail,
	}
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

type SuccessResponse struct {
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

type UploadResponse struct {
	ImageID           string                  `json:"image_id"`
	ResultURL         string                  `json:"result_url"`
	ProcessingOptions model.ProcessingOptions `json:"processing_options"`
	Message           string                  `json:"message"`
}

type ProcessingStatusResponse struct {
	Status   string `json:"status"`
	ImageID  string `json:"image_id"`
	ImageURL string `json:"image_url"`
	Message  string `json:"message"`
}

type HealthCheckResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Service   string `json:"service"`
}
