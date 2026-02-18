package position_test

import (
	"context"
	"encoding/json"
	"errors"
	"go-hris/internal/position"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

type fakePositionService struct {
	CreateFn  func(ctx context.Context, companyID string, req position.CreatePositionRequest) (position.PositionResponse, error)
	GetAllFn  func(ctx context.Context, companyID string) ([]position.PositionResponse, error)
	GetByIDFn func(ctx context.Context, companyID, id string) (position.PositionResponse, error)
	UpdateFn  func(ctx context.Context, companyID, id string, req position.UpdatePositionRequest) (position.PositionResponse, error)
	DeleteFn  func(ctx context.Context, companyID, id string) error
}

func (f *fakePositionService) Create(ctx context.Context, companyID string, req position.CreatePositionRequest) (position.PositionResponse, error) {
	return f.CreateFn(ctx, companyID, req)
}
func (f *fakePositionService) GetAll(ctx context.Context, companyID string) ([]position.PositionResponse, error) {
	return f.GetAllFn(ctx, companyID)
}
func (f *fakePositionService) GetByID(ctx context.Context, companyID, id string) (position.PositionResponse, error) {
	return f.GetByIDFn(ctx, companyID, id)
}
func (f *fakePositionService) Update(ctx context.Context, companyID, id string, req position.UpdatePositionRequest) (position.PositionResponse, error) {
	return f.UpdateFn(ctx, companyID, id, req)
}
func (f *fakePositionService) Delete(ctx context.Context, companyID, id string) error {
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
func TestPositionHandler_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		companyID := uuid.New().String()
		departmentID := uuid.New().String()
		svc := &fakePositionService{
			CreateFn: func(ctx context.Context, cid string, req position.CreatePositionRequest) (position.PositionResponse, error) {
				assert.Equal(t, companyID, cid)
				assert.Equal(t, departmentID, req.DepartmentID)
				return position.PositionResponse{ID: uuid.New().String(), Name: req.Name, CompanyID: cid, DepartmentID: req.DepartmentID}, nil
			},
		}

		h := position.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := `{"name":"HR","department_id":"` + departmentID + `"}`
		c.Request = httptest.NewRequest(http.MethodPost, "/positions", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("company_id", companyID)

		h.Create(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		env := mustDecodeEnvelope(t, w.Body.Bytes())
		assert.True(t, env.Ok)
		var got position.PositionResponse
		err := json.Unmarshal(env.Data, &got)
		assert.NoError(t, err)
		assert.Equal(t, companyID, got.CompanyID)
		assert.Equal(t, departmentID, got.DepartmentID)
		assert.Equal(t, "HR", got.Name)
	})

	t.Run("validation error", func(t *testing.T) {
		h := position.NewHandler(&fakePositionService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(http.MethodPost, "/positions", strings.NewReader(`{}`))
		c.Request.Header.Set("Content-Type", "application/json")

		h.Create(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		departmentID := uuid.New().String()
		svc := &fakePositionService{
			CreateFn: func(ctx context.Context, cid string, req position.CreatePositionRequest) (position.PositionResponse, error) {
				return position.PositionResponse{}, errors.New("failed")
			},
		}

		h := position.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := `{"name":"HR","department_id":"` + departmentID + `"}`
		c.Request = httptest.NewRequest(http.MethodPost, "/positions", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("company_id", uuid.New().String())

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

// --- Test GetAll ---
func TestPositionHandler_GetAll(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		companyID := uuid.New().String()
		departmentID := uuid.New().String()
		svc := &fakePositionService{
			GetAllFn: func(ctx context.Context, cid string) ([]position.PositionResponse, error) {
				assert.Equal(t, companyID, cid)
				return []position.PositionResponse{{ID: uuid.New().String(), Name: "HR", CompanyID: cid, DepartmentID: departmentID}}, nil
			},
		}

		h := position.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(http.MethodGet, "/positions", nil)
		c.Set("company_id", companyID)

		h.GetAll(c)

		assert.Equal(t, http.StatusOK, w.Code)
		env := mustDecodeEnvelope(t, w.Body.Bytes())
		assert.True(t, env.Ok)
		var got []position.PositionResponse
		err := json.Unmarshal(env.Data, &got)
		assert.NoError(t, err)
		assert.Len(t, got, 1)
		assert.Equal(t, companyID, got[0].CompanyID)
		assert.Equal(t, departmentID, got[0].DepartmentID)
	})
}

// --- Test GetByID ---
func TestPositionHandler_GetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		companyID := uuid.New().String()
		deptID := uuid.New().String()
		departmentID := uuid.New().String()

		svc := &fakePositionService{
			GetByIDFn: func(ctx context.Context, cid, id string) (position.PositionResponse, error) {
				assert.Equal(t, companyID, cid)
				assert.Equal(t, deptID, id)
				return position.PositionResponse{ID: id, Name: "HR", CompanyID: cid, DepartmentID: departmentID}, nil
			},
		}

		h := position.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(http.MethodGet, "/positions/"+deptID, nil)
		c.Params = []gin.Param{{Key: "id", Value: deptID}}
		c.Set("company_id", companyID)

		h.GetById(c)

		assert.Equal(t, http.StatusOK, w.Code)
		env := mustDecodeEnvelope(t, w.Body.Bytes())
		assert.True(t, env.Ok)
		var got position.PositionResponse
		err := json.Unmarshal(env.Data, &got)
		assert.NoError(t, err)
		assert.Equal(t, deptID, got.ID)
		assert.Equal(t, companyID, got.CompanyID)
		assert.Equal(t, departmentID, got.DepartmentID)
	})
}

// --- Test Update ---
func TestPositionHandler_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		companyID := uuid.New().String()
		deptID := uuid.New().String()
		departmentID := uuid.New().String()

		svc := &fakePositionService{
			UpdateFn: func(ctx context.Context, cid, id string, req position.UpdatePositionRequest) (position.PositionResponse, error) {
				assert.Equal(t, departmentID, req.DepartmentID)
				return position.PositionResponse{ID: id, Name: req.Name, CompanyID: cid, DepartmentID: req.DepartmentID}, nil
			},
		}

		h := position.NewHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := `{"name":"Finance","department_id":"` + departmentID + `"}`
		c.Request = httptest.NewRequest(http.MethodPut, "/positions/"+deptID, strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "id", Value: deptID}}
		c.Set("company_id", companyID)

		h.Update(c)

		assert.Equal(t, http.StatusOK, w.Code)
		env := mustDecodeEnvelope(t, w.Body.Bytes())
		assert.True(t, env.Ok)
		var got position.PositionResponse
		err := json.Unmarshal(env.Data, &got)
		assert.NoError(t, err)
		assert.Equal(t, deptID, got.ID)
		assert.Equal(t, "Finance", got.Name)
		assert.Equal(t, departmentID, got.DepartmentID)
	})
}

// --- Test Delete ---
func TestPositionHandler_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		companyID := uuid.New().String()
		deptID := uuid.New().String()

		svc := &fakePositionService{
			DeleteFn: func(ctx context.Context, cid, id string) error {
				assert.Equal(t, companyID, cid)
				assert.Equal(t, deptID, id)
				return nil
			},
		}

		h := position.NewHandler(svc)
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("company_id", companyID)
			c.Next()
		})
		r.DELETE("/positions/:id", h.Delete)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/positions/"+deptID, nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}
