package rbac

import (
	"go-hris/internal/shared/response"
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
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Input tidak valid", err.Error())
		return
	}

	req.EmployeeID = strings.TrimSpace(req.EmployeeID)
	req.CompanyID = strings.TrimSpace(req.CompanyID)
	req.Resource = strings.TrimSpace(req.Resource)
	req.Action = strings.TrimSpace(req.Action)

	if req.EmployeeID == "" || req.CompanyID == "" || req.Resource == "" || req.Action == "" {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "employee_id, company_id, resource, and action are required", nil)
		return
	}

	log.Printf("enforce req: %+v", req)

	allowed, err := h.service.Enforce(req)
	if err != nil {
		log.Println("error", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	response.Success(c, http.StatusOK, EnforceResponse{
		Allowed: allowed,
	}, nil)
}
