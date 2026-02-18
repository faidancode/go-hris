package attendance

import (
	"go-hris/internal/shared/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ClockIn(c *gin.Context) {
	companyID := c.GetString("company_id")
	employeeID := c.GetString("employee_id")
	if employeeID == "" {
		employeeID = c.GetString("user_id_validated")
	}

	var req ClockInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Input tidak valid", err.Error())
		return
	}

	resp, err := h.service.ClockIn(c.Request.Context(), companyID, employeeID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	response.Success(c, http.StatusCreated, resp, nil)
}

func (h *Handler) ClockOut(c *gin.Context) {
	companyID := c.GetString("company_id")
	employeeID := c.GetString("employee_id")
	if employeeID == "" {
		employeeID = c.GetString("user_id_validated")
	}

	var req ClockOutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Input tidak valid", err.Error())
		return
	}

	resp, err := h.service.ClockOut(c.Request.Context(), companyID, employeeID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	response.Success(c, http.StatusOK, resp, nil)
}

func (h *Handler) GetAll(c *gin.Context) {
	companyID := c.GetString("company_id")
	resp, err := h.service.GetAll(c.Request.Context(), companyID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if pageSize < 1 {
		pageSize = 10
	}

	total := int64(len(resp))
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > len(resp) {
		start = len(resp)
	}
	if end > len(resp) {
		end = len(resp)
	}

	meta := response.NewPaginationMeta(total, page, pageSize)
	response.Success(c, http.StatusOK, resp[start:end], &meta)
}
