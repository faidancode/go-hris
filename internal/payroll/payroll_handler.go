package payroll

import (
	"encoding/json"
	"errors"
	"go-hris/internal/shared/response"
	"net/http"
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

func (h *Handler) Create(c *gin.Context) {
	lockKey, _ := c.Get("idempotency_lock_key")
	cacheKey, _ := c.Get("idempotency_cache_key")

	if h.rdb != nil {
		if lk, ok := lockKey.(string); ok && lk != "" {
			defer h.rdb.Del(c.Request.Context(), lk)
		}
	}

	companyID := c.GetString("company_id")
	actorID := c.GetString("employee_id")
	if actorID == "" {
		actorID = c.GetString("user_id_validated")
	}

	var req CreatePayrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Input tidak valid", err.Error())
		return
	}

	resp, err := h.service.Create(c.Request.Context(), companyID, actorID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
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

	resp, err := h.service.GetAll(ctx, companyID)
	if err != nil {
		if errors.Is(err, errors.New("forbidden")) {
			response.Error(c, http.StatusForbidden, "FORBIDDEN", "forbidden", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	response.Success(c, http.StatusOK, resp, nil)
}

func (h *Handler) GetById(c *gin.Context) {
	ctx := c.Request.Context()
	targetID := c.Param("id")
	companyID := c.GetString("company_id")

	resp, err := h.service.GetByID(ctx, companyID, targetID)
	if err != nil {
		if errors.Is(err, errors.New("forbidden")) {
			response.Error(c, http.StatusForbidden, "FORBIDDEN", "forbidden", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	response.Success(c, http.StatusOK, resp, nil)
}

func (h *Handler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	companyID := c.GetString("company_id")
	actorID := c.GetString("employee_id")
	if actorID == "" {
		actorID = c.GetString("user_id_validated")
	}

	var req UpdatePayrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Input tidak valid", err.Error())
		return
	}

	resp, err := h.service.Update(ctx, companyID, actorID, id, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	response.Success(c, http.StatusOK, resp, nil)
}

func (h *Handler) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	companyID := c.GetString("company_id")

	if err := h.service.Delete(ctx, companyID, id); err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"deleted": true}, nil)
}
