package attendance_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-hris/internal/attendance"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type fakeService struct {
	clockInFn  func(ctx context.Context, companyID, employeeID string, req attendance.ClockInRequest) (attendance.AttendanceResponse, error)
	clockOutFn func(ctx context.Context, companyID, employeeID string, req attendance.ClockOutRequest) (attendance.AttendanceResponse, error)
	getAllFn   func(ctx context.Context, companyID string) ([]attendance.AttendanceResponse, error)
}

func (f *fakeService) ClockIn(ctx context.Context, companyID, employeeID string, req attendance.ClockInRequest) (attendance.AttendanceResponse, error) {
	return f.clockInFn(ctx, companyID, employeeID, req)
}
func (f *fakeService) ClockOut(ctx context.Context, companyID, employeeID string, req attendance.ClockOutRequest) (attendance.AttendanceResponse, error) {
	return f.clockOutFn(ctx, companyID, employeeID, req)
}
func (f *fakeService) GetAll(ctx context.Context, companyID string) ([]attendance.AttendanceResponse, error) {
	return f.getAllFn(ctx, companyID)
}

func TestHandler_ClockInAndGetAll(t *testing.T) {
	gin.SetMode(gin.TestMode)
	companyID := uuid.New().String()
	employeeID := uuid.New().String()

	svc := &fakeService{
		clockInFn: func(ctx context.Context, cid, eid string, req attendance.ClockInRequest) (attendance.AttendanceResponse, error) {
			assert.Equal(t, companyID, cid)
			assert.Equal(t, employeeID, eid)
			return attendance.AttendanceResponse{ID: uuid.New().String(), EmployeeID: eid, CompanyID: cid}, nil
		},
		getAllFn: func(ctx context.Context, cid string) ([]attendance.AttendanceResponse, error) {
			return []attendance.AttendanceResponse{{ID: uuid.New().String()}, {ID: uuid.New().String()}}, nil
		},
		clockOutFn: func(ctx context.Context, companyID, employeeID string, req attendance.ClockOutRequest) (attendance.AttendanceResponse, error) {
			return attendance.AttendanceResponse{}, nil
		},
	}

	h := attendance.NewHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("company_id", companyID)
	c.Set("employee_id", employeeID)
	c.Request = httptest.NewRequest(http.MethodPost, "/attendances/clock-in", strings.NewReader(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.ClockIn(c)
	assert.Equal(t, http.StatusCreated, w.Code)

	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Set("company_id", companyID)
	c2.Request = httptest.NewRequest(http.MethodGet, "/attendances?page=1&page_size=1", nil)
	h.GetAll(c2)
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Contains(t, w2.Body.String(), "\"meta\"")
}
