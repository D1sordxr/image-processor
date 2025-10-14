package image

import (
	"context"
	"github.com/D1sordxr/image-processor/internal/application/image/port"
	appPorts "github.com/D1sordxr/image-processor/internal/domain/app/port"
	domainPort "github.com/D1sordxr/image-processor/internal/domain/core/image/port"
)

type ProcessorHandler struct {
	log      appPorts.Logger
	consumer domainPort.Consumer
	uc       port.UseCase
}

func NewProcessorHandler(
	log appPorts.Logger,
	consumer domainPort.Consumer,
	uc port.UseCase,
) *ProcessorHandler {
	return &ProcessorHandler{
		log:      log,
		consumer: consumer,
		uc:       uc,
	}
}

func (h *ProcessorHandler) Start(ctx context.Context) error {
	const op = "image.ProcessorHandler.Start"

	h.log.Info("Starting image processor handler", "operation", op)

	return h.consumer.StartProcessing(ctx, h.uc.ProcessImage)
}
