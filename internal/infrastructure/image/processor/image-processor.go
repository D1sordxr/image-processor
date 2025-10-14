package processor

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/D1sordxr/image-processor/internal/domain/core/image/model"
	"image"
	"image/color"
	"image/color/palette"
	"image/gif"
	"image/jpeg"
	"image/png"
	"strings"

	"golang.org/x/image/draw"
	"golang.org/x/image/font"

	"time"

	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const (
	DefaultThumbnailSize = 150
	DefaultQuality       = 85
	DefaultWatermarkX    = 10
	DefaultWatermarkY    = 20
)

var (
	ErrWrongBounds = errors.New("wrong bounds of original image")
)

type DefaultProcessor struct{}

func NewDefaultProcessor() *DefaultProcessor {
	return &DefaultProcessor{}
}

func (p *DefaultProcessor) ProcessImage(imageData []byte, opts model.ProcessingOptions) (*model.ProcessingResult, error) {
	start := time.Now()

	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return &model.ProcessingResult{}, fmt.Errorf("failed to decode image: %w", err)
	}

	if opts.Width > 0 || opts.Height > 0 {
		img, err = p.Resize(img, opts.Width, opts.Height)
		if err != nil {
			return &model.ProcessingResult{}, fmt.Errorf("resize failed: %v", err)
		}
	}

	if opts.Thumbnail {
		size := DefaultThumbnailSize
		if opts.Width > 0 {
			size = opts.Width
		}
		img, err = p.CreateThumbnail(img, size)
		if err != nil {
			return &model.ProcessingResult{}, fmt.Errorf("thumbnail creation failed: %v", err)
		}
	}

	if opts.WatermarkText != "" {
		img, err = p.AddWatermark(img, opts.WatermarkText)
		if err != nil {
			return &model.ProcessingResult{}, fmt.Errorf("watermark adding failed: %v", err)
		}
	}

	outputFormat := opts.Format
	if outputFormat == "" {
		outputFormat = format
	}
	processedImageData, err := p.ConvertFormat(img, outputFormat, opts.Quality)
	if err != nil {
		return nil, fmt.Errorf("format conversion failed: %w", err)
	}

	return &model.ProcessingResult{
		ProcessedData:  processedImageData,
		Format:         outputFormat,
		Width:          img.Bounds().Dx(),
		Height:         img.Bounds().Dy(),
		Size:           int64(len(processedImageData)),
		ProcessingTime: time.Since(start),
	}, nil

}

func (p *DefaultProcessor) Resize(originalImage image.Image, newWidth int, newHeight int) (image.Image, error) {
	if originalImage.Bounds().Dx() <= 0 || originalImage.Bounds().Dy() <= 0 {
		return originalImage, ErrWrongBounds
	}
	if newWidth <= 0 && newHeight <= 0 {
		return originalImage, nil
	}

	resultImage := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.ApproxBiLinear.Scale(
		resultImage,
		resultImage.Bounds(),
		originalImage,
		originalImage.Bounds(),
		draw.Src,
		nil)

	return resultImage, nil
}

func (p *DefaultProcessor) CreateThumbnail(originalImage image.Image, size int) (image.Image, error) {

	originalBounds := originalImage.Bounds()
	if originalBounds.Dx() <= 0 || originalBounds.Dy() <= 0 {
		return originalImage, ErrWrongBounds
	}

	var newWidth, newHeight int
	if originalBounds.Dx() >= originalBounds.Dy() {
		newWidth = size
		ratio := float64(size) / float64(originalBounds.Dx())
		newHeight = int(float64(originalBounds.Dy()) * ratio)
	} else {
		newHeight = size
		ratio := float64(size) / float64(originalBounds.Dy())
		newWidth = int(float64(originalBounds.Dx()) * ratio)
	}

	return p.Resize(originalImage, newWidth, newHeight)
}

func (p *DefaultProcessor) AddWatermark(originalImage image.Image, text string) (image.Image, error) {
	originalBounds := originalImage.Bounds()
	if originalBounds.Dx() <= 0 || originalBounds.Dy() <= 0 {
		return originalImage, ErrWrongBounds
	}

	resultImage := image.NewRGBA(image.Rect(0, 0, originalBounds.Dx(), originalBounds.Dy()))

	draw.Draw(resultImage, image.Rect(0, 0, originalBounds.Dx(), originalBounds.Dy()), originalImage, image.Point{}, draw.Src)

	//цвет и шрифт
	watermarkFace := basicfont.Face7x13
	textColor := color.RGBA{255, 255, 255, 128}

	// Позиция знака
	textWidth := len(text) * 7
	signX := originalBounds.Dx() - textWidth - 10
	if signX < DefaultWatermarkX {
		signX = DefaultWatermarkX
	}
	signY := originalBounds.Dy() - 10
	if signY < DefaultWatermarkY {
		signY = DefaultWatermarkY
	}

	point := fixed.Point26_6{
		X: fixed.I(signX),
		Y: fixed.I(signY),
	}

	drawer := &font.Drawer{
		Dst:  resultImage,
		Src:  image.NewUniform(textColor),
		Face: watermarkFace,
		Dot:  point,
	}
	drawer.DrawString(text)
	return resultImage, nil
}

func (p *DefaultProcessor) ConvertFormat(originalImage image.Image, format string, quality int) ([]byte, error) {
	if originalImage.Bounds().Dx() <= 0 || originalImage.Bounds().Dy() <= 0 {
		return nil, ErrWrongBounds
	}

	var buf bytes.Buffer
	var err error

	if quality <= 0 {
		quality = DefaultQuality
	}

	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		err = jpeg.Encode(&buf, originalImage, &jpeg.Options{Quality: quality})
	case "png":
		err = png.Encode(&buf, originalImage)
	case "gif":
		if gifImg, ok := originalImage.(*image.Paletted); ok {
			err = gif.Encode(&buf, gifImg, &gif.Options{})
		} else {
			paletted := image.NewPaletted(originalImage.Bounds(), palette.Plan9)
			draw.Draw(paletted, originalImage.Bounds(), originalImage, image.Point{}, draw.Src)
			err = gif.Encode(&buf, paletted, &gif.Options{})
		}

	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to encode %s: %v", format, err)
	}
	return buf.Bytes(), nil
}
