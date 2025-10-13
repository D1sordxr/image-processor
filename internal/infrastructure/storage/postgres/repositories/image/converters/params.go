package converters

import (
	"database/sql"
	"time"

	"github.com/D1sordxr/image-processor/internal/domain/core/image/options"
	"github.com/D1sordxr/image-processor/internal/domain/core/image/vo"
	"github.com/D1sordxr/image-processor/internal/infrastructure/storage/postgres/repositories/image/gen"
	"github.com/google/uuid"
)

func ToCreateImageParams(params options.ImageCreateParams) gen.CreateImageParams {
	return gen.CreateImageParams{
		ID:           params.ID,
		OriginalName: params.OriginalName,
		FileName:     params.FileName.String(),
		Status:       params.Status.String(),
		ResultUrl:    toNullStringFromResultURL(params.ResultURL),
		Size:         params.Size,
		Format:       params.Format,
		UploadedAt:   params.UploadedAt,
	}
}

func ToUpdateImageParams(id uuid.UUID, params options.ImageUpdateParams) gen.UpdateImageParams {
	return gen.UpdateImageParams{
		ID:        id,
		Status:    toStringFromStatus(params.Status),
		ResultUrl: toNullStringFromResultURL(params.ResultURL),
	}
}

func ToUpdateImageStatusParams(id uuid.UUID, status vo.Status) gen.UpdateImageStatusParams {
	return gen.UpdateImageStatusParams{
		ID:     id,
		Status: status.String(),
	}
}

// ToCreateProcessedImageParams конвертирует параметры создания обработанного изображения
func ToCreateProcessedImageParams(params options.ProcessedImageCreateParams) gen.CreateProcessedImageParams {
	return gen.CreateProcessedImageParams{
		ImageID:       params.ImageID,
		Width:         params.Width,
		Height:        params.Height,
		ProcessedName: params.ProcessedName,
		ProcessedAt:   params.ProcessedAt,
	}
}

// ToUpdateProcessedImageParams конвертирует параметры обновления обработанного изображения
func ToUpdateProcessedImageParams(params options.ProcessedImageCreateParams) gen.UpdateProcessedImageParams {
	return gen.UpdateProcessedImageParams{
		ImageID:       params.ImageID,
		Width:         params.Width,
		Height:        params.Height,
		ProcessedName: params.ProcessedName,
		ProcessedAt:   params.ProcessedAt,
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
		Format:   toNullString(params.Format),
		FromDate: toNullTime(params.FromDate),
		ToDate:   toNullTime(params.ToDate),
		Offset:   toNullInt32(params.Offset),
		Limit:    toNullInt32(params.Limit),
	}
}

// ToCountImagesWithFiltersParams конвертирует параметры для подсчета
func ToCountImagesWithFiltersParams(params options.ImageListParams) gen.CountImagesWithFiltersParams {
	return gen.CountImagesWithFiltersParams{
		Status:   toNullStringFromStatus(params.Status),
		Format:   toNullString(params.Format),
		FromDate: toNullTime(params.FromDate),
		ToDate:   toNullTime(params.ToDate),
	}
}

// ToGetRecentProcessedImagesParams конвертирует параметры для недавно обработанных
func ToGetRecentProcessedImagesParams(params options.RecentProcessedImagesParams) gen.GetRecentProcessedImagesParams {
	return gen.GetRecentProcessedImagesParams{
		ProcessedAt: params.Since,
		Limit:       params.Limit,
	}
}

// Вспомогательные функции

func toNullStringFromStatus(s *vo.Status) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: s.String(), Valid: true}
}

func toStringFromStatus(s *vo.Status) string {
	if s == nil {
		return ""
	}
	return s.String()
}

func toNullStringFromResultURL(url *vo.ResultUrl) sql.NullString {
	if url == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: url.String(), Valid: true}
}

func toNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

func toNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func toNullInt32(i *int32) sql.NullInt32 {
	if i == nil {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: *i, Valid: true}
}

func toNullInt64(i *int64) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *i, Valid: true}
}
