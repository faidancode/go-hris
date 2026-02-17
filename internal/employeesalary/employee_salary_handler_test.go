package employeesalary_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-hris/internal/employeesalary"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type fakeEmployeeSalaryService struct {
	createFn  func(ctx context.Context, companyID string, req employeesalary.CreateEmployeeSalaryRequest) (employeesalary.EmployeeSalaryResponse, error)
	getAllFn  func(ctx context.Context, companyID string) ([]employeesalary.EmployeeSalaryResponse, error)
	getByIDFn func(ctx context.Context, companyID, id string) (employeesalary.EmployeeSalaryResponse, error)
	updateFn  func(ctx context.Context, companyID, id string, req employeesalary.UpdateEmployeeSalaryRequest) (employeesalary.EmployeeSalaryResponse, error)
	deleteFn  func(ctx context.Context, companyID, id string) error
}

func (f *fakeEmployeeSalaryService) Create(ctx context.Context, companyID string, req employeesalary.CreateEmployeeSalaryRequest) (employeesalary.EmployeeSalaryResponse, error) {
	return f.createFn(ctx, companyID, req)
}
func (f *fakeEmployeeSalaryService) GetAll(ctx context.Context, companyID string) ([]employeesalary.EmployeeSalaryResponse, error) {
	return f.getAllFn(ctx, companyID)
}
func (f *fakeEmployeeSalaryService) GetByID(ctx context.Context, companyID, id string) (employeesalary.EmployeeSalaryResponse, error) {
	return f.getByIDFn(ctx, companyID, id)
}
func (f *fakeEmployeeSalaryService) Update(ctx context.Context, companyID, id string, req employeesalary.UpdateEmployeeSalaryRequest) (employeesalary.EmployeeSalaryResponse, error) {
	return f.updateFn(ctx, companyID, id, req)
}
func (f *fakeEmployeeSalaryService) Delete(ctx context.Context, companyID, id string) error {
	return f.deleteFn(ctx, companyID, id)
}

func TestEmployeeSalaryHandler_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		companyID := uuid.New().String()
		employeeID := uuid.New().String()

		svc := &fakeEmployeeSalaryService{
			createFn: func(ctx context.Context, cid string, req employeesalary.CreateEmployeeSalaryRequest) (employeesalary.EmployeeSalaryResponse, error) {
				assert.Equal(t, companyID, cid)
				assert.Equal(t, employeeID, req.EmployeeID)
				return employeesalary.EmployeeSalaryResponse{
					ID:            uuid.New().String(),
					EmployeeID:    req.EmployeeID,
					BaseSalary:    req.BaseSalary,
					EffectiveDate: req.EffectiveDate,
				}, nil
			},
		}

		h := employeesalary.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := `{"employee_id":"` + employeeID + `","base_salary":10000000,"effective_date":"2026-02-01"}`
		req := httptest.NewRequest(http.MethodPost, "/employee-salaries", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
		c.Set("company_id", companyID)

		h.Create(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), employeeID)
	})

	t.Run("validation error", func(t *testing.T) {
		h := employeesalary.NewHandler(&fakeEmployeeSalaryService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := httptest.NewRequest(http.MethodPost, "/employee-salaries", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
		c.Set("company_id", uuid.New().String())

		h.Create(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		svc := &fakeEmployeeSalaryService{
			createFn: func(ctx context.Context, cid string, req employeesalary.CreateEmployeeSalaryRequest) (employeesalary.EmployeeSalaryResponse, error) {
				return employeesalary.EmployeeSalaryResponse{}, errors.New("create failed")
			},
		}

		h := employeesalary.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := `{"employee_id":"` + uuid.New().String() + `","base_salary":10000000,"effective_date":"2026-02-01"}`
		req := httptest.NewRequest(http.MethodPost, "/employee-salaries", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
		c.Set("company_id", uuid.New().String())

		h.Create(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestEmployeeSalaryHandler_GetAll(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		companyID := uuid.New().String()
		employeeID := uuid.New().String()

		svc := &fakeEmployeeSalaryService{
			getAllFn: func(ctx context.Context, cid string) ([]employeesalary.EmployeeSalaryResponse, error) {
				assert.Equal(t, companyID, cid)
				return []employeesalary.EmployeeSalaryResponse{
					{
						ID:            uuid.New().String(),
						EmployeeID:    employeeID,
						BaseSalary:    10000000,
						EffectiveDate: "2026-02-01",
					},
				}, nil
			},
		}

		h := employeesalary.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/employee-salaries", nil)
		c.Set("company_id", companyID)

		h.GetAll(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), employeeID)
	})

	t.Run("service error", func(t *testing.T) {
		svc := &fakeEmployeeSalaryService{
			getAllFn: func(ctx context.Context, cid string) ([]employeesalary.EmployeeSalaryResponse, error) {
				return nil, errors.New("db error")
			},
		}

		h := employeesalary.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/employee-salaries", nil)
		c.Set("company_id", uuid.New().String())

		h.GetAll(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestEmployeeSalaryHandler_GetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		companyID := uuid.New().String()
		salaryID := uuid.New().String()

		svc := &fakeEmployeeSalaryService{
			getByIDFn: func(ctx context.Context, cid, id string) (employeesalary.EmployeeSalaryResponse, error) {
				assert.Equal(t, companyID, cid)
				assert.Equal(t, salaryID, id)
				return employeesalary.EmployeeSalaryResponse{
					ID:            id,
					EmployeeID:    uuid.New().String(),
					BaseSalary:    10000000,
					EffectiveDate: "2026-02-01",
				}, nil
			},
		}

		h := employeesalary.NewHandler(svc)
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("company_id", companyID)
			c.Next()
		})
		r.GET("/employee-salaries/:id", h.GetById)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/employee-salaries/"+salaryID, nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), salaryID)
	})

	t.Run("service error", func(t *testing.T) {
		svc := &fakeEmployeeSalaryService{
			getByIDFn: func(ctx context.Context, cid, id string) (employeesalary.EmployeeSalaryResponse, error) {
				return employeesalary.EmployeeSalaryResponse{}, errors.New("not found")
			},
		}

		h := employeesalary.NewHandler(svc)
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("company_id", uuid.New().String())
			c.Next()
		})
		r.GET("/employee-salaries/:id", h.GetById)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/employee-salaries/"+uuid.New().String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestEmployeeSalaryHandler_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		companyID := uuid.New().String()
		salaryID := uuid.New().String()
		employeeID := uuid.New().String()

		svc := &fakeEmployeeSalaryService{
			updateFn: func(ctx context.Context, cid, id string, req employeesalary.UpdateEmployeeSalaryRequest) (employeesalary.EmployeeSalaryResponse, error) {
				assert.Equal(t, companyID, cid)
				assert.Equal(t, salaryID, id)
				return employeesalary.EmployeeSalaryResponse{
					ID:            uuid.New().String(),
					EmployeeID:    req.EmployeeID,
					BaseSalary:    req.BaseSalary,
					EffectiveDate: req.EffectiveDate,
				}, nil
			},
		}

		h := employeesalary.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := `{"employee_id":"` + employeeID + `","base_salary":12000000,"effective_date":"2026-03-01"}`
		req := httptest.NewRequest(http.MethodPut, "/employee-salaries/"+salaryID, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
		c.Params = []gin.Param{{Key: "id", Value: salaryID}}
		c.Set("company_id", companyID)

		h.Update(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), employeeID)
		assert.Contains(t, w.Body.String(), "12000000")
	})

	t.Run("validation error", func(t *testing.T) {
		h := employeesalary.NewHandler(&fakeEmployeeSalaryService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := httptest.NewRequest(http.MethodPut, "/employee-salaries/123", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
		c.Params = []gin.Param{{Key: "id", Value: "123"}}
		c.Set("company_id", uuid.New().String())

		h.Update(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		svc := &fakeEmployeeSalaryService{
			updateFn: func(ctx context.Context, cid, id string, req employeesalary.UpdateEmployeeSalaryRequest) (employeesalary.EmployeeSalaryResponse, error) {
				return employeesalary.EmployeeSalaryResponse{}, errors.New("update failed")
			},
		}

		h := employeesalary.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := `{"employee_id":"` + uuid.New().String() + `","base_salary":12000000,"effective_date":"2026-03-01"}`
		req := httptest.NewRequest(http.MethodPut, "/employee-salaries/123", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
		c.Params = []gin.Param{{Key: "id", Value: "123"}}
		c.Set("company_id", uuid.New().String())

		h.Update(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestEmployeeSalaryHandler_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		companyID := uuid.New().String()
		salaryID := uuid.New().String()

		svc := &fakeEmployeeSalaryService{
			deleteFn: func(ctx context.Context, cid, id string) error {
				assert.Equal(t, companyID, cid)
				assert.Equal(t, salaryID, id)
				return nil
			},
		}

		h := employeesalary.NewHandler(svc)
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("company_id", companyID)
			c.Next()
		})
		r.DELETE("/employee-salaries/:id", h.Delete)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/employee-salaries/"+salaryID, nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		svc := &fakeEmployeeSalaryService{
			deleteFn: func(ctx context.Context, cid, id string) error {
				return errors.New("delete failed")
			},
		}

		h := employeesalary.NewHandler(svc)
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("company_id", uuid.New().String())
			c.Next()
		})
		r.DELETE("/employee-salaries/:id", h.Delete)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/employee-salaries/"+uuid.New().String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
