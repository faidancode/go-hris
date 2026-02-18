package payroll_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-hris/internal/payroll"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type apiEnvelope struct {
	Ok    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error *apiError       `json:"error"`
}

func mustDecodeEnvelope(t *testing.T, body []byte) apiEnvelope {
	t.Helper()
	var env apiEnvelope
	err := json.Unmarshal(body, &env)
	assert.NoError(t, err)
	return env
}

type fakePayrollService struct {
	createFn  func(ctx context.Context, companyID, actorID string, req payroll.CreatePayrollRequest) (payroll.PayrollResponse, error)
	getAllFn  func(ctx context.Context, companyID string) ([]payroll.PayrollResponse, error)
	getByIDFn func(ctx context.Context, companyID, id string) (payroll.PayrollResponse, error)
	updateFn  func(ctx context.Context, companyID, actorID, id string, req payroll.UpdatePayrollRequest) (payroll.PayrollResponse, error)
	deleteFn  func(ctx context.Context, companyID, id string) error
}

func (f *fakePayrollService) Create(ctx context.Context, companyID, actorID string, req payroll.CreatePayrollRequest) (payroll.PayrollResponse, error) {
	return f.createFn(ctx, companyID, actorID, req)
}

func (f *fakePayrollService) GetAll(ctx context.Context, companyID string) ([]payroll.PayrollResponse, error) {
	return f.getAllFn(ctx, companyID)
}

func (f *fakePayrollService) GetByID(ctx context.Context, companyID, id string) (payroll.PayrollResponse, error) {
	return f.getByIDFn(ctx, companyID, id)
}

func (f *fakePayrollService) Update(ctx context.Context, companyID, actorID, id string, req payroll.UpdatePayrollRequest) (payroll.PayrollResponse, error) {
	return f.updateFn(ctx, companyID, actorID, id, req)
}

func (f *fakePayrollService) Delete(ctx context.Context, companyID, id string) error {
	return f.deleteFn(ctx, companyID, id)
}

func TestPayrollHandler_Create(t *testing.T) {
	t.Run("success uses user_id_validated fallback", func(t *testing.T) {
		companyID := uuid.New().String()
		actorID := uuid.New().String()
		employeeID := uuid.New().String()

		svc := &fakePayrollService{
			createFn: func(ctx context.Context, cid, aid string, req payroll.CreatePayrollRequest) (payroll.PayrollResponse, error) {
				assert.Equal(t, companyID, cid)
				assert.Equal(t, actorID, aid)
				assert.Equal(t, employeeID, req.EmployeeID)
				return payroll.PayrollResponse{
					ID:         uuid.New().String(),
					CompanyID:  cid,
					EmployeeID: req.EmployeeID,
					Status:     payroll.StatusDraft,
					BaseSalary: req.BaseSalary,
					Allowance:  req.Allowance,
					Deduction:  req.Deduction,
					NetSalary:  req.BaseSalary + req.Allowance - req.Deduction,
					CreatedBy:  actorID,
				}, nil
			},
		}

		h := payroll.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := `{"employee_id":"` + employeeID + `","period_start":"2026-02-01","period_end":"2026-02-28","base_salary":10000000,"allowance":200000,"deduction":50000}`
		c.Request = httptest.NewRequest(http.MethodPost, "/payrolls", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("company_id", companyID)
		c.Set("user_id_validated", actorID)

		h.Create(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		env := mustDecodeEnvelope(t, w.Body.Bytes())
		assert.True(t, env.Ok)
		var got payroll.PayrollResponse
		err := json.Unmarshal(env.Data, &got)
		assert.NoError(t, err)
		assert.Equal(t, companyID, got.CompanyID)
		assert.Equal(t, employeeID, got.EmployeeID)
		assert.Equal(t, payroll.StatusDraft, got.Status)
		assert.Equal(t, actorID, got.CreatedBy)
		assert.Equal(t, int64(10150000), got.NetSalary)
	})

	t.Run("negative validation error", func(t *testing.T) {
		h := payroll.NewHandler(&fakePayrollService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(http.MethodPost, "/payrolls", strings.NewReader(`{}`))
		c.Request.Header.Set("Content-Type", "application/json")

		h.Create(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("negative service error", func(t *testing.T) {
		svc := &fakePayrollService{
			createFn: func(ctx context.Context, companyID, actorID string, req payroll.CreatePayrollRequest) (payroll.PayrollResponse, error) {
				return payroll.PayrollResponse{}, errors.New("create failed")
			},
		}

		h := payroll.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := `{"employee_id":"` + uuid.New().String() + `","period_start":"2026-02-01","period_end":"2026-02-28","base_salary":10000000}`
		c.Request = httptest.NewRequest(http.MethodPost, "/payrolls", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("company_id", uuid.New().String())
		c.Set("employee_id", uuid.New().String())

		h.Create(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		env := mustDecodeEnvelope(t, w.Body.Bytes())
		assert.False(t, env.Ok)
		if assert.NotNil(t, env.Error) {
			assert.Equal(t, "INTERNAL_ERROR", env.Error.Code)
			assert.Equal(t, "Internal server error", env.Error.Message)
		}
	})
}

func TestPayrollHandler_GetAll(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		companyID := uuid.New().String()

		svc := &fakePayrollService{
			getAllFn: func(ctx context.Context, cid string) ([]payroll.PayrollResponse, error) {
				assert.Equal(t, companyID, cid)
				return []payroll.PayrollResponse{
					{ID: uuid.New().String(), Status: payroll.StatusDraft},
				}, nil
			},
		}

		h := payroll.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(http.MethodGet, "/payrolls", nil)
		c.Set("company_id", companyID)

		h.GetAll(c)

		assert.Equal(t, http.StatusOK, w.Code)
		env := mustDecodeEnvelope(t, w.Body.Bytes())
		assert.True(t, env.Ok)
		var got []payroll.PayrollResponse
		err := json.Unmarshal(env.Data, &got)
		assert.NoError(t, err)
		assert.Len(t, got, 1)
		assert.Equal(t, payroll.StatusDraft, got[0].Status)
	})

	t.Run("negative service error", func(t *testing.T) {
		svc := &fakePayrollService{
			getAllFn: func(ctx context.Context, cid string) ([]payroll.PayrollResponse, error) {
				return nil, errors.New("db error")
			},
		}

		h := payroll.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(http.MethodGet, "/payrolls", nil)
		c.Set("company_id", uuid.New().String())

		h.GetAll(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		env := mustDecodeEnvelope(t, w.Body.Bytes())
		assert.False(t, env.Ok)
		if assert.NotNil(t, env.Error) {
			assert.Equal(t, "INTERNAL_ERROR", env.Error.Code)
			assert.Equal(t, "Internal server error", env.Error.Message)
		}
	})
}

func TestPayrollHandler_GetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		companyID := uuid.New().String()
		payrollID := uuid.New().String()

		svc := &fakePayrollService{
			getByIDFn: func(ctx context.Context, cid, id string) (payroll.PayrollResponse, error) {
				assert.Equal(t, companyID, cid)
				assert.Equal(t, payrollID, id)
				return payroll.PayrollResponse{ID: id, CompanyID: cid}, nil
			},
		}

		h := payroll.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(http.MethodGet, "/payrolls/"+payrollID, nil)
		c.Params = []gin.Param{{Key: "id", Value: payrollID}}
		c.Set("company_id", companyID)

		h.GetById(c)

		assert.Equal(t, http.StatusOK, w.Code)
		env := mustDecodeEnvelope(t, w.Body.Bytes())
		assert.True(t, env.Ok)
		var got payroll.PayrollResponse
		err := json.Unmarshal(env.Data, &got)
		assert.NoError(t, err)
		assert.Equal(t, payrollID, got.ID)
		assert.Equal(t, companyID, got.CompanyID)
	})

	t.Run("negative service error", func(t *testing.T) {
		svc := &fakePayrollService{
			getByIDFn: func(ctx context.Context, cid, id string) (payroll.PayrollResponse, error) {
				return payroll.PayrollResponse{}, errors.New("not found")
			},
		}

		h := payroll.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(http.MethodGet, "/payrolls/"+uuid.New().String(), nil)
		c.Params = []gin.Param{{Key: "id", Value: uuid.New().String()}}
		c.Set("company_id", uuid.New().String())

		h.GetById(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		env := mustDecodeEnvelope(t, w.Body.Bytes())
		assert.False(t, env.Ok)
		if assert.NotNil(t, env.Error) {
			assert.Equal(t, "INTERNAL_ERROR", env.Error.Code)
			assert.Equal(t, "Internal server error", env.Error.Message)
		}
	})
}

func TestPayrollHandler_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		companyID := uuid.New().String()
		actorID := uuid.New().String()
		targetID := uuid.New().String()
		employeeID := uuid.New().String()

		svc := &fakePayrollService{
			updateFn: func(ctx context.Context, cid, aid, id string, req payroll.UpdatePayrollRequest) (payroll.PayrollResponse, error) {
				assert.Equal(t, companyID, cid)
				assert.Equal(t, actorID, aid)
				assert.Equal(t, targetID, id)
				assert.Equal(t, payroll.StatusProcessed, req.Status)
				return payroll.PayrollResponse{
					ID:         id,
					CompanyID:  cid,
					EmployeeID: req.EmployeeID,
					Status:     req.Status,
					CreatedBy:  actorID,
				}, nil
			},
		}

		h := payroll.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := `{"employee_id":"` + employeeID + `","period_start":"2026-03-01","period_end":"2026-03-31","base_salary":12000000,"allowance":300000,"deduction":100000,"status":"PROCESSED"}`
		c.Request = httptest.NewRequest(http.MethodPut, "/payrolls/"+targetID, strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "id", Value: targetID}}
		c.Set("company_id", companyID)
		c.Set("employee_id", actorID)

		h.Update(c)

		assert.Equal(t, http.StatusOK, w.Code)
		env := mustDecodeEnvelope(t, w.Body.Bytes())
		assert.True(t, env.Ok)
		var got payroll.PayrollResponse
		err := json.Unmarshal(env.Data, &got)
		assert.NoError(t, err)
		assert.Equal(t, targetID, got.ID)
		assert.Equal(t, companyID, got.CompanyID)
		assert.Equal(t, employeeID, got.EmployeeID)
		assert.Equal(t, payroll.StatusProcessed, got.Status)
	})

	t.Run("negative validation error", func(t *testing.T) {
		h := payroll.NewHandler(&fakePayrollService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(http.MethodPut, "/payrolls/123", strings.NewReader(`{}`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "id", Value: "123"}}

		h.Update(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("negative service error", func(t *testing.T) {
		svc := &fakePayrollService{
			updateFn: func(ctx context.Context, companyID, actorID, id string, req payroll.UpdatePayrollRequest) (payroll.PayrollResponse, error) {
				return payroll.PayrollResponse{}, errors.New("update failed")
			},
		}

		h := payroll.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		employeeID := uuid.New().String()

		body := `{"employee_id":"` + employeeID + `","period_start":"2026-03-01","period_end":"2026-03-31","base_salary":12000000,"allowance":300000,"deduction":100000,"status":"PROCESSED"}`
		c.Request = httptest.NewRequest(http.MethodPut, "/payrolls/123", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "id", Value: "123"}}
		c.Set("company_id", uuid.New().String())
		c.Set("employee_id", uuid.New().String())

		h.Update(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		env := mustDecodeEnvelope(t, w.Body.Bytes())
		assert.False(t, env.Ok)
		if assert.NotNil(t, env.Error) {
			assert.Equal(t, "INTERNAL_ERROR", env.Error.Code)
			assert.Equal(t, "Internal server error", env.Error.Message)
		}
	})
}

func TestPayrollHandler_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		companyID := uuid.New().String()
		targetID := uuid.New().String()

		svc := &fakePayrollService{
			deleteFn: func(ctx context.Context, cid, id string) error {
				assert.Equal(t, companyID, cid)
				assert.Equal(t, targetID, id)
				return nil
			},
		}

		h := payroll.NewHandler(svc)
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("company_id", companyID)
			c.Next()
		})
		r.DELETE("/payrolls/:id", h.Delete)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/payrolls/"+targetID, nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("negative service error", func(t *testing.T) {
		svc := &fakePayrollService{
			deleteFn: func(ctx context.Context, cid, id string) error {
				return errors.New("delete failed")
			},
		}

		h := payroll.NewHandler(svc)
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("company_id", uuid.New().String())
			c.Next()
		})
		r.DELETE("/payrolls/:id", h.Delete)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/payrolls/"+uuid.New().String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		env := mustDecodeEnvelope(t, w.Body.Bytes())
		assert.False(t, env.Ok)
		if assert.NotNil(t, env.Error) {
			assert.Equal(t, "INTERNAL_ERROR", env.Error.Code)
			assert.Equal(t, "Internal server error", env.Error.Message)
		}
	})
}
