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
		salaries.GET("",
			middleware.RateLimitByUser(1, 5),
			middleware.RBACAuthorize(rbacService, "salary", "read"),
			handler.GetAll,
		)
		salaries.GET("/:id",
			middleware.RateLimitByUser(2, 5),
			middleware.RBACAuthorize(rbacService, "salary", "read"),
			handler.GetById,
		)
		salaries.POST("",
			middleware.RateLimitByUser(0.1, 1),
			middleware.RBACAuthorize(rbacService, "salary", "update"),
			handler.Create,
		)
		salaries.PUT("/:id",
			middleware.RateLimitByUser(0.1, 1),
			middleware.RBACAuthorize(rbacService, "salary", "update"),
			handler.Update,
		)
		salaries.DELETE("/:id",
			middleware.RateLimitByUser(0.05, 1),
			middleware.RBACAuthorize(rbacService, "salary", "update"),
			handler.Delete,
		)
	}
}
