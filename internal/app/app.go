package app

import (
	"go-hris/internal/shared/connection"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func BuildApp(router *gin.Engine) error {
	// 1. Setup Infrastructure
	_, err := connection.ConnectGORMWithRetry(os.Getenv("DB_URL"), 5)
	if err != nil {
		return err
	}
	log.Println("✅ Database connection established")

	// Register Modules & Routes
	// registerModules(router, db, redisClient, cloudinaryService)

	return nil
}
