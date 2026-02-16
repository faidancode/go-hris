package main

import (
	"log"
	"os"
	"time"

	"go-hris/internal/app"
	"go-hris/internal/bootstrap"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	// apperror.Init() initialize custom error handling (if needed)
	r := gin.Default()

	// build dependency + routes
	if err := app.BuildApp(r); err != nil {
		log.Fatal(err)
	}

	auditLogger := bootstrap.NewStdoutAuditLogger()
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	bootstrap.StartHTTPServer(
		r,
		bootstrap.ServerConfig{
			Port:         port,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		auditLogger,
	)
}
