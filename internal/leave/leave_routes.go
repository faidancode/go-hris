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
		leaves.GET("", middleware.RBACAuthorize(rbacService, "leave", "read"), handler.GetAll)
		leaves.GET("/:id", middleware.RBACAuthorize(rbacService, "leave", "read"), handler.GetById)
		leaves.POST("", middleware.RBACAuthorize(rbacService, "leave", "create"), handler.Create)
		leaves.PUT("/:id", middleware.RBACAuthorize(rbacService, "leave", "create"), handler.Update)
		leaves.POST("/:id/submit", middleware.RBACAuthorize(rbacService, "leave", "create"), handler.Submit)
		leaves.POST("/:id/approve", middleware.RBACAuthorize(rbacService, "leave", "approve"), handler.Approve)
		leaves.POST("/:id/reject", middleware.RBACAuthorize(rbacService, "leave", "approve"), handler.Reject)
		leaves.DELETE("/:id", middleware.RBACAuthorize(rbacService, "leave", "approve"), handler.Delete)
	}
}
