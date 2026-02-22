package payroll

import (
	"go-hris/internal/middleware"
	"go-hris/internal/rbac"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RegisterRoutes(
	r *gin.RouterGroup,
	handler *Handler,
	rbacService rbac.Service,
	rdb ...*redis.Client,
) {
	var redisClient *redis.Client
	if len(rdb) > 0 {
		redisClient = rdb[0]
	}

	payrolls := r.Group("/payrolls")
	payrolls.Use(middleware.AuthMiddleware())
	{
		payrolls.GET("",
			middleware.RateLimitByUser(2, 5),
			middleware.RBACAuthorize(rbacService, "payroll", "read"),
			handler.GetAll,
		)
		payrolls.GET("/:id",
			middleware.RateLimitByUser(2, 5),
			middleware.RBACAuthorize(rbacService, "payroll", "read"),
			handler.GetById,
		)
		payrolls.GET("/:id/breakdown",
			middleware.RateLimitByUser(2, 5),
			middleware.RBACAuthorize(rbacService, "payroll", "read"),
			handler.GetBreakdown,
		)
		payrolls.GET("/:id/payslip/download",
			middleware.RateLimitByUser(0.5, 1), // Ketat karena proses generate PDF berat
			middleware.RBACAuthorize(rbacService, "payroll", "read"),
			handler.DownloadPayslip,
		)

		// POST Create dengan Idempotency
		createMiddleware := []gin.HandlerFunc{
			middleware.RateLimitByUser(0.1, 1),
			middleware.RBACAuthorize(rbacService, "payroll", "create"),
		}
		if redisClient != nil {
			createMiddleware = append([]gin.HandlerFunc{middleware.Idempotency(redisClient)}, createMiddleware...)
		}
		payrolls.POST("", append(createMiddleware, handler.Create)...)

		payrolls.POST("/:id/regenerate",
			middleware.RateLimitByUser(0.1, 1),
			middleware.RBACAuthorize(rbacService, "payroll", "create"),
			handler.Regenerate,
		)
		payrolls.POST("/:id/approve",
			middleware.RateLimitByUser(0.2, 1),
			middleware.RBACAuthorize(rbacService, "payroll", "approve"),
			handler.Approve,
		)
		payrolls.POST("/:id/mark-paid",
			middleware.RateLimitByUser(0.1, 1),
			middleware.RBACAuthorize(rbacService, "payroll", "pay"),
			handler.MarkAsPaid,
		)
		payrolls.DELETE("/:id",
			middleware.RateLimitByUser(0.05, 1),
			middleware.RBACAuthorize(rbacService, "payroll", "delete"),
			handler.Delete,
		)
	}
}
