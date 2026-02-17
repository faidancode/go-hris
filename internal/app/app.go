package app

import (
	"go-hris/internal/shared/connection"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func BuildApp(router *gin.Engine) error {
	// 1. Setup Infrastructure
	_, err := connection.ConnectGORMWithRetry(
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
	log.Println("✅ Database connection established")

	redisClient, err := connection.ConnectRedisWithRetry(os.Getenv("REDIS_ADDR"), 5)
	if err != nil {
		return err
	}
	log.Println("✅ Redis connection established")
	_ = redisClient

	// Register Modules & Routes
	// registerModules(router, db, redisClient, cloudinaryService)

	return nil
}
