package model

import (
	"time"

	"github.com/D1sordxr/image-processor/internal/domain/core/image/vo"
	"github.com/google/uuid"
)

type FileInfo struct {
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
	ETag     string `json:"etag"`
}

type ImageMetadata struct {
	ID            uuid.UUID      `json:"id"`
	OriginalName  string         `json:"original_name"`
	Format        string         `json:"format"`
	Size          int64          `json:"size"`
	FileName      vo.Filename    `json:"file_name"`
	Status        vo.Status      `json:"status"` // "uploaded", "processing", "completed", "failed"
	ResultURL     vo.ResultUrl   `json:"result_url,omitempty"`
	UploadedAt    time.Time      `json:"uploaded_at"`
	ProcessedData *ProcessedData `json:"processed_data"`
}

type ProcessedData struct {
	Width         int       `json:"width"`
	Height        int       `json:"height"`
	ProcessedName string    `json:"processed_name"`
	ProcessedAt   time.Time `json:"processed_at"`
}

type ProcessingImage struct {
	ImageID   string            `json:"image_id"`
	Options   ProcessingOptions `json:"options"`
	Timestamp time.Time         `json:"timestamp"`
}
