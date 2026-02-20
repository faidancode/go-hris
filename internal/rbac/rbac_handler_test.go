package rbac_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"go-hris/internal/domain"
	"go-hris/internal/rbac"
	rbacMock "go-hris/internal/rbac/mock"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type apiEnvelope struct {
	Ok   bool            `json:"ok"`
	Data json.RawMessage `json:"data"`
}

func TestHandler_Management(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := rbacMock.NewMockService(ctrl)
	handler := rbac.NewHandler(mockService)

	router := gin.Default()
	// Set company_id context as it's required by some handlers
	router.Use(func(c *gin.Context) {
		c.Set("company_id", "comp-1")
		c.Next()
	})

	router.GET("/api/v1/rbac/roles", handler.ListRoles)
	router.GET("/api/v1/rbac/roles/:id", handler.GetRole)
	router.POST("/api/v1/rbac/roles", handler.CreateRole)
	router.PUT("/api/v1/rbac/roles/:id", handler.UpdateRole)
	router.DELETE("/api/v1/rbac/roles/:id", handler.DeleteRole)
	router.GET("/api/v1/rbac/permissions", handler.ListPermissions)

	t.Run("ListRoles - Success", func(t *testing.T) {
		mockService.EXPECT().ListRoles("comp-1").Return([]domain.RoleResponse{
			{ID: "role-1", Name: "Admin"},
		}, nil)

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/rbac/roles", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("GetRole - Success", func(t *testing.T) {
		mockService.EXPECT().GetRole("role-1").Return(&domain.RoleResponse{ID: "role-1", Name: "Admin"}, nil)

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/rbac/roles/role-1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("GetRole - Not Found", func(t *testing.T) {
		mockService.EXPECT().GetRole("wrong").Return(nil, errors.New("not found"))

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/rbac/roles/wrong", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("CreateRole - Success", func(t *testing.T) {
		body := domain.CreateRoleRequest{Name: "New Role"}
		jsonBody, _ := json.Marshal(body)

		mockService.EXPECT().CreateRole("comp-1", gomock.Any()).Return(nil)

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/rbac/roles", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("UpdateRole - Success", func(t *testing.T) {
		body := domain.UpdateRoleRequest{Name: "Updated Name"}
		jsonBody, _ := json.Marshal(body)

		mockService.EXPECT().UpdateRole("role-1", gomock.Any()).Return(nil)

		req, _ := http.NewRequest(http.MethodPut, "/api/v1/rbac/roles/role-1", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("DeleteRole - Success", func(t *testing.T) {
		mockService.EXPECT().DeleteRole("role-1").Return(nil)

		req, _ := http.NewRequest(http.MethodDelete, "/api/v1/rbac/roles/role-1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("ListPermissions - Success", func(t *testing.T) {
		mockService.EXPECT().ListPermissions().Return([]domain.PermissionResponse{
			{ID: "p1", Resource: "user", Action: "read"},
		}, nil)

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/rbac/permissions", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestHandler_Enforce(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := rbacMock.NewMockService(ctrl)
	handler := rbac.NewHandler(mockService)

	router := gin.Default()
	router.POST("/api/v1/rbac/enforce", handler.Enforce)

	t.Run("Success", func(t *testing.T) {
		body := domain.EnforceRequest{
			EmployeeID: "emp-1",
			CompanyID:  "comp-1",
			Resource:   "user",
			Action:     "read",
		}
		jsonBody, _ := json.Marshal(body)

		mockService.EXPECT().Enforce(gomock.Any()).Return(true, nil)

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/rbac/enforce", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Validation Error", func(t *testing.T) {
		body := map[string]string{"invalid": "data"}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/rbac/enforce", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
