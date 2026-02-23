package app

import (
	"fmt"
	"go-hris/internal/shared/connection"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func BuildApp(router *gin.Engine, logger *zap.Logger) error {
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

	// Register Modules & Routes
	if err := registerModules(router, sqlDB, gormDB, redisClient, logger); err != nil {
		return err
	}

	payslipDir := os.Getenv("PAYSLIP_STORAGE_DIR")
	if payslipDir == "" {
		payslipDir = filepath.Join("storage", "payslips")
	}
	if err := os.MkdirAll(payslipDir, 0o755); err != nil {
		return err
	}
	router.Static("/files/payslips", payslipDir)

	return nil
}
