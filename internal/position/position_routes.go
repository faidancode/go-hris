// internal/position/delivery/http/routes.go

package position

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
	positions := r.Group("/positions")

	positions.Use(middleware.AuthMiddleware())

	{
		positions.GET("", middleware.RBACAuthorize(rbacService, "position", "read"), h.GetAll)
		positions.POST("", middleware.RBACAuthorize(rbacService, "position", "create"), h.Create)
		positions.GET("/:id", middleware.RBACAuthorize(rbacService, "position", "read"), h.GetById)
		positions.PUT("/:id", middleware.RBACAuthorize(rbacService, "position", "update"), h.Update)
		positions.DELETE("/:id", middleware.RBACAuthorize(rbacService, "position", "delete"), h.Delete)
	}
}
