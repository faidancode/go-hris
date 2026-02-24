package user_test

import (
	"context"
	"errors"
	"go-hris/internal/shared/apperror"
	"go-hris/internal/user"
	usererrors "go-hris/internal/user/errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type fakeUserService struct {
	GetAllFn             func(ctx context.Context, companyID string) ([]user.UserResponse, error)
	GetAllWithRolesFn    func(ctx context.Context, companyID string) ([]user.UserWithRolesResponse, error)
	GetByIDFn            func(ctx context.Context, companyID, id string) (user.UserResponse, error)
	CreateFn             func(ctx context.Context, companyID string, req user.CreateUserRequest) (user.UserResponse, error)
	GetCompanyUsersFn    func(ctx context.Context, companyID string) ([]user.UserResponse, error)
	AssignRoleFn         func(ctx context.Context, companyID, userID, roleName string) error
	ToggleStatusFn       func(ctx context.Context, companyID, id string, isActive bool) error
	ChangePasswordFn     func(ctx context.Context, companyID, id, current, new string) error
	ResetPasswordFn      func(ctx context.Context, companyID, id, new string) error
	ForceResetPasswordFn func(ctx context.Context, companyID, id, new string) error
}

func (f *fakeUserService) GetAll(ctx context.Context, cid string) ([]user.UserResponse, error) {
	return f.GetAllFn(ctx, cid)
}

func (f *fakeUserService) GetByID(ctx context.Context, cid, id string) (user.UserResponse, error) {
	return f.GetByIDFn(ctx, cid, id)
}

func (f *fakeUserService) GetAllWithRoles(ctx context.Context, cid string) ([]user.UserWithRolesResponse, error) {
	return f.GetAllWithRolesFn(ctx, cid)
}

func (f *fakeUserService) Create(ctx context.Context, cid string, req user.CreateUserRequest) (user.UserResponse, error) {
	return f.CreateFn(ctx, cid, req)
}
func (f *fakeUserService) GetCompanyUsers(ctx context.Context, cid string) ([]user.UserResponse, error) {
	return f.GetCompanyUsersFn(ctx, cid)
}
func (f *fakeUserService) AssignRole(ctx context.Context, cid, userID, roleName string) error {
	return f.AssignRoleFn(ctx, cid, userID, roleName)
}
func (f *fakeUserService) ToggleStatus(ctx context.Context, cid, id string, isActive bool) error {
	return f.ToggleStatusFn(ctx, cid, id, isActive)
}
func (f *fakeUserService) ChangePassword(ctx context.Context, cid, id, current, new string) error {
	return f.ChangePasswordFn(ctx, cid, id, current, new)
}
func (f *fakeUserService) ResetPassword(ctx context.Context, cid, id, new string) error {
	return f.ResetPasswordFn(ctx, cid, id, new)
}
func (f *fakeUserService) ForceResetPassword(ctx context.Context, cid, id, new string) error {
	return f.ForceResetPasswordFn(ctx, cid, id, new)
}

func setupHandler(svc user.Service) *user.Handler {
	return user.NewHandler(svc, zap.NewNop())
}

func TestUserHandler_GetAll(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		companyID := uuid.New().String()

		svc := &fakeUserService{
			GetAllFn: func(ctx context.Context, cid string) ([]user.UserResponse, error) {
				assert.Equal(t, companyID, cid)
				return []user.UserResponse{
					{ID: uuid.New().String(), Email: "user@mail.com"},
				}, nil
			},
		}

		h := setupHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		c.Request = req
		c.Set("company_id", companyID)

		h.GetAll(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "user@mail.com")
	})

	t.Run("service error", func(t *testing.T) {
		svc := &fakeUserService{
			GetAllFn: func(ctx context.Context, cid string) ([]user.UserResponse, error) {
				return nil, errors.New("service error")
			},
		}

		h := setupHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		c.Request = req
		c.Set("company_id", uuid.New().String())

		h.GetAll(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestUserHandler_GetByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		companyID := uuid.New().String()
		userID := uuid.New().String()

		svc := &fakeUserService{
			GetByIDFn: func(ctx context.Context, cid, id string) (user.UserResponse, error) {
				assert.Equal(t, companyID, cid)
				assert.Equal(t, userID, id)
				return user.UserResponse{
					ID:    id,
					Email: "user@mail.com",
				}, nil
			},
		}

		h := setupHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := httptest.NewRequest(http.MethodGet, "/users/"+userID, nil)
		c.Request = req
		c.Set("company_id", companyID)
		c.Params = gin.Params{
			{Key: "id", Value: userID},
		}

		h.GetById(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "user@mail.com")
	})

	t.Run("service error", func(t *testing.T) {
		svc := &fakeUserService{
			GetByIDFn: func(ctx context.Context, cid, id string) (user.UserResponse, error) {
				return user.UserResponse{}, errors.New("not found")
			},
		}

		h := setupHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
		c.Request = req
		c.Set("company_id", uuid.New().String())
		c.Params = gin.Params{
			{Key: "id", Value: "1"},
		}

		h.GetById(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestUserHandler_Create(t *testing.T) {

	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		companyID := uuid.New().String()
		employeeID := uuid.New().String()

		svc := &fakeUserService{
			CreateFn: func(ctx context.Context, cid string, req user.CreateUserRequest) (user.UserResponse, error) {
				assert.Equal(t, companyID, cid)
				assert.Equal(t, "admin@example.com", req.Email)
				assert.Equal(t, employeeID, req.EmployeeID) // ✅ tambahan assert
				return user.UserResponse{
					ID:         uuid.New().String(),
					Email:      req.Email,
					EmployeeID: req.EmployeeID,
				}, nil
			},
		}

		h := setupHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := `{
			"email":"admin@example.com",
			"password":"12345678",
			"employee_id":"` + employeeID + `"
		}`

		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		c.Request = req
		c.Set("company_id", companyID)

		h.Create(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "admin@example.com")
	})

	t.Run("validation error", func(t *testing.T) {
		h := setupHandler(&fakeUserService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")

		c.Request = req
		c.Set("company_id", uuid.New().String())

		h.Create(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("duplicate email returns conflict", func(t *testing.T) {
		svc := &fakeUserService{
			CreateFn: func(ctx context.Context, cid string, req user.CreateUserRequest) (user.UserResponse, error) {
				return user.UserResponse{}, usererrors.ErrUserAlreadyExists
			},
		}

		h := setupHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := `{
			"email":"admin@example.com",
			"password":"12345678",
			"employee_id":"` + uuid.New().String() + `"
		}`

		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		c.Request = req
		c.Set("company_id", uuid.New().String())

		h.Create(c)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), apperror.CodeConflict)
	})
}

func TestUserHandler_GetCompanyUsers(t *testing.T) {

	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		companyID := uuid.New().String()

		svc := &fakeUserService{
			GetCompanyUsersFn: func(ctx context.Context, cid string) ([]user.UserResponse, error) {
				assert.Equal(t, companyID, cid)
				return []user.UserResponse{
					{ID: uuid.New().String(), Email: "a@company.com"},
					{ID: uuid.New().String(), Email: "b@company.com"},
				}, nil
			},
		}

		h := setupHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(http.MethodGet, "/users", nil)
		c.Set("company_id", companyID)

		h.GetCompanyUsers(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "a@company.com")
		assert.Contains(t, w.Body.String(), "b@company.com")
	})
}

func TestUserHandler_ToggleStatus(t *testing.T) {

	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		companyID := uuid.New().String()
		userID := uuid.New().String()

		svc := &fakeUserService{
			ToggleStatusFn: func(ctx context.Context, cid, id string, active bool) error {
				assert.Equal(t, companyID, cid)
				assert.Equal(t, userID, id)
				assert.True(t, active)
				return nil
			},
		}

		h := setupHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := `{"is_active":true}`
		req := httptest.NewRequest(http.MethodPatch, "/users/"+userID, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		c.Request = req
		c.Params = []gin.Param{{Key: "id", Value: userID}}
		c.Set("company_id", companyID)

		h.ToggleStatus(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestUserHandler_ChangePassword(t *testing.T) {

	gin.SetMode(gin.TestMode)

	t.Run("wrong password", func(t *testing.T) {
		svc := &fakeUserService{
			ChangePasswordFn: func(ctx context.Context, cid, id, current, new string) error {
				return usererrors.ErrWrongPassword
			},
		}

		h := setupHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := httptest.NewRequest(http.MethodPost, "/users/123/password",
			strings.NewReader(`{"current_password":"x","new_password":"y"}`))
		req.Header.Set("Content-Type", "application/json")

		c.Request = req
		c.Params = []gin.Param{{Key: "id", Value: "123"}}
		c.Set("company_id", uuid.New().String())

		h.ChangePassword(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUserHandler_ResetPassword(t *testing.T) {

	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := &fakeUserService{
			ResetPasswordFn: func(ctx context.Context, cid, id, new string) error {
				return nil
			},
		}

		h := setupHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := httptest.NewRequest(http.MethodPost, "/users/123/reset",
			strings.NewReader(`{"new_password":"newpass"}`))
		req.Header.Set("Content-Type", "application/json")

		c.Request = req
		c.Params = []gin.Param{{Key: "id", Value: "123"}}
		c.Set("company_id", uuid.New().String())

		h.ResetPassword(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestUserHandler_ForceResetPassword(t *testing.T) {

	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := &fakeUserService{
			ForceResetPasswordFn: func(ctx context.Context, cid, id, new string) error {
				return nil
			},
		}

		h := setupHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := httptest.NewRequest(http.MethodPost, "/users/123/force-reset",
			strings.NewReader(`{"new_password":"newpass"}`))
		req.Header.Set("Content-Type", "application/json")

		c.Request = req
		c.Params = []gin.Param{{Key: "id", Value: "123"}}
		c.Set("company_id", uuid.New().String())

		h.ForceResetPassword(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("validation error", func(t *testing.T) {
		svc := &fakeUserService{}

		h := setupHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := httptest.NewRequest(http.MethodPost, "/users/123/force-reset",
			strings.NewReader(`{}`)) // missing new_password
		req.Header.Set("Content-Type", "application/json")

		c.Request = req
		c.Params = []gin.Param{{Key: "id", Value: "123"}}
		c.Set("company_id", uuid.New().String())

		h.ForceResetPassword(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		svc := &fakeUserService{
			ForceResetPasswordFn: func(ctx context.Context, cid, id, new string) error {
				return errors.New("db error")
			},
		}

		h := setupHandler(svc)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := httptest.NewRequest(http.MethodPost, "/users/123/force-reset",
			strings.NewReader(`{"new_password":"newpass"}`))
		req.Header.Set("Content-Type", "application/json")

		c.Request = req
		c.Params = []gin.Param{{Key: "id", Value: "123"}}
		c.Set("company_id", uuid.New().String())

		h.ForceResetPassword(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
