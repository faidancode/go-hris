package attendance

import (
	"go-hris/internal/middleware"
	"go-hris/internal/rbac"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, h *Handler, rbacService rbac.Service) {
	attendances := r.Group("/attendances")
	attendances.Use(middleware.AuthMiddleware())
	{
		attendances.GET("", middleware.RBACAuthorize(rbacService, "attendance", "read"), h.GetAll)
		attendances.POST("/clock-in", middleware.RBACAuthorize(rbacService, "attendance", "create"), h.ClockIn)
		attendances.POST("/clock-out", middleware.RBACAuthorize(rbacService, "attendance", "create"), h.ClockOut)
	}
}
