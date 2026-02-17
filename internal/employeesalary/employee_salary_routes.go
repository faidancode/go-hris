package employeesalary

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
	salaries := r.Group("/employee-salaries")
	salaries.Use(middleware.AuthMiddleware())
	{
		salaries.GET("", middleware.RBACAuthorize(rbacService, "salary", "read"), handler.GetAll)
		salaries.GET("/:id", middleware.RBACAuthorize(rbacService, "salary", "read"), handler.GetById)
		salaries.POST("", middleware.RBACAuthorize(rbacService, "salary", "update"), handler.Create)
		salaries.PUT("/:id", middleware.RBACAuthorize(rbacService, "salary", "update"), handler.Update)
		salaries.DELETE("/:id", middleware.RBACAuthorize(rbacService, "salary", "update"), handler.Delete)
	}
}
