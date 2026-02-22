package company

import (
	companyerrors "go-hris/internal/company/errors"
	"go-hris/internal/shared/apperror"
	"go-hris/internal/shared/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	service Service
	logger  *zap.Logger
}

func NewHandler(service Service, logger ...*zap.Logger) *Handler {
	l := zap.L().Named("company.handler")
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0].Named("company.handler")
	}
	return &Handler{service: service, logger: l}
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

func (h *Handler) UpsertRegistration(c *gin.Context) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		err := apperror.New(
			apperror.CodeUnauthorized,
			"Unauthorized",
			http.StatusUnauthorized,
		)
		httpErr := apperror.ToHTTP(err)
		c.JSON(httpErr.Status, httpErr)
		return
	}

	var req UpsertCompanyRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := apperror.ToHTTP(apperror.MapValidationError(err))
		c.JSON(httpErr.Status, httpErr)
		return
	}

	if err := h.service.UpsertRegistration(c.Request.Context(), companyID, req); err != nil {
		httpErr := apperror.ToHTTP(err)
		c.JSON(httpErr.Status, httpErr)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) DeleteRegistration(c *gin.Context) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		httpErr := apperror.ToHTTP(apperror.New(apperror.CodeUnauthorized, "Unauthorized", http.StatusUnauthorized))
		c.JSON(httpErr.Status, httpErr)
		return
	}
	typeParam := c.Param("type")
	if typeParam == "" {
		// Gunakan error yang sudah ada atau buat baru
		httpErr := apperror.ToHTTP(companyerrors.ErrInvalidRegistrationType)
		c.JSON(httpErr.Status, httpErr)
		return
	}
	regType := RegistrationType(typeParam)

	if err := h.service.DeleteRegistration(c.Request.Context(), companyID, regType); err != nil {
		h.logger.Error("failed to delete registration", zap.Error(err))

		httpErr := apperror.ToHTTP(err)
		c.JSON(httpErr.Status, httpErr)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) ListRegistrations(c *gin.Context) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		err := apperror.New(
			apperror.CodeUnauthorized,
			"Unauthorized",
			http.StatusUnauthorized,
		)
		httpErr := apperror.ToHTTP(err)
		c.JSON(httpErr.Status, httpErr)
		return
	}

	result, err := h.service.ListRegistrations(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
