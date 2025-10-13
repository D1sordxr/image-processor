package dto

import (
	"github.com/D1sordxr/image-processor/internal/domain/core/image/model"
)

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
