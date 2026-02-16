package rbac

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ContextKey string

const (
	ContextEmployeeID ContextKey = "employee_id"
	ContextCompanyID  ContextKey = "company_id"
)

// Middleware factory
func Authorize(service Service, resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Ambil dari context (biasanya di-set oleh JWT middleware)
		employeeID, ok1 := c.Get(string(ContextEmployeeID))
		companyID, ok2 := c.Get(string(ContextCompanyID))

		if !ok1 || !ok2 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "missing auth context",
			})
			c.Abort()
			return
		}

		req := EnforceRequest{
			EmployeeID: employeeID.(string),
			CompanyID:  companyID.(string),
			Resource:   resource,
			Action:     action,
		}

		allowed, err := service.Enforce(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "forbidden",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
