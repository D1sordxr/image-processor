package converters

import (
	"database/sql"
	"github.com/D1sordxr/image-processor/internal/domain/core/image/options"
	"github.com/D1sordxr/image-processor/internal/domain/core/image/vo"
	"github.com/D1sordxr/image-processor/internal/infrastructure/storage/postgres/repositories/image/gen"
	"github.com/D1sordxr/image-processor/pkg/sqlutils"
)

func ToCreateImageParams(params options.ImageCreateParams) gen.CreateImageParams {
	resultUrlStr := params.ResultURL.String()
	return gen.CreateImageParams{
		ID:           params.ID,
		OriginalName: params.OriginalName,
		FileName:     params.FileName.String(),
		Status:       params.Status.String(),
		ResultUrl:    sqlutils.ToNullableString(&resultUrlStr),
		Size:         params.Size,
		Format:       params.Format,
		UploadedAt:   params.UploadedAt,
	}
}

func ToUpdateImageStatusParams(params options.ImageUpdateParams) gen.UpdateImageStatusParams {
	return gen.UpdateImageStatusParams{
		ID:     params.ImageID,
		Status: params.Status.String(),
	}
}

// ToCreateProcessedImageParams конвертирует параметры создания обработанного изображения
func ToCreateProcessedImageParams(params options.ProcessedImageCreateParams) gen.CreateProcessedImageParams {
	return gen.CreateProcessedImageParams{
		ImageID:     params.ImageID,
		Width:       int32(params.Width),
		Height:      int32(params.Height),
		ProcessedAt: params.ProcessedAt,
	}
}

// ToListImagesParams конвертирует параметры пагинации
func ToListImagesParams(params options.PaginationParams) gen.ListImagesParams {
	return gen.ListImagesParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	}
}

// ToListImagesByStatusParams конвертирует параметры списка по статусу
func ToListImagesByStatusParams(status vo.Status, pagination options.PaginationParams) gen.ListImagesByStatusParams {
	return gen.ListImagesByStatusParams{
		Status: status.String(),
		Limit:  pagination.Limit,
		Offset: pagination.Offset,
	}
}

// ToListImagesWithFiltersParams конвертирует параметры фильтрации
func ToListImagesWithFiltersParams(params options.ImageListParams) gen.ListImagesWithFiltersParams {
	return gen.ListImagesWithFiltersParams{
		Status:   toNullStringFromStatus(params.Status),
		Format:   sqlutils.ToNullableString(params.Format),
		FromDate: sqlutils.ToNullableTime(params.FromDate),
		ToDate:   sqlutils.ToNullableTime(params.ToDate),
		Offset:   sqlutils.ToNullableInt32(params.Offset),
		Limit:    sqlutils.ToNullableInt32(params.Limit),
	}
}

// ToGetRecentProcessedImagesParams конвертирует параметры для недавно обработанных
func ToGetRecentProcessedImagesParams(params options.RecentProcessedImagesParams) gen.GetRecentProcessedImagesParams {
	return gen.GetRecentProcessedImagesParams{
		ProcessedAt: params.Since,
		Limit:       params.Limit,
	}
}

func toNullStringFromStatus(s *vo.Status) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: s.String(), Valid: true}
}
