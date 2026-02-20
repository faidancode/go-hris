package middleware

import (
	"go-hris/internal/domain"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ContextKey string

const (
	ContextEmployeeID ContextKey = "employee_id"
	ContextCompanyID  ContextKey = "company_id"
)

// RBACService adalah interface lokal.
// Apapun package yang punya method Enforce(domain.EnforceRequest) bisa masuk ke sini.
type RBACService interface {
	Enforce(req domain.EnforceRequest) (bool, error)
}

func RBACAuthorize(service RBACService, resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		employeeID, ok1 := c.Get(string(ContextEmployeeID))
		companyID, ok2 := c.Get(string(ContextCompanyID))

		if !ok1 || !ok2 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth context"})
			c.Abort()
			return
		}

		// Menggunakan struct domain
		req := domain.EnforceRequest{
			EmployeeID: employeeID.(string),
			CompanyID:  companyID.(string),
			Resource:   resource,
			Action:     action,
		}

		allowed, err := service.Enforce(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error":    "forbidden",
				"message":  "You do not have permission to access this resource",
				"required": resource + ":" + action,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
