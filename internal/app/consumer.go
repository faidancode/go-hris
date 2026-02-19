package app

import (
	"context"
	"fmt"
	"go-hris/internal/employeesalary"
	"go-hris/internal/events"
	"go-hris/internal/messaging/kafka/consumer"
	"go-hris/internal/shared/connection"
	"os"
	"os/signal"
	"syscall"

	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

func RunConsumer() error {
	logger := zap.L().Named("app.consumer")

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

	employeeSalaryRepo := employeesalary.NewRepository(gormDB)
	employeeSalaryService := employeesalary.NewService(sqlDB, employeeSalaryRepo)

	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:        []string{kafkaBroker},
		Topic:          events.EmployeeCreatedTopic,
		GroupID:        "go-hris-employee-salary",
		CommitInterval: 0,
		StartOffset:    kafkago.FirstOffset,
	})
	defer reader.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go consumer.ConsumeEmployeeLifecycle(ctx, reader, employeeSalaryService, logger)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("consumer shutting down")
	cancel()

	return nil
}
