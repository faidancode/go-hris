package company

import (
	"go-hris/internal/middleware"
	"go-hris/internal/rbac"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, handler *Handler, rbacService rbac.Service) {
	company := r.Group("/companies")
	company.Use(middleware.AuthMiddleware())
	{
		company.GET("/me", handler.GetMe)
		company.PUT("/me", middleware.RBACAuthorize(rbacService, "company", "update"), handler.UpdateMe)
		company.POST("/:id/registrations", handler.UpsertRegistration)
		company.GET("/:id/registrations", handler.ListRegistrations)
		company.DELETE("/:id/registrations/:type", handler.DeleteRegistration)
	}
}
