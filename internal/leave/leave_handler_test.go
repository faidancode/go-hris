package leave_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-hris/internal/leave"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type fakeLeaveService struct {
	createFn  func(ctx context.Context, companyID, actorID string, req leave.CreateLeaveRequest) (leave.LeaveResponse, error)
	getAllFn  func(ctx context.Context, companyID string) ([]leave.LeaveResponse, error)
	getByIDFn func(ctx context.Context, companyID, id string) (leave.LeaveResponse, error)
	updateFn  func(ctx context.Context, companyID, actorID, id string, req leave.UpdateLeaveRequest) (leave.LeaveResponse, error)
	deleteFn  func(ctx context.Context, companyID, id string) error
}

func (f *fakeLeaveService) Create(ctx context.Context, companyID, actorID string, req leave.CreateLeaveRequest) (leave.LeaveResponse, error) {
	return f.createFn(ctx, companyID, actorID, req)
}
func (f *fakeLeaveService) GetAll(ctx context.Context, companyID string) ([]leave.LeaveResponse, error) {
	return f.getAllFn(ctx, companyID)
}
func (f *fakeLeaveService) GetByID(ctx context.Context, companyID, id string) (leave.LeaveResponse, error) {
	return f.getByIDFn(ctx, companyID, id)
}
func (f *fakeLeaveService) Update(ctx context.Context, companyID, actorID, id string, req leave.UpdateLeaveRequest) (leave.LeaveResponse, error) {
	return f.updateFn(ctx, companyID, actorID, id, req)
}
func (f *fakeLeaveService) Delete(ctx context.Context, companyID, id string) error {
	return f.deleteFn(ctx, companyID, id)
}

func TestLeaveHandler_Create(t *testing.T) {
	t.Run("success uses user_id_validated fallback", func(t *testing.T) {
		companyID := uuid.New().String()
		actorID := uuid.New().String()
		employeeID := uuid.New().String()

		svc := &fakeLeaveService{
			createFn: func(ctx context.Context, cid, aid string, req leave.CreateLeaveRequest) (leave.LeaveResponse, error) {
				assert.Equal(t, companyID, cid)
				assert.Equal(t, actorID, aid)
				assert.Equal(t, employeeID, req.EmployeeID)
				return leave.LeaveResponse{
					ID:         uuid.New().String(),
					CompanyID:  cid,
					EmployeeID: req.EmployeeID,
					LeaveType:  req.LeaveType,
					StartDate:  req.StartDate,
					EndDate:    req.EndDate,
					TotalDays:  2,
					Reason:     req.Reason,
					Status:     leave.StatusPending,
					CreatedBy:  aid,
				}, nil
			},
		}

		h := leave.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := `{"employee_id":"` + employeeID + `","leave_type":"ANNUAL","start_date":"2026-03-10","end_date":"2026-03-11","reason":"Family matters"}`
		c.Request = httptest.NewRequest(http.MethodPost, "/leaves", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("company_id", companyID)
		c.Set("user_id_validated", actorID)

		h.Create(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		var got leave.LeaveResponse
		err := json.Unmarshal(w.Body.Bytes(), &got)
		assert.NoError(t, err)
		assert.Equal(t, companyID, got.CompanyID)
		assert.Equal(t, employeeID, got.EmployeeID)
		assert.Equal(t, "ANNUAL", got.LeaveType)
		assert.Equal(t, 2, got.TotalDays)
		assert.Equal(t, leave.StatusPending, got.Status)
		assert.Equal(t, actorID, got.CreatedBy)
	})

	t.Run("negative validation error", func(t *testing.T) {
		h := leave.NewHandler(&fakeLeaveService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/leaves", strings.NewReader(`{}`))
		c.Request.Header.Set("Content-Type", "application/json")

		h.Create(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("negative service error", func(t *testing.T) {
		svc := &fakeLeaveService{
			createFn: func(ctx context.Context, companyID, actorID string, req leave.CreateLeaveRequest) (leave.LeaveResponse, error) {
				return leave.LeaveResponse{}, errors.New("create failed")
			},
		}
		h := leave.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := `{"employee_id":"` + uuid.New().String() + `","leave_type":"ANNUAL","start_date":"2026-03-10","end_date":"2026-03-11"}`
		c.Request = httptest.NewRequest(http.MethodPost, "/leaves", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("company_id", uuid.New().String())
		c.Set("employee_id", uuid.New().String())

		h.Create(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "create failed")
	})
}

func TestLeaveHandler_GetAll(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		companyID := uuid.New().String()
		svc := &fakeLeaveService{
			getAllFn: func(ctx context.Context, cid string) ([]leave.LeaveResponse, error) {
				assert.Equal(t, companyID, cid)
				return []leave.LeaveResponse{
					{ID: uuid.New().String(), CompanyID: cid, LeaveType: "SICK", Status: leave.StatusPending},
				}, nil
			},
		}

		h := leave.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/leaves", nil)
		c.Set("company_id", companyID)

		h.GetAll(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var got []leave.LeaveResponse
		err := json.Unmarshal(w.Body.Bytes(), &got)
		assert.NoError(t, err)
		assert.Len(t, got, 1)
		assert.Equal(t, "SICK", got[0].LeaveType)
		assert.Equal(t, leave.StatusPending, got[0].Status)
	})

	t.Run("negative service error", func(t *testing.T) {
		svc := &fakeLeaveService{
			getAllFn: func(ctx context.Context, cid string) ([]leave.LeaveResponse, error) {
				return nil, errors.New("db error")
			},
		}
		h := leave.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/leaves", nil)
		c.Set("company_id", uuid.New().String())

		h.GetAll(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "db error")
	})
}

func TestLeaveHandler_GetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		companyID := uuid.New().String()
		leaveID := uuid.New().String()
		svc := &fakeLeaveService{
			getByIDFn: func(ctx context.Context, cid, id string) (leave.LeaveResponse, error) {
				assert.Equal(t, companyID, cid)
				assert.Equal(t, leaveID, id)
				return leave.LeaveResponse{ID: id, CompanyID: cid, LeaveType: "ANNUAL"}, nil
			},
		}
		h := leave.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/leaves/"+leaveID, nil)
		c.Params = []gin.Param{{Key: "id", Value: leaveID}}
		c.Set("company_id", companyID)

		h.GetById(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var got leave.LeaveResponse
		err := json.Unmarshal(w.Body.Bytes(), &got)
		assert.NoError(t, err)
		assert.Equal(t, leaveID, got.ID)
		assert.Equal(t, companyID, got.CompanyID)
		assert.Equal(t, "ANNUAL", got.LeaveType)
	})

	t.Run("negative service error", func(t *testing.T) {
		svc := &fakeLeaveService{
			getByIDFn: func(ctx context.Context, cid, id string) (leave.LeaveResponse, error) {
				return leave.LeaveResponse{}, errors.New("not found")
			},
		}
		h := leave.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/leaves/"+uuid.New().String(), nil)
		c.Params = []gin.Param{{Key: "id", Value: uuid.New().String()}}
		c.Set("company_id", uuid.New().String())

		h.GetById(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "not found")
	})
}

func TestLeaveHandler_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		companyID := uuid.New().String()
		actorID := uuid.New().String()
		leaveID := uuid.New().String()
		employeeID := uuid.New().String()
		approvedBy := uuid.New().String()

		svc := &fakeLeaveService{
			updateFn: func(ctx context.Context, cid, aid, id string, req leave.UpdateLeaveRequest) (leave.LeaveResponse, error) {
				assert.Equal(t, companyID, cid)
				assert.Equal(t, actorID, aid)
				assert.Equal(t, leaveID, id)
				assert.Equal(t, leave.StatusApproved, req.Status)
				return leave.LeaveResponse{
					ID:         id,
					CompanyID:  cid,
					EmployeeID: req.EmployeeID,
					Status:     req.Status,
					ApprovedBy: req.ApprovedBy,
				}, nil
			},
		}

		h := leave.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := `{"employee_id":"` + employeeID + `","leave_type":"ANNUAL","start_date":"2026-06-01","end_date":"2026-06-03","reason":"Family trip","status":"APPROVED","approved_by":"` + approvedBy + `"}`
		c.Request = httptest.NewRequest(http.MethodPut, "/leaves/"+leaveID, strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "id", Value: leaveID}}
		c.Set("company_id", companyID)
		c.Set("employee_id", actorID)

		h.Update(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var got leave.LeaveResponse
		err := json.Unmarshal(w.Body.Bytes(), &got)
		assert.NoError(t, err)
		assert.Equal(t, leaveID, got.ID)
		assert.Equal(t, companyID, got.CompanyID)
		assert.Equal(t, employeeID, got.EmployeeID)
		assert.Equal(t, leave.StatusApproved, got.Status)
		assert.NotNil(t, got.ApprovedBy)
		assert.Equal(t, approvedBy, *got.ApprovedBy)
	})

	t.Run("negative validation error", func(t *testing.T) {
		h := leave.NewHandler(&fakeLeaveService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/leaves/123", strings.NewReader(`{}`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "id", Value: "123"}}

		h.Update(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("negative service error", func(t *testing.T) {
		svc := &fakeLeaveService{
			updateFn: func(ctx context.Context, companyID, actorID, id string, req leave.UpdateLeaveRequest) (leave.LeaveResponse, error) {
				return leave.LeaveResponse{}, errors.New("update failed")
			},
		}
		h := leave.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := `{"employee_id":"` + uuid.New().String() + `","leave_type":"ANNUAL","start_date":"2026-06-01","end_date":"2026-06-02","status":"PENDING"}`
		c.Request = httptest.NewRequest(http.MethodPut, "/leaves/123", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "id", Value: "123"}}
		c.Set("company_id", uuid.New().String())
		c.Set("employee_id", uuid.New().String())

		h.Update(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "update failed")
	})
}

func TestLeaveHandler_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		companyID := uuid.New().String()
		leaveID := uuid.New().String()
		svc := &fakeLeaveService{
			deleteFn: func(ctx context.Context, cid, id string) error {
				assert.Equal(t, companyID, cid)
				assert.Equal(t, leaveID, id)
				return nil
			},
		}

		h := leave.NewHandler(svc)
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("company_id", companyID)
			c.Next()
		})
		r.DELETE("/leaves/:id", h.Delete)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/leaves/"+leaveID, nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("negative service error", func(t *testing.T) {
		svc := &fakeLeaveService{
			deleteFn: func(ctx context.Context, cid, id string) error {
				return errors.New("delete failed")
			},
		}

		h := leave.NewHandler(svc)
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("company_id", uuid.New().String())
			c.Next()
		})
		r.DELETE("/leaves/:id", h.Delete)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/leaves/"+uuid.New().String(), nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "delete failed")
	})
}
