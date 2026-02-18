package rbac

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Enforce(c *gin.Context) {
	var req EnforceRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.EmployeeID = strings.TrimSpace(req.EmployeeID)
	req.CompanyID = strings.TrimSpace(req.CompanyID)
	req.Resource = strings.TrimSpace(req.Resource)
	req.Action = strings.TrimSpace(req.Action)

	if req.EmployeeID == "" || req.CompanyID == "" || req.Resource == "" || req.Action == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "employee_id, company_id, resource, and action are required",
		})
		return
	}

	log.Printf("enforce req: %+v", req)

	allowed, err := h.service.Enforce(req)
	if err != nil {
		log.Println("error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, EnforceResponse{
		Allowed: allowed,
	})
}
