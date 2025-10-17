package queue

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/D1sordxr/image-processor/internal/infrastructure/config"
	"github.com/D1sordxr/image-processor/internal/infrastructure/ticker"
	"github.com/segmentio/kafka-go"
	wbfKafka "github.com/wb-go/wbf/kafka"
)

type Connection struct {
	cfg  *config.Kafka
	conn *kafka.Conn
	*wbfKafka.Producer
	*wbfKafka.Consumer
	isClosed atomic.Bool
}

func New(cfg config.Kafka) *Connection {
	return &Connection{
		cfg:      &cfg,
		Producer: nil,
		Consumer: nil,
		isClosed: atomic.Bool{},
	}
}

const (
	partitions        = 3
	replicationFactor = 1
)

func (w *Connection) createTopic() error {
	if err := w.conn.CreateTopics(kafka.TopicConfig{
		Topic:             w.cfg.ImageTopic,
		NumPartitions:     partitions,
		ReplicationFactor: replicationFactor,
	}); err != nil {
		return fmt.Errorf("create topic: %w", err)
	}
	return nil
}

func (w *Connection) healthCheck() error {
	if w.isClosed.Load() {
		return fmt.Errorf("connection closed")
	}

	if _, err := w.conn.ReadPartitions(); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	return nil
}

func (w *Connection) Run(ctx context.Context) error {
	const op = "kafka.Connection.Run"

	conn, err := kafka.Dial("tcp", w.cfg.Address)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	w.conn = conn
	w.Producer = wbfKafka.NewProducer(w.cfg.PrepWbfProducer())
	w.Consumer = wbfKafka.NewConsumer(w.cfg.PrepWbfConsumer())

	if w.cfg.CreateTopic {
		if err = w.createTopic(); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	healthTicker := ticker.NewHealth()
	defer healthTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-healthTicker.C:
			if err = w.healthCheck(); err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
		}
	}
}

func (w *Connection) Shutdown(ctx context.Context) error {
	const op = "kafka.Connection.Shutdown"

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	if w.isClosed.Load() {
		return fmt.Errorf("connection closed")
	}
	defer w.isClosed.Store(true)

	var errs []error
	errChan := make(chan error, 3)
	done := make(chan struct{})

	wg := sync.WaitGroup{}

	closeResource := func(name string, closer func() error) {
		if closer != nil {
			if err := closer(); err != nil {
				errChan <- fmt.Errorf("%s close: %w", name, err)
			}
		}
	}

	wg.Go(func() { closeResource("connection", w.conn.Close) })
	wg.Go(func() { closeResource("consumer", w.Consumer.Close) })
	wg.Go(func() { closeResource("producer", w.Producer.Close) })

	go func() {
		wg.Wait()
		close(done)
		close(errChan)
	}()

	go func() {
		for err := range errChan {
			errs = append(errs, err)
		}
	}()

	select {
	case <-done:
		if len(errs) > 0 {
			return fmt.Errorf("%s: multiple errors: %v", op, errs)
		}
		return nil
	case <-ctx.Done():
		return fmt.Errorf("%s: %w", op, ctx.Err())
	}
}
