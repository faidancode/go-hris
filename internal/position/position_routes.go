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
		// Akses baca untuk dropdown di form karyawan atau struktur organisasi
		positions.GET("",
			middleware.RateLimitByUser(5, 10),
			middleware.RBACAuthorize(rbacService, "position", "read"),
			h.GetAll,
		)
		positions.GET("/:id",
			middleware.RateLimitByUser(5, 10),
			middleware.RBACAuthorize(rbacService, "position", "read"),
			h.GetById,
		)

		// Mutasi data jabatan (jarang dilakukan, hanya saat restrukturisasi)
		positions.POST("",
			middleware.RateLimitByUser(0.2, 1),
			middleware.RBACAuthorize(rbacService, "position", "create"),
			h.Create,
		)
		positions.PUT("/:id",
			middleware.RateLimitByUser(0.2, 1),
			middleware.RBACAuthorize(rbacService, "position", "update"),
			h.Update,
		)
		positions.DELETE("/:id",
			middleware.RateLimitByUser(0.1, 1),
			middleware.RBACAuthorize(rbacService, "position", "delete"),
			h.Delete,
		)
	}
}
