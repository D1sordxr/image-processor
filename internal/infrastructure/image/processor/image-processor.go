package processor

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	"image/gif"
	"image/jpeg"
	"image/png"
	"strings"
	"time"

	"github.com/D1sordxr/image-processor/internal/domain/core/image/model"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const (
	DefaultThumbnailSize = 150
	DefaultQuality       = 85
	DefaultWatermarkX    = 0
	DefaultWatermarkY    = 0
)

var (
	ErrWrongBounds            = errors.New("wrong bounds of original image")
	ErrEmptyImageData         = errors.New("empty image data")
	ErrInvalidDimensions      = errors.New("invalid dimensions: width and height must be non-negative")
	ErrInvalidQuality         = errors.New("invalid quality: must be between 0 and 100")
	ErrUnsupportedFormat      = errors.New("unsupported format")
	ErrImageDecodeFailed      = errors.New("failed to decode image")
	ErrResizeFailed           = errors.New("resize failed")
	ErrThumbnailFailed        = errors.New("thumbnail creation failed")
	ErrWatermarkFailed        = errors.New("watermark adding failed")
	ErrFormatConversionFailed = errors.New("format conversion failed")
	ErrImageEncodeFailed      = errors.New("failed to encode image")
)

const (
	opProcessImage    = "image.Processor.Process"
	opResize          = "image.Processor.Resize"
	opCreateThumbnail = "image.Processor.CreateThumbnail"
	opAddWatermark    = "image.Processor.AddWatermark"
	opConvertFormat   = "image.Processor.ConvertFormat"
)

type Processor struct{}

func New() *Processor {
	return &Processor{}
}

func (p *Processor) ProcessImage(imageData []byte, opts model.ProcessingOptions) (*model.ProcessingResult, error) {
	const op = opProcessImage
	start := time.Now()

	if len(imageData) == 0 {
		return nil, fmt.Errorf("%s: %w", op, ErrEmptyImageData)
	}

	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %w", op, ErrImageDecodeFailed, err)
	}

	if opts.Width > 0 || opts.Height > 0 {
		img, err = p.Resize(img, opts.Width, opts.Height)
		if err != nil {
			return nil, fmt.Errorf("%s: %w: %w", op, ErrResizeFailed, err)
		}
	}

	if opts.Thumbnail {
		size := DefaultThumbnailSize
		if opts.Width > 0 {
			size = opts.Width
		}
		img, err = p.CreateThumbnail(img, size)
		if err != nil {
			return nil, fmt.Errorf("%s: %w: %w", op, ErrThumbnailFailed, err)
		}
	}

	if opts.WatermarkText != "" {
		img, err = p.AddWatermark(img, opts.WatermarkText)
		if err != nil {
			return nil, fmt.Errorf("%s: %w: %w", op, ErrWatermarkFailed, err)
		}
	}

	outputFormat := opts.Format
	if outputFormat == "" {
		outputFormat = format
	}
	processedImageData, err := p.ConvertFormat(img, outputFormat, opts.Quality)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %w", op, ErrFormatConversionFailed, err)
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

func (p *Processor) Resize(originalImage image.Image, newWidth int, newHeight int) (image.Image, error) {
	const op = opResize

	if originalImage.Bounds().Dx() <= 0 || originalImage.Bounds().Dy() <= 0 {
		return nil, fmt.Errorf("%s: %w", op, ErrWrongBounds)
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

func (p *Processor) CreateThumbnail(originalImage image.Image, size int) (image.Image, error) {
	const op = opCreateThumbnail

	originalBounds := originalImage.Bounds()
	if originalBounds.Dx() <= 0 || originalBounds.Dy() <= 0 {
		return nil, fmt.Errorf("%s: %w", op, ErrWrongBounds)
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

func (p *Processor) AddWatermark(originalImage image.Image, text string) (image.Image, error) {
	const op = opAddWatermark

	originalBounds := originalImage.Bounds()
	if originalBounds.Dx() <= 0 || originalBounds.Dy() <= 0 {
		return nil, fmt.Errorf("%s: %w", op, ErrWrongBounds)
	}

	resultImage := image.NewRGBA(image.Rect(0, 0, originalBounds.Dx(), originalBounds.Dy()))

	draw.Draw(resultImage, image.Rect(0, 0, originalBounds.Dx(), originalBounds.Dy()), originalImage, image.Point{}, draw.Src)

	// цвет и шрифт
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

func (p *Processor) ConvertFormat(originalImage image.Image, format string, quality int) ([]byte, error) {
	const op = opConvertFormat

	if originalImage.Bounds().Dx() <= 0 || originalImage.Bounds().Dy() <= 0 {
		return nil, fmt.Errorf("%s: %w", op, ErrWrongBounds)
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
		return nil, fmt.Errorf("%s: %w: %s", op, ErrUnsupportedFormat, format)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %w", op, ErrImageEncodeFailed, err)
	}
	return buf.Bytes(), nil
}
