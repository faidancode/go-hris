package user

import (
	"go-hris/internal/middleware"
	"go-hris/internal/rbac"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RegisterRoutes(
	r *gin.RouterGroup,
	handler *Handler,
	rbacService rbac.Service,
	logger *zap.Logger,
) {
	users := r.Group("/users")
	users.Use(middleware.AuthMiddleware())
	users.Use(middleware.ContextLogger(logger))
	{
		users.GET("",
			middleware.RateLimitByUser(3, 10),
			middleware.RBACAuthorize(rbacService, "user", "read"),
			handler.GetAll,
		)

		users.GET("/:id",
			middleware.RateLimitByUser(3, 10),
			middleware.RBACAuthorize(rbacService, "user", "read"),
			handler.GetById,
		)

		users.POST("",
			middleware.RateLimitByUser(0.1, 1),
			middleware.RBACAuthorize(rbacService, "user", "create"),
			handler.Create,
		)

		users.PATCH("/:id/status",
			middleware.RateLimitByUser(0.5, 2),
			middleware.RBACAuthorize(rbacService, "user", "update"),
			handler.ToggleStatus,
		)

		// Self reset password (misalnya user reset miliknya sendiri)
		users.POST("/:id/reset-password",
			middleware.RateLimitByUser(0.5, 2),
			middleware.RBACAuthorize(rbacService, "user", "update"),
			handler.ResetPassword,
		)

		// Admin force reset password
		users.POST("/:id/force-reset-password",
			middleware.RateLimitByUser(0.5, 2),
			middleware.RBACAuthorize(rbacService, "user", "update"),
			handler.ForceResetPassword,
		)
	}
}
