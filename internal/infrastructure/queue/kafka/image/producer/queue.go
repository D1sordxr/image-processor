package producer

import (
	"context"
	"encoding/json"
	"fmt"

	appPorts "github.com/D1sordxr/image-processor/internal/domain/app/port"
	"github.com/D1sordxr/image-processor/internal/domain/core/image/model"
	"github.com/D1sordxr/image-processor/internal/domain/core/image/options"
	"github.com/D1sordxr/image-processor/pkg/logger"
	wbfKafka "github.com/wb-go/wbf/kafka"
)

type Producer struct {
	log      appPorts.Logger
	producer *wbfKafka.Producer
	topic    string
}

func New(
	log appPorts.Logger,
	producer *wbfKafka.Producer,
	topic string,
) *Producer {
	return &Producer{
		log:      log,
		producer: producer,
		topic:    topic,
	}
}

func (p *Producer) Publish(ctx context.Context, message *model.ProcessingImage) error {
	const op = "image.Producer.Publish"
	logFields := logger.WithFields("operation", op, "image_id", message.ImageID)

	messageBytes, err := json.Marshal(message)
	if err != nil {
		p.log.Error("Failed to marshal message", logFields("error", err)...)
		return fmt.Errorf("%s: failed to marshal message: %w", op, err)
	}

	err = p.producer.SendWithRetry(ctx, options.BrokerStrategy, []byte(message.ImageID), messageBytes)
	if err != nil {
		p.log.Error("Failed to publish task to Kafka", logFields("error", err)...)
		return fmt.Errorf("%s: failed to publish task to Kafka: %w", op, err)
	}

	p.log.Info("Image task published to Kafka", logFields(
		"topic", p.topic,
	)...)

	return nil
}

func (p *Producer) Close() error {
	const op = "image.Producer.Close"

	if err := p.producer.Close(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
