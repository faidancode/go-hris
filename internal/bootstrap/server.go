package bootstrap

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// StartHTTPServer menjalankan Gin server dengan graceful shutdown
func StartHTTPServer(
	router *gin.Engine,
	cfg ServerConfig,
	auditLogger AuditLogger,
) {
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	go func() {
		zap.L().Info("HTTP server running", zap.String("port", cfg.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Fatal("ListenAndServe error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	zap.L().Info("Shutdown signal received", zap.String("signal", sig.String()))

	// Audit log BEFORE shutdown
	auditLogger.Log(context.Background(), AuditLog{
		Action:  "SERVER_SHUTDOWN",
		Message: "Server is shutting down",
		Meta: map[string]any{
			"signal": sig.String(),
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		zap.L().Error("Forced shutdown", zap.Error(err))
	} else {
		zap.L().Info("Server exited gracefully")
	}
}
