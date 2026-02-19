package app

import (
	"context"
	"fmt"
	"go-hris/internal/messaging/kafka"
	"go-hris/internal/messaging/kafka/producer"
	"go-hris/internal/shared/connection"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func RunWorker() error {
	logger := zap.L().Named("app.worker")

	gormDB, err := connection.ConnectGORMWithRetry(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_SSLMODE"),
		5,
	)
	if err != nil {
		return err
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		return fmt.Errorf("KAFKA_BROKER is required")
	}

	kafkaWriter, err := connection.ConnectKafkaWithRetry(kafkaBroker, 5)
	if err != nil {
		return err
	}
	defer kafkaWriter.Close()

	outboxRepo := kafka.NewOutboxRepository(sqlDB)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go producer.ProcessOutboxEvents(
		ctx,
		outboxRepo,
		kafkaWriter,
		logger,
		3*time.Second,
	)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("worker shutting down")
	cancel()

	return nil
}
