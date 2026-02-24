package user_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-hris/internal/user"
	mock_user "go-hris/internal/user/mock"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func setup(t *testing.T) (*mock_user.MockRepository, user.Service) {
	ctrl := gomock.NewController(t)
	mockRepo := mock_user.NewMockRepository(ctrl)
	svc := user.NewService(mockRepo)
	return mockRepo, svc
}

func TestUserService_GetAll(t *testing.T) {
	ctx := context.Background()
	companyID := uuid.New().String()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mock_user.NewMockRepository(ctrl)
		svc := user.NewService(mockRepo)

		mockRepo.EXPECT().
			FindAllByCompany(gomock.Any(), companyID).
			Return([]user.User{
				{
					ID:       uuid.New(),
					Email:    "john@mail.com",
					IsActive: true,
				},
			}, nil)

		res, err := svc.GetAll(ctx, companyID)

		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, "john@mail.com", res[0].Email)
	})

	t.Run("repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mock_user.NewMockRepository(ctrl)
		svc := user.NewService(mockRepo)

		mockRepo.EXPECT().
			FindAllByCompany(gomock.Any(), companyID).
			Return(nil, errors.New("db error"))

		res, err := svc.GetAll(ctx, companyID)

		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestUserService_GetByID(t *testing.T) {
	ctx := context.Background()
	companyID := uuid.New().String()
	userID := uuid.New().String()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mock_user.NewMockRepository(ctrl)
		svc := user.NewService(mockRepo)

		mockRepo.EXPECT().
			FindByID(gomock.Any(), companyID, userID).
			Return(&user.User{
				ID:       uuid.MustParse(userID),
				Email:    "john@mail.com",
				IsActive: true,
			}, nil)

		res, err := svc.GetByID(ctx, companyID, userID)

		assert.NoError(t, err)
		assert.Equal(t, userID, res.ID)
		assert.Equal(t, "john@mail.com", res.Email)
	})

	t.Run("repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mock_user.NewMockRepository(ctrl)
		svc := user.NewService(mockRepo)

		mockRepo.EXPECT().
			FindByID(gomock.Any(), companyID, userID).
			Return(nil, errors.New("db error"))

		res, err := svc.GetByID(ctx, companyID, userID)

		assert.Error(t, err)
		assert.Equal(t, user.UserResponse{}, res)
	})
}

func TestUserService_Create(t *testing.T) {
	ctx := context.Background()
	companyID := uuid.New().String()
	req := user.CreateUserRequest{
		EmployeeID: uuid.New().String(),
		Email:      "john@mail.com",
		Password:   "password123",
	}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mock_user.NewMockRepository(ctrl)
		svc := user.NewService(mockRepo)

		mockRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(nil)

		res, err := svc.Create(ctx, companyID, req)

		assert.NoError(t, err)
		assert.Equal(t, req.Email, res.Email)
		assert.True(t, res.IsActive)
	})

	t.Run("repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		companyID := uuid.New().String()
		mockRepo := mock_user.NewMockRepository(ctrl)
		svc := user.NewService(mockRepo)

		mockRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(errors.New("db error"))

		res, err := svc.Create(ctx, companyID, req)

		assert.Error(t, err)
		assert.Equal(t, user.UserResponse{}, res)
	})
}

func TestUserService_GetCompanyUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_user.NewMockRepository(ctrl)
	svc := user.NewService(mockRepo)

	ctx := context.Background()
	companyID := uuid.New().String()

	t.Run("success", func(t *testing.T) {
		users := []user.User{
			{
				ID:         uuid.New(),
				EmployeeID: uuid.New(),
				Email:      "user1@mail.com",
				IsActive:   true,
				CreatedAt:  time.Now(),
			},
		}

		mockRepo.EXPECT().
			FindAllByCompany(ctx, companyID).
			Return(users, nil)

		res, err := svc.GetCompanyUsers(ctx, companyID)

		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, users[0].Email, res[0].Email)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.EXPECT().
			FindAllByCompany(ctx, companyID).
			Return(nil, errors.New("db error"))

		res, err := svc.GetCompanyUsers(ctx, companyID)

		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestUserService_ToggleStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_user.NewMockRepository(ctrl)
	svc := user.NewService(mockRepo)

	ctx := context.Background()
	companyID := uuid.New().String()
	userID := uuid.New().String()

	t.Run("success", func(t *testing.T) {
		u := &user.User{
			ID:         uuid.MustParse(userID),
			EmployeeID: uuid.New(),
			Email:      "user@mail.com",
			IsActive:   true,
		}

		mockRepo.EXPECT().
			FindByID(ctx, companyID, userID).
			Return(u, nil)

		mockRepo.EXPECT().
			Update(ctx, u).
			Return(nil)

		err := svc.ToggleStatus(ctx, companyID, userID, false)

		assert.NoError(t, err)
		assert.False(t, u.IsActive)
	})

	t.Run("find error", func(t *testing.T) {
		mockRepo.EXPECT().
			FindByID(ctx, companyID, userID).
			Return(nil, errors.New("not found"))

		err := svc.ToggleStatus(ctx, companyID, userID, false)

		assert.Error(t, err)
	})

	t.Run("update error", func(t *testing.T) {
		u := &user.User{
			ID:         uuid.MustParse(userID),
			EmployeeID: uuid.New(),
			Email:      "user@mail.com",
			IsActive:   true,
		}

		mockRepo.EXPECT().
			FindByID(ctx, companyID, userID).
			Return(u, nil)

		mockRepo.EXPECT().
			Update(ctx, u).
			Return(errors.New("update failed"))

		err := svc.ToggleStatus(ctx, companyID, userID, false)

		assert.Error(t, err)
	})
}

func TestUserService_ChangePassword(t *testing.T) {
	mockRepo, svc := setup(t)

	ctx := context.Background()
	companyID := uuid.New().String()
	userID := uuid.New().String()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("oldpass"), bcrypt.DefaultCost)

	t.Run("success", func(t *testing.T) {
		u := &user.User{Password: string(hashed)}

		mockRepo.EXPECT().FindByID(ctx, companyID, userID).Return(u, nil)
		mockRepo.EXPECT().Update(ctx, u).Return(nil)

		err := svc.ChangePassword(ctx, companyID, userID, "oldpass", "newpass")
		assert.NoError(t, err)
	})

	t.Run("wrong current password", func(t *testing.T) {
		u := &user.User{Password: string(hashed)}

		mockRepo.EXPECT().FindByID(ctx, companyID, userID).Return(u, nil)

		err := svc.ChangePassword(ctx, companyID, userID, "wrong", "newpass")
		assert.Error(t, err)
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo.EXPECT().FindByID(ctx, companyID, userID).
			Return(nil, errors.New("not found"))

		err := svc.ChangePassword(ctx, companyID, userID, "oldpass", "newpass")
		assert.Error(t, err)
	})
}

func TestUserService_ResetPassword(t *testing.T) {
	mockRepo, svc := setup(t)

	ctx := context.Background()
	companyID := uuid.New().String()
	userID := uuid.New().String()

	t.Run("success", func(t *testing.T) {
		u := &user.User{}

		mockRepo.EXPECT().FindByID(ctx, companyID, userID).Return(u, nil)
		mockRepo.EXPECT().Update(ctx, u).Return(nil)

		err := svc.ResetPassword(ctx, companyID, userID, "newpass")
		assert.NoError(t, err)
	})

	t.Run("find error", func(t *testing.T) {
		mockRepo.EXPECT().FindByID(ctx, companyID, userID).
			Return(nil, errors.New("not found"))

		err := svc.ResetPassword(ctx, companyID, userID, "newpass")
		assert.Error(t, err)
	})
}

func TestUserService_ForceResetPassword(t *testing.T) {
	mockRepo, svc := setup(t)

	ctx := context.Background()
	companyID := uuid.New().String()
	userID := uuid.New().String()

	t.Run("success", func(t *testing.T) {
		u := &user.User{}

		mockRepo.EXPECT().FindByID(ctx, companyID, userID).Return(u, nil)
		mockRepo.EXPECT().Update(ctx, u).Return(nil)

		err := svc.ForceResetPassword(ctx, companyID, userID, "newpass")
		assert.NoError(t, err)
	})
}
