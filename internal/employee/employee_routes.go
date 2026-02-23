package employee

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
	employees := r.Group("/employees")
	employees.Use(middleware.AuthMiddleware())
	employees.Use(middleware.ContextLogger(logger))
	{
		employees.GET("",
			middleware.RateLimitByUser(3, 10),
			middleware.RBACAuthorize(rbacService, "employee", "read"),
			handler.GetAll,
		)

		employees.GET("/options",
			middleware.RateLimitByUser(5, 20), // Limit sedikit lebih longgar karena ringan
			middleware.RBACAuthorize(rbacService, "employee", "read"),
			handler.GetOptions,
		)

		employees.GET("/:id",
			middleware.RateLimitByUser(3, 10),
			middleware.RBACAuthorize(rbacService, "employee", "read"),
			handler.GetById,
		)

		employees.POST("",
			middleware.RateLimitByUser(0.1, 1),
			middleware.RBACAuthorize(rbacService, "employee", "create"),
			handler.Create,
		)

		employees.PUT("/:id",
			middleware.RateLimitByUser(0.5, 2),
			middleware.RBACAuthorize(rbacService, "employee", "update"),
			handler.Update,
		)

		employees.DELETE("/:id",
			middleware.RateLimitByUser(0.05, 1),
			middleware.RBACAuthorize(rbacService, "employee", "delete"),
			handler.Delete,
		)
	}
}
