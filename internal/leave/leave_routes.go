package leave

import (
	"go-hris/internal/middleware"
	"go-hris/internal/rbac"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	r *gin.RouterGroup,
	handler *Handler,
	rbacService rbac.Service,
) {
	leaves := r.Group("/leaves")
	leaves.Use(middleware.AuthMiddleware())
	{
		leaves.GET("",
			middleware.RateLimitByUser(3, 10),
			middleware.RBACAuthorize(rbacService, "leave", "read"),
			handler.GetAll,
		)
		leaves.GET("/:id",
			middleware.RateLimitByUser(3, 10),
			middleware.RBACAuthorize(rbacService, "leave", "read"),
			handler.GetById,
		)
		// Prevent double-tap
		leaves.POST("",
			middleware.RateLimitByUser(0.5, 2),
			middleware.RBACAuthorize(rbacService, "leave", "create"),
			handler.Create,
		)
		leaves.PUT("/:id",
			middleware.RateLimitByUser(0.5, 2),
			middleware.RBACAuthorize(rbacService, "leave", "create"),
			handler.Update,
		)
		leaves.POST("/:id/submit",
			middleware.RateLimitByUser(0.2, 1),
			middleware.RBACAuthorize(rbacService, "leave", "create"),
			handler.Submit,
		)
		leaves.POST("/:id/approve",
			middleware.RateLimitByUser(0.5, 1),
			middleware.RBACAuthorize(rbacService, "leave", "approve"),
			handler.Approve,
		)
		leaves.POST("/:id/reject",
			middleware.RateLimitByUser(0.5, 1),
			middleware.RBACAuthorize(rbacService, "leave", "approve"),
			handler.Reject,
		)
		leaves.DELETE("/:id",
			middleware.RateLimitByUser(0.1, 1),
			middleware.RBACAuthorize(rbacService, "leave", "delete"),
			handler.Delete,
		)
	}
}
