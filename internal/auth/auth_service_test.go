package auth_test

import (
	"context"
	"errors"
	"go-hris/internal/auth"
	autherrors "go-hris/internal/auth/errors"
	authMock "go-hris/internal/auth/mock"
	"go-hris/internal/employee"
	employeeerrors "go-hris/internal/employee/errors"
	employeeMock "go-hris/internal/employee/mock"
	rbacMock "go-hris/internal/rbac/mock"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestService_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := authMock.NewMockRepository(ctrl)
	mockRBAC := rbacMock.NewMockService(ctrl)
	mockEmployeeRepo := employeeMock.NewMockRepository(ctrl)

	service := auth.NewService(mockRepo, mockRBAC, mockEmployeeRepo)
	ctx := context.Background()

	password := "password123"
	pw, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	// Mock Data
	userID := uuid.New()
	companyID := uuid.New()
	employeeID := uuid.New()
	mockUser := &auth.User{
		ID:         userID,
		EmployeeID: &employeeID,
		CompanyID:  companyID,
		Email:      "admin@example.com",
		Password:   string(pw),
		Role:       "EMPLOYEE",
	}

	t.Run("Success Login", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByEmail(ctx, mockUser.Email).
			Return(mockUser, nil)

			// Setup EXPECT untuk RBAC
		mockRBAC.EXPECT().
			LoadCompanyPolicy(companyID.String()).
			Return(nil) // penting! kalau tidak, panic seperti error tadi

		token, refreshToken, resp, err := service.Login(ctx, mockUser.Email, password)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.NotEmpty(t, refreshToken)
		assert.Equal(t, mockUser.Email, resp.Email)
		assert.Equal(t, companyID.String(), resp.CompanyID)
	})

	t.Run("Wrong Password", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByEmail(ctx, mockUser.Email).
			Return(mockUser, nil)

		_, _, _, err := service.Login(ctx, mockUser.Email, "wrongpass")
		assert.Error(t, err)
	})

}

func TestService_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := authMock.NewMockRepository(ctrl)
	mockRBAC := rbacMock.NewMockService(ctrl)
	mockEmployeeRepo := employeeMock.NewMockRepository(ctrl)
	service := auth.NewService(mockRepo, mockRBAC, mockEmployeeRepo)
	ctx := context.Background()

	t.Run("Success Register", func(t *testing.T) {
		cID := uuid.New()
		eID := uuid.New()

		req := auth.RegisterRequest{
			CompanyID:  cID.String(),
			EmployeeID: eID.String(),
			Email:      "user@example.com",
			Name:       "John Doe",
			Password:   "password123",
		}

		// Mock Find Employee: Sesuaikan parameter ID menjadi string sesuai logic service
		mockEmployeeRepo.EXPECT().
			FindByIDAndCompany(ctx, req.CompanyID, req.EmployeeID).
			Return(&employee.Employee{
				ID:        eID,
				CompanyID: cID,
				FullName:  "John Doe",
			}, nil)

		mockRepo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(nil)

		mockRBAC.EXPECT().
			LoadCompanyPolicy(cID.String()).
			Return(nil).
			Times(1)

		resp, err := service.Register(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, req.Email, resp.Email)
		assert.Equal(t, "Employee", resp.Role)
		assert.Equal(t, cID.String(), resp.CompanyID)
	})

	t.Run("Employee Not Found", func(t *testing.T) {
		cID := uuid.New().String()
		eID := uuid.New().String()
		req := auth.RegisterRequest{
			CompanyID:  cID, // Tambahkan CompanyID agar tidak nil
			EmployeeID: eID,
			Password:   "password123",
		}

		// Gunakan variabel string langsung
		mockEmployeeRepo.EXPECT().
			FindByIDAndCompany(ctx, cID, eID).
			Return(nil, errors.New("not found"))

		_, err := service.Register(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, employeeerrors.ErrEmployeeNotFound, err)
	})

	t.Run("Error Register - Duplicate Email", func(t *testing.T) {
		cID := uuid.New()
		eID := uuid.New()
		req := auth.RegisterRequest{
			CompanyID:  cID.String(),
			EmployeeID: eID.String(),
			Email:      "duplicate@example.com",
			Password:   "password123",
		}

		mockEmployeeRepo.EXPECT().
			FindByIDAndCompany(ctx, req.CompanyID, req.EmployeeID).
			Return(&employee.Employee{ID: eID, CompanyID: cID}, nil)

		mockRepo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(errors.New("duplicate key error"))

		_, err := service.Register(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, autherrors.ErrEmailAlreadyRegistered, err)
	})
}
