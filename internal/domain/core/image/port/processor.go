package port

import "github.com/D1sordxr/image-processor/internal/domain/core/image/model"

type ImageProcessor interface {
	ProcessImage(imageData []byte, options model.ProcessingOptions) (*model.ProcessingResult, error)
}
