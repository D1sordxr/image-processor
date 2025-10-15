package handler

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/D1sordxr/image-processor/internal/application/image/input"
	"github.com/D1sordxr/image-processor/internal/application/image/port"
	appPorts "github.com/D1sordxr/image-processor/internal/domain/app/port"
	"github.com/D1sordxr/image-processor/internal/domain/core/image/model"
	"github.com/D1sordxr/image-processor/internal/domain/core/image/vo"
	sharedVO "github.com/D1sordxr/image-processor/internal/domain/core/shared/vo"
	"github.com/D1sordxr/image-processor/internal/transport/http/api/image/dto"
	"github.com/D1sordxr/image-processor/pkg/logger"

	"github.com/wb-go/wbf/ginext"
)

const (
	MaxFileSize = 10 << 20 // 10MB
	CacheMaxAge = 3600     // 1 hour

	ContentTypeJPEG = "image/jpeg"
	ContentTypePNG  = "image/png"
	ContentTypeGIF  = "image/gif"

	ErrImageRequired   = "No image file provided"
	ErrFileTooLarge    = "File too large"
	ErrInvalidImage    = "Invalid image file"
	ErrImageIDRequired = "Image ID is required"
	ErrImageNotFound   = "Image not found"
)

type Handler struct {
	uc      port.UseCase
	log     appPorts.Logger
	baseURL string
}

func New(uc port.UseCase, log appPorts.Logger, baseURL sharedVO.BaseURL) *Handler {
	return &Handler{
		uc:      uc,
		log:     log,
		baseURL: baseURL.String(),
	}
}

func (h *Handler) UploadNewImage(c *ginext.Context) {
	const op = "image.Handler.UploadNewImage"
	logFields := logger.WithFields("operation", op)

	h.log.Info("Starting image upload", logFields()...)

	imageHeader, err := c.FormFile("image")
	if err != nil {
		h.log.Error("No image file provided", logFields("error", err)...)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   ErrImageRequired,
			Details: "Please provide an image file using 'image' form field",
		})
		return
	}

	if err = h.validateFile(imageHeader); err != nil {
		h.log.Error("File validation failed", logFields("error", err)...)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   ErrFileTooLarge,
			Details: fmt.Sprintf("Maximum file size is %dMB", MaxFileSize>>20),
		})
		return
	}

	imageFile, err := imageHeader.Open()
	if err != nil {
		h.log.Error("Failed to open uploaded file", logFields("error", err)...)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to open uploaded file",
			Details: err.Error(),
		})
		return
	}
	defer func() { _ = imageFile.Close() }()

	imageData, err := io.ReadAll(imageFile)
	if err != nil {
		h.log.Error("Failed to read file data", logFields("error", err)...)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to read file data",
			Details: err.Error(),
		})
		return
	}

	if !h.isValidImage(imageHeader.Filename, imageData) {
		h.log.Error("Invalid image file", logFields("filename", imageHeader.Filename)...)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   ErrInvalidImage,
			Details: "Please provide a valid JPEG, PNG, or GIF image",
		})
		return
	}

	opts, err := h.parseProcessingOptions(c)
	if err != nil {
		h.log.Error("Invalid processing options", logFields("error", err)...)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid processing options",
			Details: err.Error(),
		})
		return
	}

	result, err := h.uc.Upload(c.Request.Context(), input.UploadImageInput{
		ImageData: imageData,
		Filename:  imageHeader.Filename,
		Options:   opts,
	})
	if err != nil {
		h.log.Error("Failed to upload image", logFields("error", err)...)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to upload image",
			Details: err.Error(),
		})
		return
	}

	h.log.Info("Image uploaded successfully", logFields(
		"image_id", result.ImageID,
		"file_size", len(imageData),
	)...)

	c.JSON(http.StatusAccepted, dto.SuccessResponse{
		Message: result.Message,
		Data: dto.UploadResponse{
			ImageID:           result.ImageID,
			ResultURL:         h.buildImageURL(result.ImageID),
			ProcessingOptions: opts,
			Message:           result.Message,
		},
	})
}

func (h *Handler) GetProcessedImage(c *ginext.Context) {
	const op = "image.Handler.GetProcessedImage"
	logFields := logger.WithFields("operation", op)

	imageID := c.Param("id")
	if imageID == "" {
		h.log.Error("Image ID is required", logFields()...)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: ErrImageIDRequired,
		})
		return
	}

	h.log.Info("Getting processed image", logFields("image_id", imageID)...)

	result, err := h.uc.Get(c.Request.Context(), input.GetImageInput{
		ImageID: imageID,
	})
	if err != nil {
		h.log.Error("Failed to get image", logFields("error", err, "image_id", imageID)...)
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   ErrImageNotFound,
				Details: fmt.Sprintf("Image with ID %s not found", imageID),
			})
		} else {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "Failed to get image",
				Details: err.Error(),
			})
		}
		return
	}

	if result.Metadata.Status != vo.StatusCompleted {
		h.log.Info("Image not yet processed", logFields("image_id", imageID, "status", result.Metadata.Status)...)
		c.JSON(http.StatusOK, dto.SuccessResponse{
			Message: "Image is still being processed",
			Data: dto.ProcessingStatusResponse{
				Status:   result.Metadata.Status.String(),
				ImageID:  result.Metadata.ID.String(),
				ImageURL: h.buildImageURL(result.Metadata.ID.String()),
				Message:  "Image is still being processed",
			},
		})
		return
	}

	h.log.Info("Returning processed image", logFields("image_id", imageID, "format", result.Metadata.Format)...)

	c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", CacheMaxAge))
	contentType := h.getContentType(result.Metadata.Format)
	c.Data(http.StatusOK, contentType, result.ImageData)
}

func (h *Handler) GetImageStatus(c *ginext.Context) {
	const op = "image.Handler.GetImageStatus"
	logFields := logger.WithFields("operation", op)

	imageID := c.Param("id")
	if imageID == "" {
		h.log.Error("Image ID is required", logFields()...)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: ErrImageIDRequired,
		})
		return
	}

	h.log.Info("Getting image status", logFields("image_id", imageID)...)

	result, err := h.uc.GetStatus(c.Request.Context(), input.GetImageStatusInput{
		ImageID: imageID,
	})
	if err != nil {
		h.log.Error("Failed to get image status", logFields("error", err, "image_id", imageID)...)
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   ErrImageNotFound,
				Details: fmt.Sprintf("Image with ID %s not found", imageID),
			})
		} else {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "Failed to get image status",
				Details: err.Error(),
			})
		}
		return
	}

	h.log.Info("Image status retrieved", logFields("image_id", imageID, "status", result.Status)...)

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Data: dto.ProcessingStatusResponse{
			Status:   result.Status,
			ImageID:  imageID,
			ImageURL: h.buildImageURL(imageID),
			Message:  fmt.Sprintf("Image status: %s", result.Status),
		},
	})
}

func (h *Handler) DeleteImage(c *ginext.Context) {
	const op = "image.Handler.DeleteImage"
	logFields := logger.WithFields("operation", op)

	imageID := c.Param("id")
	if imageID == "" {
		h.log.Error("Image ID is required", logFields()...)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: ErrImageIDRequired,
		})
		return
	}

	h.log.Info("Deleting image", logFields("image_id", imageID)...)

	result, err := h.uc.Delete(c.Request.Context(), input.DeleteImageInput{
		ImageID: imageID,
	})
	if err != nil {
		h.log.Error("Failed to delete image", logFields("error", err, "image_id", imageID)...)
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   ErrImageNotFound,
				Details: fmt.Sprintf("Image with ID %s not found", imageID),
			})
		} else {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "Failed to delete image",
				Details: err.Error(),
			})
		}
		return
	}

	h.log.Info("Image deleted successfully", logFields("image_id", imageID)...)

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: result.Message,
		Data: map[string]string{
			"image_id": imageID,
		},
	})
}

func (h *Handler) ProcessImageSync(c *ginext.Context) {
	const op = "image.Handler.ProcessImageSync"
	logFields := logger.WithFields("operation", op)

	imageID := c.Param("id")
	if imageID == "" {
		h.log.Error("Image ID is required", logFields()...)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: ErrImageIDRequired,
		})
		return
	}

	h.log.Warn("ProcessSync called but not implemented", logFields("image_id", imageID)...)

	result, err := h.uc.ProcessSync(c.Request.Context(), input.ProcessImageSyncInput{
		ImageID: imageID,
	})
	if err != nil {
		h.log.Error("Failed to process image synchronously", logFields("error", err, "image_id", imageID)...)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to process image",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: result.Message,
		Data: map[string]string{
			"image_id": imageID,
		},
	})
}

func (h *Handler) HealthCheck(c *ginext.Context) {
	const op = "image.Handler.HealthCheck"
	logFields := logger.WithFields("operation", op)

	h.log.Info("Health check requested", logFields()...)

	response := dto.HealthCheckResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Service:   "github.com/D1sordxr/image-processor-service",
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Data: response,
	})
}

func (h *Handler) ServeFrontend(c *ginext.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

func (h *Handler) RegisterRoutes(router *ginext.RouterGroup) {
	router.POST("/upload", h.UploadNewImage)
	router.GET("/image/:id", h.GetProcessedImage)
	router.GET("/image/:id/status", h.GetImageStatus)
	// router.POST("/image/:id/process", h.ProcessImageSync)
	router.DELETE("/image/:id", h.DeleteImage)
	router.GET("/health", h.HealthCheck)
	router.GET("/", h.ServeFrontend)
}

func (h *Handler) validateFile(fileHeader *multipart.FileHeader) error {
	if fileHeader.Size > MaxFileSize {
		return fmt.Errorf("file size %d exceeds maximum %d", fileHeader.Size, MaxFileSize)
	}
	return nil
}

func (h *Handler) isValidImage(filename string, data []byte) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExtensions := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
	}
	if !validExtensions[ext] {
		return false
	}

	if len(data) < 8 {
		return false
	}

	// Check file signatures
	switch {
	case data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF: // JPEG
		return true
	case data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47: // PNG
		return true
	case string(data[:6]) == "GIF87a" || string(data[:6]) == "GIF89a": // GIF
		return true
	default:
		return false
	}
}

func (h *Handler) getContentType(format string) string {
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		return ContentTypeJPEG
	case "png":
		return ContentTypePNG
	case "gif":
		return ContentTypeGIF
	default:
		return "application/octet-stream"
	}
}

func (h *Handler) buildImageURL(imageID string) string {
	return fmt.Sprintf("%s/image/%s", h.baseURL, imageID)
}

func (h *Handler) parseProcessingOptions(c *ginext.Context) (model.ProcessingOptions, error) {
	var (
		opts    = model.ProcessingOptions{}
		readOpt = func(key string) string {
			if val := c.PostForm(key); val != "" {
				return val
			}
			return c.Query(key)
		}
	)

	// Validate and parse width
	if widthStr := readOpt("width"); widthStr != "" {
		width, err := strconv.Atoi(widthStr)
		if err != nil || width <= 0 {
			return opts, fmt.Errorf("invalid width: must be positive integer")
		}
		opts.Width = width
	}

	// Validate and parse height
	if heightStr := readOpt("height"); heightStr != "" {
		height, err := strconv.Atoi(heightStr)
		if err != nil || height <= 0 {
			return opts, fmt.Errorf("invalid height: must be positive integer")
		}
		opts.Height = height
	}

	// Validate and parse quality
	if qualityStr := readOpt("quality"); qualityStr != "" {
		quality, err := strconv.Atoi(qualityStr)
		if err != nil || quality < 1 || quality > 100 {
			return opts, fmt.Errorf("invalid quality: must be between 1 and 100")
		}
		opts.Quality = quality
	}

	// Validate and parse format
	if format := readOpt("format"); format != "" {
		format = strings.ToLower(format)
		if format != "jpeg" && format != "jpg" && format != "png" && format != "gif" {
			return opts, fmt.Errorf("invalid format: supported formats are jpeg, png, gif")
		}
		opts.Format = format
	}

	// Parse watermark text
	if watermark := readOpt("watermark"); watermark != "" {
		opts.WatermarkText = watermark
	}

	// Validate and parse thumbnail flag
	if thumbnail := readOpt("thumbnail"); thumbnail != "" {
		thumb, err := strconv.ParseBool(thumbnail)
		if err != nil {
			return opts, fmt.Errorf("invalid thumbnail value: must be true or false")
		}
		opts.Thumbnail = thumb
	}

	return opts, nil
}
