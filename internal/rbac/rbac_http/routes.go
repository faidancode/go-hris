package rbac_http

import (
	"go-hris/internal/middleware"
	"go-hris/internal/rbac"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, handler *rbac.Handler, service rbac.Service) {
	group := r.Group("/rbac")
	group.Use(middleware.AuthMiddleware())
	{
		group.POST("/enforce", handler.Enforce)

		// Management
		group.GET("/roles", middleware.RBACAuthorize(service, "role", "read"), handler.ListRoles)
		group.GET("/roles/:id", middleware.RBACAuthorize(service, "role", "read"), handler.GetRole)
		group.POST("/roles", middleware.RBACAuthorize(service, "role", "manage"), handler.CreateRole)
		group.PUT("/roles/:id", middleware.RBACAuthorize(service, "role", "manage"), handler.UpdateRole)
		group.DELETE("/roles/:id", middleware.RBACAuthorize(service, "role", "manage"), handler.DeleteRole)

		group.GET("/permissions", middleware.RBACAuthorize(service, "role", "manage"), handler.ListPermissions)
	}
}
