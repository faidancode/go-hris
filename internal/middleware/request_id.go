package middleware

import (
	"go-hris/internal/shared/contextutil"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-ID")
		if rid == "" {
			rid = uuid.New().String()
		}

		// Set di Gin Context
		c.Set("request_id", rid)

		// Propagasi ke Standard Context menggunakan helper dari shared
		ctx := contextutil.WithRequestID(c.Request.Context(), rid)
		c.Request = c.Request.WithContext(ctx)

		c.Header("X-Request-ID", rid)
		c.Next()
	}
}
