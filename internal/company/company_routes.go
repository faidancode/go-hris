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
		// 1. Get My Company (Paling sering dipanggil oleh dashboard/profile)
		// Rate: 2 req/detik, Burst: 10 (Aman untuk refresh berkali-kali)
		company.GET("/me",
			middleware.RateLimitByUser(2, 10),
			handler.GetMe,
		)

		// 2. Update My Company (Jarang dilakukan)
		// Rate: 0.1 req/detik (1x per 10 detik), Burst: 1
		company.PUT("/me",
			middleware.RateLimitByUser(0.1, 1),
			middleware.RBACAuthorize(rbacService, "company", "update"),
			handler.UpdateMe,
		)

		// 3. Upsert Registration (Tindakan krusial/administratif)
		// Rate: 0.5 req/detik (1x per 2 detik), Burst: 1
		company.POST("/:id/registrations",
			middleware.RateLimitByUser(0.5, 1),
			middleware.RBACAuthorize(rbacService, "company", "update"),
			handler.UpsertRegistration,
		)

		// 4. List Registrations (Hanya Admin)
		// Rate: 1 req/detik, Burst: 5
		company.GET("/:id/registrations",
			middleware.RateLimitByUser(1, 5),
			middleware.RBACAuthorize(rbacService, "company", "read"),
			handler.ListRegistrations,
		)

		// 5. Delete Registration (Sangat krusial)
		// Rate: 0.1 req/detik, Burst: 1
		company.DELETE("/:id/registrations/:type",
			middleware.RateLimitByUser(0.1, 1),
			middleware.RBACAuthorize(rbacService, "company", "delete"),
			handler.DeleteRegistration,
		)
	}
}
