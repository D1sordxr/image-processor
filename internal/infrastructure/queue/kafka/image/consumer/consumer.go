package consumer

import (
	"context"
	"encoding/json"
	appPorts "github.com/D1sordxr/image-processor/internal/domain/app/port"
	"github.com/D1sordxr/image-processor/internal/domain/core/image/model"
	"github.com/D1sordxr/image-processor/internal/domain/core/image/options"
	"github.com/D1sordxr/image-processor/pkg/logger"
	"github.com/segmentio/kafka-go"
	wbfKafka "github.com/wb-go/wbf/kafka"
)

type Consumer struct {
	log      appPorts.Logger
	consumer *wbfKafka.Consumer
	topic    string
}

func New(
	log appPorts.Logger,
	consumer *wbfKafka.Consumer,
	topic string,
) *Consumer {
	return &Consumer{
		log:      log,
		consumer: consumer,
		topic:    topic,
	}
}

func (c *Consumer) StartProcessing(
	ctx context.Context,
	processor func(context.Context, *model.ProcessingImage) error,
) error {
	const op = "Consumer.StartProcessing"
	logFields := logger.WithFields("operation", op)

	c.log.Info("Starting Kafka consumer", logFields("topic", c.topic)...)

	messages := make(chan kafka.Message, 128)

	go func() {
		defer close(messages)
		c.consumer.StartConsuming(ctx, messages, options.BrokerStrategy)
	}()

	for {
		select {
		case <-ctx.Done():
			c.log.Info("Context cancelled, stopping consumer", logFields()...)
			return nil

		case msg, ok := <-messages:
			if !ok {
				c.log.Info("Messages channel closed", logFields()...)
				return nil
			}

			var processingImage model.ProcessingImage
			if err := json.Unmarshal(msg.Value, &processingImage); err != nil {
				c.log.Error("Failed to unmarshal message",
					logFields("error", err, "offset", msg.Offset)...)
				continue
			}

			msgLogFields := logger.WithFields(
				"operation", op,
				"image_id", processingImage.ImageID,
				"offset", msg.Offset,
			)

			if err := processor(ctx, &processingImage); err != nil {
				c.log.Error("Failed to process image", msgLogFields("error", err)...)
				continue
			}

			if err := c.consumer.Commit(ctx, msg); err != nil {
				c.log.Error("Failed to commit message", msgLogFields("error", err)...)
			} else {
				c.log.Debug("Successfully committed message", msgLogFields()...)
			}
		}
	}
}
