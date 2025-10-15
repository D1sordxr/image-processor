package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	appPorts "github.com/D1sordxr/image-processor/internal/domain/app/port"
	"github.com/D1sordxr/image-processor/internal/domain/core/image/model"
	"github.com/D1sordxr/image-processor/internal/domain/core/image/options"
	"github.com/D1sordxr/image-processor/pkg/logger"
	wbfKafka "github.com/wb-go/wbf/kafka"
)

type Queue struct {
	log      appPorts.Logger
	producer *wbfKafka.Producer
	topic    string
}

func New(
	log appPorts.Logger,
	producer *wbfKafka.Producer,
	topic string,
) *Queue {
	return &Queue{
		log:      log,
		producer: producer,
		topic:    topic,
	}
}

func (q *Queue) PublishImageTask(ctx context.Context, imageID string) error {
	const op = "image.Queue.PublishImageTask"
	logFields := logger.WithFields("operation", op, "image_id", imageID)

	message := model.ProcessingImage{
		ImageID:   imageID,
		Timestamp: time.Now(),
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		q.log.Error("Failed to marshal message", logFields("error", err)...)
		return fmt.Errorf("%s: failed to marshal message: %w", op, err)
	}

	err = q.producer.SendWithRetry(ctx, options.BrokerStrategy, []byte(imageID), messageBytes)
	if err != nil {
		q.log.Error("Failed to publish task to Kafka", logFields("error", err)...)
		return fmt.Errorf("%s: failed to publish task to Kafka: %w", op, err)
	}

	q.log.Info("Image task published to Kafka", logFields(
		"topic", q.topic,
	)...)

	return nil
}

func (q *Queue) Close() error {
	const op = "image.Queue.Close"

	if err := q.producer.Close(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
