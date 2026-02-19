package employee

import (
	"errors"
	"go-hris/internal/shared/response"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	service Service
	logger  *zap.Logger
}

func NewHandler(service Service, logger ...*zap.Logger) *Handler {
	l := zap.L().Named("employee.handler")
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0].Named("employee.handler")
	}
	return &Handler{service: service, logger: l}
}

func (h *Handler) Create(c *gin.Context) {
	companyID := c.GetString("company_id")
	h.logger.Debug("http create employee", zap.String("company_id", companyID))
	var req CreateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("http create employee validation failed", zap.Error(err))
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Input tidak valid", err.Error())
		return
	}

	resp, err := h.service.Create(c.Request.Context(), companyID, req)
	if err != nil {
		h.logger.Error("http create employee failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	response.Success(c, http.StatusCreated, resp, nil)
}

func (h *Handler) GetAll(c *gin.Context) {
	ctx := c.Request.Context()
	companyID := c.GetString("company_id")
	h.logger.Debug("http get all employees", zap.String("company_id", companyID))

	resp, err := h.service.GetAll(ctx, companyID)
	if err != nil {
		h.logger.Error("http get all employees failed", zap.Error(err))
		if errors.Is(err, errors.New("forbidden")) {
			response.Error(c, http.StatusForbidden, "FORBIDDEN", "forbidden", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	q := strings.TrimSpace(strings.ToLower(c.Query("q")))
	if q != "" {
		filtered := make([]EmployeeResponse, 0, len(resp))
		for _, e := range resp {
			if strings.Contains(strings.ToLower(e.FullName), q) || strings.Contains(strings.ToLower(e.Email), q) {
				filtered = append(filtered, e)
			}
		}
		resp = filtered
	}

	sortBy := strings.ToLower(strings.TrimSpace(c.DefaultQuery("sort_by", "name")))
	sortDir := strings.ToLower(strings.TrimSpace(c.DefaultQuery("sort_dir", "asc")))
	if sortDir != "desc" {
		sortDir = "asc"
	}
	sort.Slice(resp, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "email":
			less = strings.ToLower(resp[i].Email) < strings.ToLower(resp[j].Email)
		case "id":
			less = resp[i].ID < resp[j].ID
		default:
			less = strings.ToLower(resp[i].FullName) < strings.ToLower(resp[j].FullName)
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
	ctx := c.Request.Context()
	targetID := c.Param("id")
	companyID := c.GetString("company_id")
	h.logger.Debug("http get employee by id",
		zap.String("company_id", companyID),
		zap.String("employee_id", targetID),
	)

	resp, err := h.service.GetByID(ctx, companyID, targetID)
	if err != nil {
		h.logger.Error("http get employee by id failed", zap.Error(err))
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
	h.logger.Debug("http update employee",
		zap.String("company_id", companyID),
		zap.String("employee_id", id),
	)
	var req UpdateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("http update employee validation failed", zap.Error(err))
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Input tidak valid", err.Error())
		return
	}

	resp, err := h.service.Update(ctx, companyID, id, req)
	if err != nil {
		h.logger.Error("http update employee failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	response.Success(c, http.StatusOK, resp, nil)
}

func (h *Handler) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	companyID := c.GetString("company_id")
	h.logger.Debug("http delete employee",
		zap.String("company_id", companyID),
		zap.String("employee_id", id),
	)

	if err := h.service.Delete(ctx, companyID, id); err != nil {
		h.logger.Error("http delete employee failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"deleted": true}, nil)
}
