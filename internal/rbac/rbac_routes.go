package rbac

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine, handler *Handler) {
	group := r.Group("/rbac")
	{
		group.POST("/enforce", handler.Enforce)
	}
}
