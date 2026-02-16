package employee

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
	employees := r.Group("/employees")

	employees.Use(middleware.AuthMiddleware())

	{
		employees.GET("/", middleware.RBACAuthorize(rbacService, "employee", "read"), handler.GetAll)
		employees.GET("/:id", middleware.RBACAuthorize(rbacService, "employee", "read"), handler.GetById)
		employees.POST("/", middleware.RBACAuthorize(rbacService, "employee", "create"), handler.Create)
	}
}
