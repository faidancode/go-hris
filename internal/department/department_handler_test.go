package department_test

import (
	"context"
	"errors"
	"go-hris/internal/department"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type fakeDepartmentService struct {
	CreateFn  func(ctx context.Context, companyID string, req department.CreateDepartmentRequest) (department.DepartmentResponse, error)
	GetAllFn  func(ctx context.Context, companyID string) ([]department.DepartmentResponse, error)
	GetByIDFn func(ctx context.Context, companyID, id string) (department.DepartmentResponse, error)
	UpdateFn  func(ctx context.Context, companyID, id string, req department.UpdateDepartmentRequest) (department.DepartmentResponse, error)
	DeleteFn  func(ctx context.Context, companyID, id string) error
}

func (f *fakeDepartmentService) Create(ctx context.Context, companyID string, req department.CreateDepartmentRequest) (department.DepartmentResponse, error) {
	return f.CreateFn(ctx, companyID, req)
}
func (f *fakeDepartmentService) GetAll(ctx context.Context, companyID string) ([]department.DepartmentResponse, error) {
	return f.GetAllFn(ctx, companyID)
}
func (f *fakeDepartmentService) GetByID(ctx context.Context, companyID, id string) (department.DepartmentResponse, error) {
	return f.GetByIDFn(ctx, companyID, id)
}
func (f *fakeDepartmentService) Update(ctx context.Context, companyID, id string, req department.UpdateDepartmentRequest) (department.DepartmentResponse, error) {
	return f.UpdateFn(ctx, companyID, id, req)
}
func (f *fakeDepartmentService) Delete(ctx context.Context, companyID, id string) error {
	return f.DeleteFn(ctx, companyID, id)
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func withCompany(companyID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("company_id", companyID)
		c.Next()
	}
}

// --- Test Create ---
func TestDepartmentHandler_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		companyID := uuid.New().String()
		svc := &fakeDepartmentService{
			CreateFn: func(ctx context.Context, cid string, req department.CreateDepartmentRequest) (department.DepartmentResponse, error) {
				assert.Equal(t, companyID, cid)
				return department.DepartmentResponse{ID: uuid.New().String(), Name: req.Name, CompanyID: cid}, nil
			},
		}

		h := department.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := `{"name":"HR"}`
		c.Request = httptest.NewRequest(http.MethodPost, "/departments", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("company_id", companyID)

		h.Create(c)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("validation error", func(t *testing.T) {
		h := department.NewHandler(&fakeDepartmentService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(http.MethodPost, "/departments", strings.NewReader(`{}`))
		c.Request.Header.Set("Content-Type", "application/json")

		h.Create(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		svc := &fakeDepartmentService{
			CreateFn: func(ctx context.Context, cid string, req department.CreateDepartmentRequest) (department.DepartmentResponse, error) {
				return department.DepartmentResponse{}, errors.New("failed")
			},
		}

		h := department.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := `{"name":"HR"}`
		c.Request = httptest.NewRequest(http.MethodPost, "/departments", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("company_id", uuid.New().String())

		h.Create(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// --- Test GetAll ---
func TestDepartmentHandler_GetAll(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		companyID := uuid.New().String()
		svc := &fakeDepartmentService{
			GetAllFn: func(ctx context.Context, cid string) ([]department.DepartmentResponse, error) {
				assert.Equal(t, companyID, cid)
				return []department.DepartmentResponse{{ID: uuid.New().String(), Name: "HR"}}, nil
			},
		}

		h := department.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(http.MethodGet, "/departments", nil)
		c.Set("company_id", companyID)

		h.GetAll(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// --- Test GetByID ---
func TestDepartmentHandler_GetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		companyID := uuid.New().String()
		deptID := uuid.New().String()

		svc := &fakeDepartmentService{
			GetByIDFn: func(ctx context.Context, cid, id string) (department.DepartmentResponse, error) {
				assert.Equal(t, companyID, cid)
				assert.Equal(t, deptID, id)
				return department.DepartmentResponse{ID: id, Name: "HR", CompanyID: cid}, nil
			},
		}

		h := department.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(http.MethodGet, "/departments/"+deptID, nil)
		c.Params = []gin.Param{{Key: "id", Value: deptID}}
		c.Set("company_id", companyID)

		h.GetById(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// --- Test Update ---
func TestDepartmentHandler_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		companyID := uuid.New().String()
		deptID := uuid.New().String()

		svc := &fakeDepartmentService{
			UpdateFn: func(ctx context.Context, cid, id string, req department.UpdateDepartmentRequest) (department.DepartmentResponse, error) {
				return department.DepartmentResponse{ID: id, Name: req.Name, CompanyID: cid}, nil
			},
		}

		h := department.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := `{"name":"Finance"}`
		c.Request = httptest.NewRequest(http.MethodPut, "/departments/"+deptID, strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "id", Value: deptID}}
		c.Set("company_id", companyID)

		h.Update(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// --- Test Delete ---
func TestDepartmentHandler_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		companyID := uuid.New().String()
		deptID := uuid.New().String()

		svc := &fakeDepartmentService{
			DeleteFn: func(ctx context.Context, cid, id string) error {
				assert.Equal(t, companyID, cid)
				assert.Equal(t, deptID, id)
				return nil
			},
		}

		h := department.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(http.MethodDelete, "/departments/"+deptID, nil)
		c.Params = []gin.Param{{Key: "id", Value: deptID}}
		c.Set("company_id", companyID)

		h.Delete(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
