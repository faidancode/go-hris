package company

import (
	"go-hris/internal/shared/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) GetMe(c *gin.Context) {
	companyID, ok := c.Get("company_id")
	if !ok {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Company ID not found in context", nil)
		return
	}

	comp, err := h.service.GetByID(c.Request.Context(), companyID.(string))
	if err != nil {
		response.Error(c, http.StatusNotFound, "NOT_FOUND", "Company not found", err.Error())
		return
	}

	response.Success(c, http.StatusOK, comp, nil)
}

func (h *Handler) UpdateMe(c *gin.Context) {
	companyID, ok := c.Get("company_id")
	if !ok {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Company ID not found in context", nil)
		return
	}

	var req UpdateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid input", err.Error())
		return
	}

	comp, err := h.service.Update(c.Request.Context(), companyID.(string), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update company", err.Error())
		return
	}

	response.Success(c, http.StatusOK, comp, nil)
}
