package converters

import (
	"github.com/D1sordxr/image-processor/internal/domain/core/image/model"
	"github.com/D1sordxr/image-processor/internal/domain/core/image/vo"
	"github.com/D1sordxr/image-processor/internal/infrastructure/storage/postgres/repositories/image/gen"
)

func ToDomainImage(dbImage gen.Image) model.ImageMetadata {
	filename := vo.Filename(dbImage.FileName)
	status := vo.NewStatus(dbImage.Status)
	resultURL := vo.ResultUrl(dbImage.ResultUrl.String)
	return model.ImageMetadata{
		ID:            dbImage.ID,
		OriginalName:  dbImage.OriginalName,
		Format:        dbImage.Format,
		Size:          dbImage.Size,
		FileName:      filename,
		Status:        status,
		ResultURL:     resultURL,
		UploadedAt:    dbImage.UploadedAt,
		ProcessedData: nil,
	}
}

func ToDomainImageWithProcessedData(row gen.GetImageWithProcessedDataRow) (*model.ImageMetadata, error) {
	image := ToDomainImage(gen.Image{
		ID:           row.ID,
		OriginalName: row.OriginalName,
		FileName:     row.FileName,
		Status:       row.Status,
		ResultUrl:    row.ResultUrl,
		Size:         row.Size,
		Format:       row.Format,
		UploadedAt:   row.UploadedAt,
	})

	if row.Width.Valid && row.Height.Valid && row.ProcessedName.Valid && row.ProcessedAt.Valid {
		image.ProcessedData = &model.ProcessedData{
			Width:         int(row.Width.Int32),
			Height:        int(row.Height.Int32),
			ProcessedName: row.ProcessedName.String,
			ProcessedAt:   row.ProcessedAt.Time,
		}
	}

	return image, nil
}
