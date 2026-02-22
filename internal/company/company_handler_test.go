package company_test

import (
	"context"
	"errors"
	"go-hris/internal/company"
	companyerrors "go-hris/internal/company/errors"
	"go-hris/internal/shared/apperror"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

//
// ==============================
// FAKE SERVICE
// ==============================
//

type fakeCompanyService struct {
	GetByIDFn            func(ctx context.Context, id string) (*company.CompanyResponse, error)
	GetByEmailFn         func(ctx context.Context, email string) (*company.CompanyResponse, error)
	UpdateFn             func(ctx context.Context, id string, req company.UpdateCompanyRequest) (*company.CompanyResponse, error)
	UpsertRegistrationFn func(ctx context.Context, companyID string, req company.UpsertCompanyRegistrationRequest) error
	ListRegistrationsFn  func(ctx context.Context, companyID string) ([]company.CompanyRegistrationResponse, error)
	DeleteRegistrationFn func(ctx context.Context, companyID string, regType company.RegistrationType) error
}

func (f *fakeCompanyService) GetByID(ctx context.Context, id string) (*company.CompanyResponse, error) {
	return f.GetByIDFn(ctx, id)
}
func (f *fakeCompanyService) GetByEmail(ctx context.Context, email string) (*company.CompanyResponse, error) {
	return f.GetByEmailFn(ctx, email)
}

func (f *fakeCompanyService) Update(ctx context.Context, id string, req company.UpdateCompanyRequest) (*company.CompanyResponse, error) {
	return f.UpdateFn(ctx, id, req)
}

func (f *fakeCompanyService) UpsertRegistration(ctx context.Context, companyID string, req company.UpsertCompanyRegistrationRequest) error {
	return f.UpsertRegistrationFn(ctx, companyID, req)
}

func (f *fakeCompanyService) ListRegistrations(ctx context.Context, companyID string) ([]company.CompanyRegistrationResponse, error) {
	return f.ListRegistrationsFn(ctx, companyID)
}

func (f *fakeCompanyService) DeleteRegistration(ctx context.Context, companyID string, regType company.RegistrationType) error {
	return f.DeleteRegistrationFn(ctx, companyID, regType)
}

// folder: internal/company/handler_test.go

func setupHandlerTest(t *testing.T, svc company.Service) (*company.Handler, *httptest.ResponseRecorder, *gin.Context) {
	t.Helper()

	logger := zaptest.NewLogger(t)
	h := company.NewHandler(svc, logger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request = req

	return h, w, c
}

//
// ==============================
// GET ME
// ==============================
//

func TestCompanyHandler_GetMe(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		compID := uuid.New().String()

		svc := &fakeCompanyService{
			GetByIDFn: func(ctx context.Context, id string) (*company.CompanyResponse, error) {
				assert.Equal(t, compID, id)
				return &company.CompanyResponse{
					ID:   compID,
					Name: "Test Co",
				}, nil
			},
		}

		h, w, c := setupHandlerTest(t, svc)

		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		c.Request = req
		c.Set("company_id", compID)

		h.GetMe(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Test Co")
	})

	t.Run("service error", func(t *testing.T) {
		svc := &fakeCompanyService{
			GetByIDFn: func(ctx context.Context, id string) (*company.CompanyResponse, error) {
				return nil, errors.New("db error")
			},
		}

		h, w, c := setupHandlerTest(t, svc)

		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		c.Request = req
		c.Set("company_id", uuid.New().String())

		h.GetMe(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

//
// ==============================
// UPSERT REGISTRATION
// ==============================
//

func TestCompanyHandler_UpsertRegistration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)

	t.Run("success", func(t *testing.T) {
		compID := uuid.New().String()

		svc := &fakeCompanyService{
			UpsertRegistrationFn: func(ctx context.Context, cid string, req company.UpsertCompanyRegistrationRequest) error {
				assert.Equal(t, compID, cid)
				assert.Equal(t, company.RegistrationType("npwp"), req.Type)
				return nil
			},
		}

		h := company.NewHandler(svc, logger)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := `{"type":"npwp","number":"123456"}`
		req := httptest.NewRequest(http.MethodPost, "/registrations", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		c.Request = req
		c.Set("company_id", compID)

		h.UpsertRegistration(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid payload", func(t *testing.T) {
		svc := &fakeCompanyService{}
		h := company.NewHandler(svc, logger)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := httptest.NewRequest(http.MethodPost, "/registrations", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")

		c.Request = req
		c.Set("company_id", uuid.New().String())

		h.UpsertRegistration(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("conflict error", func(t *testing.T) {
		compID := uuid.New().String()

		svc := &fakeCompanyService{
			UpsertRegistrationFn: func(ctx context.Context, cid string, req company.UpsertCompanyRegistrationRequest) error {
				return companyerrors.ErrRegistrationAlreadyExists
			},
		}

		h := company.NewHandler(svc, logger)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := `{"type":"npwp","number":"123456"}`
		req := httptest.NewRequest(http.MethodPost, "/registrations", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		c.Request = req
		c.Set("company_id", compID)

		h.UpsertRegistration(c)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), apperror.CodeConflict)
	})
}

//
// ==============================
// GET REGISTRATIONS
// ==============================
//

func TestCompanyHandler_ListRegistrations(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		compID := uuid.New().String()

		svc := &fakeCompanyService{
			ListRegistrationsFn: func(ctx context.Context, cid string) ([]company.CompanyRegistrationResponse, error) {
				return []company.CompanyRegistrationResponse{
					{Type: "npwp", Number: "123"},
				}, nil
			},
		}

		h, w, c := setupHandlerTest(t, svc)

		req := httptest.NewRequest(http.MethodGet, "/registrations", nil)
		c.Request = req
		c.Set("company_id", compID)

		h.ListRegistrations(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "npwp")
	})

	t.Run("service error", func(t *testing.T) {
		compID := uuid.New().String()

		svc := &fakeCompanyService{
			ListRegistrationsFn: func(ctx context.Context, cid string) ([]company.CompanyRegistrationResponse, error) {
				return nil, errors.New("database error")
			},
		}

		h, w, c := setupHandlerTest(t, svc)

		req := httptest.NewRequest(http.MethodGet, "/registrations", nil)
		c.Request = req
		c.Set("company_id", compID)

		h.ListRegistrations(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "error")
	})

	t.Run("missing company id", func(t *testing.T) {
		svc := &fakeCompanyService{}
		h, w, c := setupHandlerTest(t, svc)

		req := httptest.NewRequest(http.MethodGet, "/registrations", nil)
		c.Request = req
		// jangan set company_id sama sekali

		h.ListRegistrations(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

//
// ==============================
// DELETE REGISTRATION
// ==============================
//

func TestCompanyHandler_DeleteRegistration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		compID := uuid.New().String()

		svc := &fakeCompanyService{
			DeleteRegistrationFn: func(ctx context.Context, cid string, regType company.RegistrationType) error {
				assert.Equal(t, company.RegistrationType("npwp"), regType)
				return nil
			},
		}

		h, w, c := setupHandlerTest(t, svc)

		req := httptest.NewRequest(http.MethodDelete, "/registrations/npwp", nil)
		c.Params = []gin.Param{{Key: "type", Value: "npwp"}}

		c.Request = req
		c.Set("company_id", compID)

		h.DeleteRegistration(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		compID := uuid.New().String()

		svc := &fakeCompanyService{
			DeleteRegistrationFn: func(ctx context.Context, cid string, regType company.RegistrationType) error {
				return errors.New("delete failed")
			},
		}
		h, w, c := setupHandlerTest(t, svc)

		req := httptest.NewRequest(http.MethodDelete, "/registrations/npwp", nil)
		c.Params = []gin.Param{{Key: "type", Value: "npwp"}}

		c.Request = req
		c.Set("company_id", compID)

		h.DeleteRegistration(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("missing registration type param", func(t *testing.T) {
		compID := uuid.New().String()

		svc := &fakeCompanyService{}
		h, w, c := setupHandlerTest(t, svc)

		req := httptest.NewRequest(http.MethodDelete, "/registrations/", nil)
		c.Request = req
		c.Set("company_id", compID)

		// intentionally not setting c.Params

		h.DeleteRegistration(c)

		assert.NotEqual(t, http.StatusOK, w.Code)
	})
}
