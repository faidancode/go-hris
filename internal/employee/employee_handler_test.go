package employee_test

import (
	"context"
	"encoding/json"
	"errors"
	autherrors "go-hris/internal/auth/errors"
	"go-hris/internal/employee"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type fakeEmployeeService struct {
	CreateFn  func(ctx context.Context, req employee.CreateEmployeeRequest) (employee.EmployeeResponse, error)
	GetByIDFn func(ctx context.Context, requesterID, targetID string, hasReadAll bool) (employee.EmployeeResponse, error)
	GetAllFn  func(ctx context.Context, requesterID, targetID string, hasReadAll bool) ([]employee.EmployeeResponse, error)
}

func (f *fakeEmployeeService) Create(ctx context.Context, req employee.CreateEmployeeRequest) (employee.EmployeeResponse, error) {
	return f.CreateFn(ctx, req)
}

func (f *fakeEmployeeService) GetByID(ctx context.Context, requesterID, targetID string, hasReadAll bool) (employee.EmployeeResponse, error) {
	return f.GetByIDFn(ctx, requesterID, targetID, hasReadAll)
}

func (f *fakeEmployeeService) GetAll(ctx context.Context, requesterID, targetID string, hasReadAll bool) ([]employee.EmployeeResponse, error) {
	return f.GetAllFn(ctx, requesterID, targetID, hasReadAll)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestCreateEmployee(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		svc := &fakeEmployeeService{
			CreateFn: func(ctx context.Context, req employee.CreateEmployeeRequest) (employee.EmployeeResponse, error) {
				assert.Equal(t, "John", req.Name)
				return employee.EmployeeResponse{
					ID:   "uuid-1",
					Name: req.Name,
				}, nil
			},
		}

		r := setupTestRouter()
		h := employee.NewHandler(svc)

		r.POST("/employees", h.Create)

		body := `{
			"name":"John",
			"email":"john@test.com",
			"company_id":"company-1"
		}`

		req := httptest.NewRequest(http.MethodPost, "/employees", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("validation error", func(t *testing.T) {
		svc := &fakeEmployeeService{}
		r := setupTestRouter()
		h := employee.NewHandler(svc)

		r.POST("/employees", h.Create)

		req := httptest.NewRequest(http.MethodPost, "/employees", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		svc := &fakeEmployeeService{
			CreateFn: func(ctx context.Context, req employee.CreateEmployeeRequest) (employee.EmployeeResponse, error) {
				return employee.EmployeeResponse{}, errors.New("failed")
			},
		}

		r := setupTestRouter()
		h := employee.NewHandler(svc)

		r.POST("/employees", h.Create)

		body := `{"name":"John","email":"john@test.com","company_id":"company-1"}`

		req := httptest.NewRequest(http.MethodPost, "/employees", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestGetEmployeeByID(t *testing.T) {
	t.Run("success - get own profile", func(t *testing.T) {
		svc := &fakeEmployeeService{
			GetByIDFn: func(ctx context.Context, requesterID, targetID string, hasReadAll bool) (employee.EmployeeResponse, error) {
				// Memastikan handler meneruskan ID yang benar dari context
				assert.Equal(t, "emp-123", requesterID)
				assert.Equal(t, "emp-123", targetID)
				return employee.EmployeeResponse{ID: targetID, Name: "John Doe"}, nil
			},
		}

		r := setupTestRouter()
		h := employee.NewHandler(svc)

		// Mock middleware untuk set context
		r.GET("/employees/:id", func(c *gin.Context) {
			c.Set("employee_id", "emp-123")
			c.Set("has_read_all", false) // Bukan HR
			h.GetById(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/employees/emp-123", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("forbidden - access others without permission", func(t *testing.T) {
		svc := &fakeEmployeeService{
			GetByIDFn: func(ctx context.Context, reqID, tarID string, hasAll bool) (employee.EmployeeResponse, error) {
				// Service mengembalikan error forbidden
				return employee.EmployeeResponse{}, autherrors.ErrForbidden
			},
		}

		r := setupTestRouter()
		h := employee.NewHandler(svc)

		r.GET("/employees/:id", func(c *gin.Context) {
			c.Set("employee_id", "emp-regular")
			c.Set("has_read_all", false)
			h.GetById(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/employees/emp-boss", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestGetAllEmployees(t *testing.T) {
	t.Run("success - as HR", func(t *testing.T) {
		svc := &fakeEmployeeService{
			GetAllFn: func(ctx context.Context, companyID, requesterID string, hasReadAll bool) ([]employee.EmployeeResponse, error) {
				assert.True(t, hasReadAll)
				return []employee.EmployeeResponse{
					{ID: "1", Name: "Emp 1"},
					{ID: "2", Name: "Emp 2"},
				}, nil
			},
		}

		r := setupTestRouter()
		h := employee.NewHandler(svc)

		r.GET("/employees", func(c *gin.Context) {
			c.Set("employee_id", "hr-1")
			c.Set("has_read_all", true)
			h.GetAll(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/employees?company_id=comp-1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp []employee.EmployeeResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Len(t, resp, 2)
	})

	t.Run("error - database connection failed", func(t *testing.T) {
		svc := &fakeEmployeeService{
			GetAllFn: func(ctx context.Context, cID, rID string, hasAll bool) ([]employee.EmployeeResponse, error) {
				return nil, errors.New("db error")
			},
		}

		r := setupTestRouter()
		h := employee.NewHandler(svc)

		r.GET("/employees", func(c *gin.Context) {
			c.Set("employee_id", "hr-1")
			c.Set("has_read_all", true)
			h.GetAll(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/employees", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("failed - as employee (forbidden)", func(t *testing.T) {
		// Stub service yang menolak akses jika bukan HR
		svc := &fakeEmployeeService{
			GetAllFn: func(ctx context.Context, companyID, requesterID string, hasReadAll bool) ([]employee.EmployeeResponse, error) {
				// Dalam logika bisnis, jika hasReadAll false untuk endpoint GetAll,
				// maka user tidak berhak melihat list karyawan lain.
				if !hasReadAll {
					return nil, autherrors.ErrForbidden
				}
				return []employee.EmployeeResponse{}, nil
			},
		}

		r := setupTestRouter()
		h := employee.NewHandler(svc)

		// Simulasi route yang diproteksi
		r.GET("/employees", func(c *gin.Context) {
			// Simulasi data dari middleware untuk Employee biasa
			c.Set("employee_id", "emp-regular-001")
			c.Set("company_id", "comp-abc")
			c.Set("has_read_all", false) // Ini kunci kegagalannya
			h.GetAll(c)
		})

		// Request tanpa filter khusus, mencoba melihat semua
		req := httptest.NewRequest(http.MethodGet, "/employees?company_id=comp-abc", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusForbidden, w.Code)

		// Verifikasi body response jika Anda menggunakan response helper
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "forbidden", resp["error"])
	})
}
