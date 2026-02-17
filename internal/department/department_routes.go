// internal/department/delivery/http/routes.go

package department

import (
	"go-hris/internal/middleware"
	"go-hris/internal/rbac"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	r *gin.RouterGroup,
	h *Handler,
	rbacService rbac.Service,
) {
	departments := r.Group("/departments")

	departments.Use(middleware.AuthMiddleware())

	{
		departments.GET("", middleware.RBACAuthorize(rbacService, "department", "read"), h.GetAll)
		departments.POST("", middleware.RBACAuthorize(rbacService, "department", "create"), h.Create)
		departments.GET("/:id", middleware.RBACAuthorize(rbacService, "department", "read"), h.GetById)
		departments.PUT("/:id", middleware.RBACAuthorize(rbacService, "department", "update"), h.Update)
		departments.DELETE("/:id", middleware.RBACAuthorize(rbacService, "department", "delete"), h.Delete)
	}
}
