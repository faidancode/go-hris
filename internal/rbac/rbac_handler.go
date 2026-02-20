package rbac

import (
	"go-hris/internal/domain"
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
	var req domain.EnforceRequest

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

	response.Success(c, http.StatusOK, domain.EnforceResponse{
		Allowed: allowed,
	}, nil)
}

func (h *Handler) ListRoles(c *gin.Context) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Missing company context", nil)
		return
	}

	roles, err := h.service.ListRoles(companyID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	response.Success(c, http.StatusOK, roles, nil)
}

func (h *Handler) GetRole(c *gin.Context) {
	id := c.Param("id")
	role, err := h.service.GetRole(id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "NOT_FOUND", "Role tidak ditemukan", nil)
		return
	}

	response.Success(c, http.StatusOK, role, nil)
}

func (h *Handler) CreateRole(c *gin.Context) {
	companyID := c.GetString("company_id")
	var req domain.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	if err := h.service.CreateRole(companyID, req); err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	response.Success(c, http.StatusCreated, nil, nil)
}

func (h *Handler) UpdateRole(c *gin.Context) {
	id := c.Param("id")
	var req domain.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	if err := h.service.UpdateRole(id, req); err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	response.Success(c, http.StatusOK, nil, nil)
}

func (h *Handler) DeleteRole(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteRole(id); err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	response.Success(c, http.StatusOK, nil, nil)
}

func (h *Handler) ListPermissions(c *gin.Context) {
	perms, err := h.service.ListPermissions()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	response.Success(c, http.StatusOK, perms, nil)
}
