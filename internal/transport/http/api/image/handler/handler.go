package handler

import (
	"fmt"
	"github.com/D1sordxr/image-processor/internal/application/image/input"
	"github.com/D1sordxr/image-processor/internal/application/image/port"
	"github.com/D1sordxr/image-processor/internal/domain/core/image/model"
	"github.com/D1sordxr/image-processor/internal/transport/http/api/image/dto"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/wb-go/wbf/ginext"
)

type Handler struct {
	uc port.UseCase
}

func New(uc port.UseCase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) UploadNewImage(c *ginext.Context) {
	imageHeader, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "No image file provided",
			Details: "Please provide an image file using 'image' form field",
		})
		return
	}

	if imageHeader.Size > 10<<20 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "File too large",
			Details: "Maximum file size is 10MB",
		})
		return
	}

	imageFile, err := imageHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to open uploaded file",
			Details: err.Error(),
		})
		return
	}
	defer func() { _ = imageFile.Close() }()

	imageData, err := io.ReadAll(imageFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to read file data",
			Details: err.Error(),
		})
		return
	}

	if !isValidImage(imageHeader.Filename, imageData) {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid image file",
			Details: "Please provide a valid JPEG, PNG, or GIF image",
		})
		return
	}

	opts, callbackURL, err := parseProcessingOptions(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid processing options",
			Details: err.Error(),
		})
		return
	}

	result, err := h.uc.UploadImage(c.Request.Context(), input.UploadImageInput{
		ImageData:   imageData,
		Filename:    imageHeader.Filename,
		Options:     opts,
		CallbackURL: callbackURL,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to upload image",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusAccepted, dto.SuccessResponse{
		Message: result.Message,
		Data: dto.UploadResponse{
			ImageID:           result.ImageID,
			ResultURL:         fmt.Sprintf("%s/image/%s", c.Request.Host, result.ImageID),
			ProcessingOptions: opts,
			Message:           result.Message,
		},
	})
}

func (h *Handler) GetProcessedImage(c *ginext.Context) {
	imageID := c.Param("id")
	if imageID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Image ID is required",
		})
		return
	}

	result, err := h.uc.GetImage(c.Request.Context(), input.GetImageInput{
		ImageID: imageID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Image not found",
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

	if result.Metadata.Status != "completed" {
		c.JSON(http.StatusOK, dto.SuccessResponse{
			Message: "Image is still being processed",
			Data: dto.ProcessingStatusResponse{
				Status:   result.Metadata.Status,
				ImageID:  result.Metadata.ID,
				ImageURL: fmt.Sprintf("%s/image/%s", c.Request.Host, result.Metadata.ID),
				Message:  "Image is still being processed",
			},
		})
		return
	}

	contentType := getContentType(result.Metadata.Format)
	c.Data(http.StatusOK, contentType, result.ImageData)
}

func (h *Handler) GetImageStatus(c *ginext.Context) {
	imageID := c.Param("id")
	if imageID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Image ID is required",
		})
		return
	}

	result, err := h.uc.GetImageStatus(c.Request.Context(), input.GetImageStatusInput{
		ImageID: imageID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Image not found",
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

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Data: dto.ProcessingStatusResponse{
			Status:   result.Status,
			ImageID:  imageID,
			ImageURL: fmt.Sprintf("%s/image/%s", c.Request.Host, imageID),
			Message:  fmt.Sprintf("Image status: %s", result.Status),
		},
	})
}

func (h *Handler) DeleteImage(c *ginext.Context) {
	imageID := c.Param("id")
	if imageID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Image ID is required",
		})
		return
	}

	result, err := h.uc.DeleteImage(c.Request.Context(), input.DeleteImageInput{
		ImageID: imageID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Image not found",
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

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: result.Message,
		Data: map[string]string{
			"image_id": imageID,
		},
	})
}

func (h *Handler) ProcessImageSync(c *ginext.Context) {
	imageID := c.Param("id")
	if imageID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Image ID is required",
		})
		return
	}

	result, err := h.uc.ProcessImageSync(c.Request.Context(), input.ProcessImageSyncInput{
		ImageID: imageID,
	})
	if err != nil {
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
	router.POST("/image/:id/process", h.ProcessImageSync)
	router.DELETE("/image/:id", h.DeleteImage)
	router.GET("/health", h.HealthCheck)
	router.GET("/", h.ServeFrontend)
}

func isValidImage(filename string, data []byte) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExtensions := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
	}
	if !validExtensions[ext] {
		return false
	}

	if len(data) < 4 {
		return false
	}

	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return true
	}
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return true
	}
	if string(data[:6]) == "GIF87a" || string(data[:6]) == "GIF89a" {
		return true
	}

	return false
}

func getContentType(format string) string {
	switch format {
	case "jpeg", "jpg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	default:
		return "application/octet-stream"
	}
}

func parseProcessingOptions(c *ginext.Context) (model.ProcessingOptions, string, error) {
	var (
		opts        = model.ProcessingOptions{}
		callbackURL string
		readOpt     = func(key string) string {
			if val := c.PostForm(key); val != "" {
				return val
			}
			return c.Query(key)
		}
	)

	if widthStr := readOpt("width"); widthStr != "" {
		width, err := strconv.Atoi(widthStr)
		if err != nil || width <= 0 {
			return opts, callbackURL, fmt.Errorf("invalid width: must be positive integer")
		}
		opts.Width = width
	}

	if heightStr := readOpt("height"); heightStr != "" {
		height, err := strconv.Atoi(heightStr)
		if err != nil || height <= 0 {
			return opts, callbackURL, fmt.Errorf("invalid height: must be positive integer")
		}
		opts.Height = height
	}

	if qualityStr := readOpt("quality"); qualityStr != "" {
		quality, err := strconv.Atoi(qualityStr)
		if err != nil || quality < 1 || quality > 100 {
			return opts, callbackURL, fmt.Errorf("invalid quality: must be between 1 and 100")
		}
		opts.Quality = quality
	}

	if format := readOpt("format"); format != "" {
		format = strings.ToLower(format)
		if format != "jpeg" && format != "jpg" && format != "png" && format != "gif" {
			return opts, callbackURL, fmt.Errorf("invalid format: supported formats are jpeg, png, gif")
		}
		opts.Format = format
	}

	if watermark := readOpt("watermark"); watermark != "" {
		opts.WatermarkText = watermark
	}

	if thumbnail := readOpt("thumbnail"); thumbnail != "" {
		thumb, err := strconv.ParseBool(thumbnail)
		if err != nil {
			return opts, callbackURL, fmt.Errorf("invalid thumbnail value: must be true or false")
		}
		opts.Thumbnail = thumb
	}

	return opts, callbackURL, nil
}
