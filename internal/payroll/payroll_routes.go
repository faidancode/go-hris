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
		payrolls.GET("", middleware.RBACAuthorize(rbacService, "payroll", "read"), handler.GetAll)
		payrolls.GET("/:id", middleware.RBACAuthorize(rbacService, "payroll", "read"), handler.GetById)
		if redisClient != nil {
			payrolls.POST(
				"",
				middleware.Idempotency(redisClient),
				middleware.RBACAuthorize(rbacService, "payroll", "approve"),
				handler.Create,
			)
		} else {
			payrolls.POST("", middleware.RBACAuthorize(rbacService, "payroll", "approve"), handler.Create)
		}
		payrolls.PUT("/:id", middleware.RBACAuthorize(rbacService, "payroll", "approve"), handler.Update)
		payrolls.DELETE("/:id", middleware.RBACAuthorize(rbacService, "payroll", "approve"), handler.Delete)
	}
}
