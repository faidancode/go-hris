package app

import (
	"fmt"
	"go-hris/internal/shared/connection"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func BuildApp(router *gin.Engine) error {
	// 1. Setup Infrastructure
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
		log.Fatal(err)
	}
	log.Println("Database connection established")

	sqlDB, err := gormDB.DB()
	if err != nil {
		return err
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		return fmt.Errorf("REDIS_ADDR is required")
	}

	redisClient, err := connection.ConnectRedisWithRetry(redisAddr, 5)
	if err != nil {
		return err
	}
	log.Println("Redis connection established")

	kafkaWriter, err := connection.ConnectKafkaWithRetry(os.Getenv("KAFKA_BROKER"), 5)
	if err != nil {
		return err
	}
	_ = kafkaWriter
	log.Println("Kafka connection established")

	// Register Modules & Routes
	if err := registerModules(router, sqlDB, gormDB, redisClient); err != nil {
		return err
	}

	return nil
}
