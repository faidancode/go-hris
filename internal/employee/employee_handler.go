package employee

import (
	"errors"
	autherrors "go-hris/internal/auth/errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateEmployeeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) GetAll(c *gin.Context) {
	ctx := c.Request.Context()
	companyID := c.Query("company_id")
	requesterID := c.GetString("employee_id") // Dari JWT middleware

	hasReadAll := c.GetBool("has_read_all")

	// Panggil service dengan flag hasReadAll
	resp, err := h.service.GetAll(ctx, companyID, requesterID, hasReadAll)
	if err != nil {
		if errors.Is(err, autherrors.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetById(c *gin.Context) {
	ctx := c.Request.Context()
	targetID := c.Param("id")
	employeeID := c.GetString("employee_id")

	// Enforce permission
	hasReadAll := c.GetBool("has_read_all")

	resp, err := h.service.GetByID(ctx, employeeID, targetID, hasReadAll)
	if err != nil {
		if errors.Is(err, autherrors.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
