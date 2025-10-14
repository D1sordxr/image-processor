package options

import (
	"time"

	"github.com/D1sordxr/image-processor/internal/domain/core/image/vo"
	"github.com/google/uuid"
)

type ImageListParams struct {
	Status   *vo.Status
	Format   *string
	FromDate *time.Time
	ToDate   *time.Time
	Limit    *int32
	Offset   *int32
}

type ImageUpdateParams struct {
	ImageID uuid.UUID
	Status  vo.Status
}

type ProcessedImageCreateParams struct {
	ImageID       uuid.UUID
	Width         int
	Height        int
	ProcessedName string
	ProcessedAt   time.Time
}

type ImageSearchParams struct {
	OriginalName *string
	FileName     *vo.Filename
	MinSize      *int64
	MaxSize      *int64
}

type PaginationParams struct {
	Limit  int32
	Offset int32
}

type ImageCreateParams struct {
	ID           uuid.UUID
	OriginalName string
	FileName     vo.Filename
	Status       vo.Status
	ResultURL    vo.ResultUrl
	Size         int64
	Format       string
	UploadedAt   time.Time
}

type RecentProcessedImagesParams struct {
	Since time.Time
	Limit int32
}
