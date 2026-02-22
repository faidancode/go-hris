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
		// 1. GetAll & GetById (Biasanya dipanggil saat isi form Employee/Filter)
		// Rate: 5 req/detik, Burst: 10 (Longgar karena datanya kecil)
		departments.GET("",
			middleware.RateLimitByUser(5, 10),
			middleware.RBACAuthorize(rbacService, "department", "read"),
			h.GetAll,
		)
		departments.GET("/:id",
			middleware.RateLimitByUser(5, 10),
			middleware.RBACAuthorize(rbacService, "department", "read"),
			h.GetById,
		)

		// 2. Create, Update, Delete (Hanya Admin HR saat restrukturisasi)
		// Rate: 0.1 req/detik (1x per 10 detik), Burst: 1
		departments.POST("",
			middleware.RateLimitByUser(0.1, 1),
			middleware.RBACAuthorize(rbacService, "department", "create"),
			h.Create,
		)
		departments.PUT("/:id",
			middleware.RateLimitByUser(0.1, 1),
			middleware.RBACAuthorize(rbacService, "department", "update"),
			h.Update,
		)
		departments.DELETE("/:id",
			middleware.RateLimitByUser(0.1, 1),
			middleware.RBACAuthorize(rbacService, "department", "delete"),
			h.Delete,
		)
	}
}
