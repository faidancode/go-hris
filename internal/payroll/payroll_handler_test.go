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
	payrollerrors "go-hris/internal/payroll/errors"

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
	createFn          func(ctx context.Context, companyID, actorID string, req payroll.CreatePayrollRequest) (payroll.PayrollResponse, error)
	getAllFn          func(ctx context.Context, companyID string, filter payroll.GetPayrollsFilterRequest) ([]payroll.PayrollResponse, error)
	getByIDFn         func(ctx context.Context, companyID, id string) (payroll.PayrollResponse, error)
	getBreakdownFn    func(ctx context.Context, companyID, id string) (payroll.PayrollBreakdownResponse, error)
	regenerateFn      func(ctx context.Context, companyID, actorID, id string, req payroll.RegeneratePayrollRequest) (payroll.PayrollResponse, error)
	approveFn         func(ctx context.Context, companyID, actorID, id string) (payroll.PayrollResponse, error)
	markPaidFn        func(ctx context.Context, companyID, actorID, id string) (payroll.PayrollResponse, error)
	generatePayslipFn func(ctx context.Context, companyID, id string) (payroll.PayrollResponse, error)
	deleteFn          func(ctx context.Context, companyID, id string) error
}

func (f *fakePayrollService) Create(ctx context.Context, companyID, actorID string, req payroll.CreatePayrollRequest) (payroll.PayrollResponse, error) {
	return f.createFn(ctx, companyID, actorID, req)
}

func (f *fakePayrollService) GetAll(ctx context.Context, companyID string, filter payroll.GetPayrollsFilterRequest) ([]payroll.PayrollResponse, error) {
	return f.getAllFn(ctx, companyID, filter)
}

func (f *fakePayrollService) GetByID(ctx context.Context, companyID, id string) (payroll.PayrollResponse, error) {
	return f.getByIDFn(ctx, companyID, id)
}
func (f *fakePayrollService) GetBreakdown(ctx context.Context, companyID, id string) (payroll.PayrollBreakdownResponse, error) {
	return f.getBreakdownFn(ctx, companyID, id)
}

func (f *fakePayrollService) Regenerate(ctx context.Context, companyID, actorID, id string, req payroll.RegeneratePayrollRequest) (payroll.PayrollResponse, error) {
	return f.regenerateFn(ctx, companyID, actorID, id, req)
}

func (f *fakePayrollService) Approve(ctx context.Context, companyID, actorID, id string) (payroll.PayrollResponse, error) {
	return f.approveFn(ctx, companyID, actorID, id)
}

func (f *fakePayrollService) MarkAsPaid(ctx context.Context, companyID, actorID, id string) (payroll.PayrollResponse, error) {
	return f.markPaidFn(ctx, companyID, actorID, id)
}
func (f *fakePayrollService) GeneratePayslip(ctx context.Context, companyID, id string) (payroll.PayrollResponse, error) {
	return f.generatePayslipFn(ctx, companyID, id)
}

func (f *fakePayrollService) Delete(ctx context.Context, companyID, id string) error {
	return f.deleteFn(ctx, companyID, id)
}

func TestPayrollHandler_Create(t *testing.T) {
	companyID := uuid.New().String()
	actorID := uuid.New().String()
	employeeID := uuid.New().String()

	svc := &fakePayrollService{
		createFn: func(ctx context.Context, cid, aid string, req payroll.CreatePayrollRequest) (payroll.PayrollResponse, error) {
			assert.Equal(t, companyID, cid)
			assert.Equal(t, actorID, aid)
			assert.Equal(t, employeeID, req.EmployeeID)
			return payroll.PayrollResponse{ID: uuid.New().String(), Status: payroll.StatusDraft, CompanyID: cid, EmployeeID: req.EmployeeID, CreatedBy: aid}, nil
		},
	}

	h := payroll.NewHandler(svc)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"employee_id":"` + employeeID + `","period_start":"2026-02-01","period_end":"2026-02-28","base_salary":10000000}`
	c.Request = httptest.NewRequest(http.MethodPost, "/payrolls", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("company_id", companyID)
	c.Set("user_id", actorID)

	h.Create(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	env := mustDecodeEnvelope(t, w.Body.Bytes())
	assert.True(t, env.Ok)
}

func TestPayrollHandler_Regenerate_InvalidState(t *testing.T) {
	svc := &fakePayrollService{
		regenerateFn: func(ctx context.Context, companyID, actorID, id string, req payroll.RegeneratePayrollRequest) (payroll.PayrollResponse, error) {
			return payroll.PayrollResponse{}, payrollerrors.ErrRegenerateOnlyDraft
		},
	}

	h := payroll.NewHandler(svc)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := `{"base_salary":12000000}`
	c.Request = httptest.NewRequest(http.MethodPost, "/payrolls/123/regenerate", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = []gin.Param{{Key: "id", Value: "123"}}
	c.Set("company_id", uuid.New().String())
	c.Set("employee_id", uuid.New().String())

	h.Regenerate(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	env := mustDecodeEnvelope(t, w.Body.Bytes())
	assert.False(t, env.Ok)
	assert.Equal(t, "INVALID_STATE", env.Error.Code)
}

func TestPayrollHandler_GetBreakdown(t *testing.T) {
	companyID := uuid.New().String()
	payrollID := uuid.New().String()
	svc := &fakePayrollService{
		getBreakdownFn: func(ctx context.Context, cid, id string) (payroll.PayrollBreakdownResponse, error) {
			assert.Equal(t, companyID, cid)
			assert.Equal(t, payrollID, id)
			return payroll.PayrollBreakdownResponse{
				PayrollID: payrollID,
				Status:    payroll.StatusDraft,
				NetSalary: 10000000,
			}, nil
		},
	}

	h := payroll.NewHandler(svc)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/payrolls/"+payrollID+"/breakdown", nil)
	c.Params = []gin.Param{{Key: "id", Value: payrollID}}
	c.Set("company_id", companyID)

	h.GetBreakdown(c)

	assert.Equal(t, http.StatusOK, w.Code)
	env := mustDecodeEnvelope(t, w.Body.Bytes())
	assert.True(t, env.Ok)
}

func TestPayrollHandler_ApproveAndMarkPaid(t *testing.T) {
	companyID := uuid.New().String()
	actorID := uuid.New().String()
	id := uuid.New().String()

	svc := &fakePayrollService{
		approveFn: func(ctx context.Context, cid, aid, pid string) (payroll.PayrollResponse, error) {
			assert.Equal(t, companyID, cid)
			assert.Equal(t, actorID, aid)
			assert.Equal(t, id, pid)
			return payroll.PayrollResponse{ID: id, Status: payroll.StatusApproved}, nil
		},
		markPaidFn: func(ctx context.Context, cid, aid, pid string) (payroll.PayrollResponse, error) {
			return payroll.PayrollResponse{ID: id, Status: payroll.StatusPaid}, nil
		},
	}

	h := payroll.NewHandler(svc)

	wApprove := httptest.NewRecorder()
	cApprove, _ := gin.CreateTestContext(wApprove)
	cApprove.Request = httptest.NewRequest(http.MethodPost, "/payrolls/"+id+"/approve", nil)
	cApprove.Params = []gin.Param{{Key: "id", Value: id}}
	cApprove.Set("company_id", companyID)
	cApprove.Set("employee_id", actorID)
	h.Approve(cApprove)
	assert.Equal(t, http.StatusOK, wApprove.Code)

	wPaid := httptest.NewRecorder()
	cPaid, _ := gin.CreateTestContext(wPaid)
	cPaid.Request = httptest.NewRequest(http.MethodPost, "/payrolls/"+id+"/mark-paid", nil)
	cPaid.Params = []gin.Param{{Key: "id", Value: id}}
	cPaid.Set("company_id", companyID)
	cPaid.Set("employee_id", actorID)
	h.MarkAsPaid(cPaid)
	assert.Equal(t, http.StatusOK, wPaid.Code)
}

func TestPayrollHandler_Delete(t *testing.T) {
	svc := &fakePayrollService{
		deleteFn: func(ctx context.Context, cid, id string) error {
			return nil
		},
	}

	h := payroll.NewHandler(svc)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodDelete, "/payrolls/"+uuid.New().String(), nil)
	c.Params = []gin.Param{{Key: "id", Value: uuid.New().String()}}
	c.Set("company_id", uuid.New().String())

	h.Delete(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPayrollHandler_DownloadPayslip(t *testing.T) {
	companyID := uuid.New().String()
	payrollID := uuid.New().String()
	url := "/files/payslips/payslip_" + payrollID + ".pdf"
	svc := &fakePayrollService{
		getByIDFn: func(ctx context.Context, cid, id string) (payroll.PayrollResponse, error) {
			return payroll.PayrollResponse{ID: id, PayslipURL: &url}, nil
		},
	}

	h := payroll.NewHandler(svc)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/payrolls/"+payrollID+"/payslip/download", nil)
	c.Params = []gin.Param{{Key: "id", Value: payrollID}}
	c.Set("company_id", companyID)

	h.DownloadPayslip(c)

	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	assert.Equal(t, url, w.Header().Get("Location"))
}

func TestPayrollHandler_InternalError(t *testing.T) {
	svc := &fakePayrollService{
		getAllFn: func(ctx context.Context, companyID string, filter payroll.GetPayrollsFilterRequest) ([]payroll.PayrollResponse, error) {
			assert.Equal(t, "2026-02", filter.Period)
			assert.Equal(t, "draft", filter.Status)
			return nil, errors.New("boom")
		},
	}

	h := payroll.NewHandler(svc)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/payrolls?period=2026-02&status=draft", nil)
	c.Set("company_id", uuid.New().String())

	h.GetAll(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	env := mustDecodeEnvelope(t, w.Body.Bytes())
	assert.Equal(t, "INTERNAL_ERROR", env.Error.Code)
}
