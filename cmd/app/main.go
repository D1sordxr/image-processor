package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/D1sordxr/image-processor/internal/application/image/usecase"
	"github.com/D1sordxr/image-processor/internal/domain/core/shared/vo"
	"github.com/D1sordxr/image-processor/internal/infrastructure/image/processor"
	"github.com/D1sordxr/image-processor/internal/infrastructure/queue"
	"github.com/D1sordxr/image-processor/internal/infrastructure/queue/kafka/image/consumer"
	"github.com/D1sordxr/image-processor/internal/infrastructure/queue/kafka/image/producer"
	"github.com/D1sordxr/image-processor/internal/infrastructure/storage/minio"
	"github.com/D1sordxr/image-processor/internal/infrastructure/storage/minio/repositories/image/s3repo"
	"github.com/D1sordxr/image-processor/internal/infrastructure/storage/postgres/executor"
	"github.com/D1sordxr/image-processor/internal/infrastructure/storage/postgres/repositories/image/repo"
	"github.com/D1sordxr/image-processor/internal/infrastructure/storage/postgres/txmanager"
	defaultWorker "github.com/D1sordxr/image-processor/internal/infrastructure/worker"
	"github.com/D1sordxr/image-processor/internal/transport/http/api/image/handler"
	"github.com/D1sordxr/image-processor/internal/transport/kafka/handler/image"

	loadApp "github.com/D1sordxr/image-processor/internal/infrastructure/app"
	"github.com/D1sordxr/image-processor/internal/infrastructure/config"
	"github.com/D1sordxr/image-processor/internal/infrastructure/logger"
	"github.com/D1sordxr/image-processor/internal/infrastructure/storage/postgres"
	"github.com/D1sordxr/image-processor/internal/transport/http"

	"github.com/rs/zerolog"
	"github.com/wb-go/wbf/zlog"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := config.NewAppConfig()
	log := logger.New(defaultLogger)
	log.Debug("Application data", "config", cfg)

	storageConn := postgres.New(cfg.Storage)
	s3Conn := minio.New(cfg.S3Storage)
	brokerConn := queue.New(cfg.Broker)

	storageExecutor := executor.New(storageConn.Storage)
	txManager := txmanager.New(storageExecutor)

	imageRepo := repo.New(storageExecutor)
	imageS3Repo := s3repo.New(s3Conn, cfg.S3Storage.BucketName)
	imageProducer := producer.New(log, brokerConn.Producer, cfg.Broker.ImageTopic)
	imageConsumer := consumer.New(log, brokerConn.Consumer, cfg.Broker.ImageTopic)
	imageProcessor := processor.New()
	imageUC := usecase.New(
		log,
		txManager,
		imageRepo,
		imageS3Repo,
		imageProducer,
		imageProcessor,
		vo.NewBaseURL(cfg.Server.Host, cfg.Server.Port),
	)
	imageProcessorWorkerHandler := image.NewProcessorHandler(log, imageConsumer, imageUC)
	imageHttpHandler := handler.New(log, imageUC, vo.NewBaseURL(cfg.Server.Host, cfg.Server.Port))

	worker := defaultWorker.New(
		log,
		imageProcessorWorkerHandler,
	)
	httpServer := http.NewServer(
		log,
		&cfg.Server,
		imageHttpHandler,
	)

	app := loadApp.NewApp(
		log,
		brokerConn,
		storageConn,
		s3Conn,
		httpServer,
		worker,
	)
	app.Run(ctx)
}

var defaultLogger zerolog.Logger

func init() {
	zlog.Init()
	defaultLogger = zlog.Logger
}
