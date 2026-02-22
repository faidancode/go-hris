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
		// Proteksi Resource, rate limit sebelum RBAC
		// Melihat riwayat (Lebih fleksibel)
		attendances.GET("",
			middleware.RateLimitByUser(2, 10),
			middleware.RBACAuthorize(rbacService, "attendance", "read"),
			h.GetAll,
		)

		// Clock-in (Ketat: Mencegah double tap/spam)
		attendances.POST("/clock-in",
			middleware.RateLimitByUser(0.2, 1),
			middleware.RBACAuthorize(rbacService, "attendance", "create"),
			h.ClockIn,
		)

		// Clock-out (Ketat: Mencegah double tap/spam)
		attendances.POST("/clock-out",
			middleware.RateLimitByUser(0.2, 1),
			middleware.RBACAuthorize(rbacService, "attendance", "create"),
			h.ClockOut,
		)
	}
}
