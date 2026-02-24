package user

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	"go-hris/internal/shared/apperror"
	"go-hris/internal/shared/contextutil"
	"go-hris/internal/shared/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	svc    Service
	logger *zap.Logger
}

func NewHandler(service Service, logger ...*zap.Logger) *Handler {
	l := zap.L().Named("employee.handler")
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0].Named("employee.handler")
	}
	return &Handler{svc: service, logger: l}
}

func writeError(c *gin.Context, err error) {
	if appErr, ok := err.(*apperror.AppError); ok {
		c.JSON(appErr.HTTPStatus, appErr)
		return
	}

	c.JSON(http.StatusInternalServerError, apperror.New(
		apperror.CodeInternalError,
		"Internal server error",
		http.StatusInternalServerError,
	))
}

func (h *Handler) GetAll(c *gin.Context) {
	ctx := c.Request.Context()
	companyID := c.GetString("company_id")
	h.logger.Debug("http get all users", zap.String("company_id", companyID))

	resp, err := h.svc.GetAll(ctx, companyID)
	if err != nil {
		writeError(c, err)
		return
	}

	q := strings.TrimSpace(strings.ToLower(c.Query("q")))
	if q != "" {
		filtered := make([]UserResponse, 0, len(resp))
		for _, u := range resp {
			if strings.Contains(strings.ToLower(u.Email), q) {
				filtered = append(filtered, u)
			}
		}
		resp = filtered
	}

	sortBy := strings.ToLower(strings.TrimSpace(c.DefaultQuery("sort_by", "email")))
	sortDir := strings.ToLower(strings.TrimSpace(c.DefaultQuery("sort_dir", "asc")))
	if sortDir != "desc" {
		sortDir = "asc"
	}

	sort.Slice(resp, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "id":
			less = resp[i].ID < resp[j].ID
		default:
			less = strings.ToLower(resp[i].Email) < strings.ToLower(resp[j].Email)
		}
		if sortDir == "desc" {
			return !less
		}
		return less
	})

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

func (h *Handler) GetById(c *gin.Context) {
	companyID := c.GetString("company_id")
	id := c.Param("id")

	ctx := contextutil.WithLogger(c.Request.Context(), h.logger)

	res, err := h.svc.GetByID(ctx, companyID, id)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *Handler) Create(c *gin.Context) {
	companyID := c.GetString("company_id")

	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apperror.New(
			apperror.CodeInvalidInput,
			err.Error(),
			http.StatusBadRequest,
		))
		return
	}

	ctx := contextutil.WithLogger(c.Request.Context(), h.logger)

	res, err := h.svc.Create(ctx, companyID, req)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, res)
}

func (h *Handler) GetCompanyUsers(c *gin.Context) {
	companyID := c.GetString("company_id")

	ctx := contextutil.WithLogger(c.Request.Context(), h.logger)

	res, err := h.svc.GetCompanyUsers(ctx, companyID)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *Handler) ToggleStatus(c *gin.Context) {
	companyID := c.GetString("company_id")
	id := c.Param("id")

	var body struct {
		IsActive *bool `json:"is_active" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, apperror.New(
			apperror.CodeInvalidInput,
			err.Error(),
			http.StatusBadRequest,
		))
		return
	}

	ctx := contextutil.WithLogger(c.Request.Context(), h.logger)

	if err := h.svc.ToggleStatus(ctx, companyID, id, *body.IsActive); err != nil {
		writeError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) ChangePassword(c *gin.Context) {
	companyID := c.GetString("company_id")
	id := c.Param("id")

	var body struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, apperror.New(
			apperror.CodeInvalidInput,
			err.Error(),
			http.StatusBadRequest,
		))
		return
	}

	ctx := contextutil.WithLogger(c.Request.Context(), h.logger)

	if err := h.svc.ChangePassword(ctx, companyID, id, body.CurrentPassword, body.NewPassword); err != nil {
		writeError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) ResetPassword(c *gin.Context) {
	companyID := c.GetString("company_id")
	id := c.Param("id")

	var body struct {
		NewPassword string `json:"new_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, apperror.New(
			apperror.CodeInvalidInput,
			err.Error(),
			http.StatusBadRequest,
		))
		return
	}

	ctx := contextutil.WithLogger(c.Request.Context(), h.logger)

	if err := h.svc.ResetPassword(ctx, companyID, id, body.NewPassword); err != nil {
		writeError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) ForceResetPassword(c *gin.Context) {
	companyID := c.GetString("company_id")
	id := c.Param("id")

	var body struct {
		NewPassword string `json:"new_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, apperror.New(
			apperror.CodeInvalidInput,
			err.Error(),
			http.StatusBadRequest,
		))
		return
	}

	ctx := contextutil.WithLogger(c.Request.Context(), h.logger)

	if err := h.svc.ForceResetPassword(ctx, companyID, id, body.NewPassword); err != nil {
		writeError(c, err)
		return
	}

	c.Status(http.StatusOK)
}
