package auth

import (
	"go-hris/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, handler *Handler) {
	auth := r.Group("/auth")
	{
		auth.GET("/me", middleware.AuthMiddleware(), middleware.RateLimitByUser(2, 5), handler.Me)
		auth.POST("/login", middleware.RateLimitByIP(0.08, 5), handler.Login)
		auth.POST("/refresh", handler.RefreshToken)
		auth.POST("/logout", middleware.RateLimitByUser(2, 5), handler.Logout)
		auth.POST("/register", middleware.RateLimitByUser(2, 5), handler.Register)
		auth.POST("/register", middleware.RateLimitByIP(0.1, 1), handler.Register)
	}
}
