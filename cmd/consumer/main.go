package main

import (
	"go-hris/internal/app"
	"go-hris/internal/shared/apperror"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	_ = godotenv.Load()
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	apperror.Init()

	if err := app.RunConsumer(); err != nil {
		logger.Fatal("run consumer failed", zap.Error(err))
	}
}
