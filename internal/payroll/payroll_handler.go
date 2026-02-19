package payroll

import (
	"encoding/json"
	payrollerrors "go-hris/internal/payroll/errors"
	"go-hris/internal/shared/apperror"
	"go-hris/internal/shared/response"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Handler struct {
	service Service
	rdb     *redis.Client
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func NewHandlerWithRedis(service Service, rdb *redis.Client) *Handler {
	return &Handler{service: service, rdb: rdb}
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
	response.Error(c, httpErr.Status, httpErr.Code, httpErr.Message, httpErr.Details)
}

func (h *Handler) Create(c *gin.Context) {
	lockKey, _ := c.Get("idempotency_lock_key")
	cacheKey, _ := c.Get("idempotency_cache_key")

	if h.rdb != nil {
		if lk, ok := lockKey.(string); ok && lk != "" {
			defer h.rdb.Del(c.Request.Context(), lk)
		}
	}

	companyID := c.GetString("company_id")
	actorID := getActorID(c)

	var req CreatePayrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Input tidak valid", err.Error())
		return
	}

	resp, err := h.service.Create(c.Request.Context(), companyID, actorID, req)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}

	if h.rdb != nil {
		if ck, ok := cacheKey.(string); ok && ck != "" {
			if payload, marshalErr := json.Marshal(resp); marshalErr == nil {
				_ = h.rdb.Set(c.Request.Context(), ck, payload, 24*time.Hour).Err()
			}
		}
	}

	response.Success(c, http.StatusCreated, resp, nil)
}

func (h *Handler) GetAll(c *gin.Context) {
	ctx := c.Request.Context()
	companyID := c.GetString("company_id")
	var filterReq GetPayrollsFilterRequest
	if err := c.ShouldBindQuery(&filterReq); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Input tidak valid", err.Error())
		return
	}

	resp, err := h.service.GetAll(ctx, companyID, filterReq)
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

func (h *Handler) GetBreakdown(c *gin.Context) {
	ctx := c.Request.Context()
	targetID := c.Param("id")
	companyID := c.GetString("company_id")

	resp, err := h.service.GetBreakdown(ctx, companyID, targetID)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, resp, nil)
}

func (h *Handler) DownloadPayslip(c *gin.Context) {
	ctx := c.Request.Context()
	targetID := c.Param("id")
	companyID := c.GetString("company_id")

	resp, err := h.service.GetByID(ctx, companyID, targetID)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	if resp.PayslipURL == nil || *resp.PayslipURL == "" {
		h.writeServiceError(c, payrollerrors.ErrPayslipNotGenerated)
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, *resp.PayslipURL)
}

func (h *Handler) Regenerate(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	companyID := c.GetString("company_id")
	actorID := getActorID(c)

	var req RegeneratePayrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Input tidak valid", err.Error())
		return
	}

	resp, err := h.service.Regenerate(ctx, companyID, actorID, id, req)
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

func (h *Handler) MarkAsPaid(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	companyID := c.GetString("company_id")
	actorID := getActorID(c)

	resp, err := h.service.MarkAsPaid(ctx, companyID, actorID, id)
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
