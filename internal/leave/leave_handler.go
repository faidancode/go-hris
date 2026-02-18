package leave

import (
	"go-hris/internal/shared/apperror"
	"go-hris/internal/shared/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	service Service
	logger  *zap.Logger
}

func NewHandler(service Service, logger ...*zap.Logger) *Handler {
	l := zap.L().Named("leave.handler")
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0].Named("leave.handler")
	}
	return &Handler{service: service, logger: l}
}

func getActorID(c *gin.Context) string {
	actorID := c.GetString("employee_id")
	if actorID == "" {
		actorID = c.GetString("user_id_validated")
	}
	return actorID
}

func (h *Handler) writeServiceError(c *gin.Context, err error) {
	httpErr := apperror.ToHTTP(err)
	h.logger.Warn("leave request failed",
		zap.String("method", c.Request.Method),
		zap.String("path", c.FullPath()),
		zap.Int("status", httpErr.Status),
		zap.String("code", httpErr.Code),
		zap.String("message", httpErr.Message),
	)
	response.Error(c, httpErr.Status, httpErr.Code, httpErr.Message, httpErr.Details)
}

func (h *Handler) Create(c *gin.Context) {
	companyID := c.GetString("company_id")
	actorID := getActorID(c)
	h.logger.Debug("http create leave", zap.String("company_id", companyID), zap.String("actor_id", actorID))

	var req CreateLeaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("http create leave validation failed", zap.Error(err))
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Input tidak valid", err.Error())
		return
	}

	resp, err := h.service.Create(c.Request.Context(), companyID, actorID, req)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, resp, nil)
}

func (h *Handler) GetAll(c *gin.Context) {
	ctx := c.Request.Context()
	companyID := c.GetString("company_id")

	resp, err := h.service.GetAll(ctx, companyID)
	if err != nil {
		h.writeServiceError(c, err)
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

func (h *Handler) GetById(c *gin.Context) {
	ctx := c.Request.Context()
	targetID := c.Param("id")
	companyID := c.GetString("company_id")

	resp, err := h.service.GetByID(ctx, companyID, targetID)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, resp, nil)
}

func (h *Handler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	companyID := c.GetString("company_id")
	actorID := getActorID(c)

	var req UpdateLeaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("http update leave validation failed", zap.Error(err))
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Input tidak valid", err.Error())
		return
	}

	resp, err := h.service.Update(ctx, companyID, actorID, id, req)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, resp, nil)
}

func (h *Handler) Submit(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	companyID := c.GetString("company_id")
	actorID := getActorID(c)

	resp, err := h.service.Submit(ctx, companyID, actorID, id)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, resp, nil)
}

func (h *Handler) Approve(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	companyID := c.GetString("company_id")
	actorID := getActorID(c)

	resp, err := h.service.Approve(ctx, companyID, actorID, id)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, resp, nil)
}

func (h *Handler) Reject(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	companyID := c.GetString("company_id")
	actorID := getActorID(c)

	var req RejectLeaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("http reject leave validation failed", zap.Error(err))
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Input tidak valid", err.Error())
		return
	}

	resp, err := h.service.Reject(ctx, companyID, actorID, id, req.RejectionReason)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, resp, nil)
}

func (h *Handler) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	companyID := c.GetString("company_id")

	if err := h.service.Delete(ctx, companyID, id); err != nil {
		h.writeServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"deleted": true}, nil)
}
