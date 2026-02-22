package company_test

import (
	"context"
	"errors"
	"go-hris/internal/company"
	companyMock "go-hris/internal/company/mock"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCompanyService_GetByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := companyMock.NewMockRepository(ctrl)
	service := company.NewService(mockRepo)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		id := uuid.New()
		mockComp := &company.Company{
			ID:       id,
			Name:     "Test Company",
			Email:    "test@company.com",
			IsActive: true,
		}

		mockRepo.EXPECT().GetByID(ctx, id).Return(mockComp, nil)

		resp, err := service.GetByID(ctx, id.String())

		assert.NoError(t, err)
		assert.Equal(t, mockComp.Name, resp.Name)
		assert.Equal(t, mockComp.ID.String(), resp.ID)
	})

	t.Run("Not Found", func(t *testing.T) {
		id := uuid.New()
		mockRepo.EXPECT().GetByID(ctx, id).Return(nil, errors.New("not found"))

		_, err := service.GetByID(ctx, id.String())
		assert.Error(t, err)
	})
}

func TestCompanyService_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := companyMock.NewMockRepository(ctrl)
	service := company.NewService(mockRepo)
	ctx := context.Background()

	t.Run("Success Update Name", func(t *testing.T) {
		id := uuid.New()
		mockComp := &company.Company{
			ID:       id,
			Name:     "Old Name",
			Email:    "test@company.com",
			IsActive: true,
		}

		mockRepo.EXPECT().GetByID(ctx, id).Return(mockComp, nil)
		mockRepo.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, c *company.Company) error {
			assert.Equal(t, "New Name", c.Name)
			return nil
		})

		resp, err := service.Update(ctx, id.String(), company.UpdateCompanyRequest{
			Name: "New Name",
		})

		assert.NoError(t, err)
		assert.Equal(t, "New Name", resp.Name)
	})
}

func TestCompanyService_UpsertRegistration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := companyMock.NewMockRepository(ctrl)
	service := company.NewService(mockRepo)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		companyID := uuid.New()
		req := company.UpsertCompanyRegistrationRequest{
			Type:   company.RegistrationTypeNPWP,
			Number: "123456789",
		}

		mockRepo.EXPECT().
			UpsertRegistration(ctx, gomock.Any()).
			Return(nil)

		err := service.UpsertRegistration(ctx, companyID.String(), req)
		assert.NoError(t, err)
	})

	t.Run("Invalid Company ID", func(t *testing.T) {
		req := company.UpsertCompanyRegistrationRequest{
			Type:   company.RegistrationTypeNPWP,
			Number: "123456789",
		}

		err := service.UpsertRegistration(ctx, "invalid-uuid", req)
		assert.Error(t, err)
	})

	t.Run("Invalid Registration Type", func(t *testing.T) {
		companyID := uuid.New()
		req := company.UpsertCompanyRegistrationRequest{
			Type:   "",
			Number: "123456789",
		}

		err := service.UpsertRegistration(ctx, companyID.String(), req)
		assert.Error(t, err)
	})

	t.Run("Repo Error", func(t *testing.T) {
		companyID := uuid.New()
		req := company.UpsertCompanyRegistrationRequest{
			Type:   company.RegistrationTypeNPWP,
			Number: "123456789",
		}

		mockRepo.EXPECT().
			UpsertRegistration(ctx, gomock.Any()).
			Return(errors.New("db error"))

		err := service.UpsertRegistration(ctx, companyID.String(), req)
		assert.Error(t, err)
	})
}

func TestCompanyService_ListRegistrations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := companyMock.NewMockRepository(ctrl)
	service := company.NewService(mockRepo)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		companyID := uuid.New()
		regID := uuid.New()

		mockData := []company.CompanyRegistration{
			{
				ID:        regID,
				CompanyID: companyID,
				Type:      company.RegistrationTypeNPWP,
				Number:    "123456789",
			},
		}

		mockRepo.EXPECT().
			GetRegistrationsByCompanyID(ctx, companyID).
			Return(mockData, nil)

		resp, err := service.ListRegistrations(ctx, companyID.String())

		assert.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, regID.String(), resp[0].ID)
		assert.Equal(t, company.RegistrationTypeNPWP, resp[0].Type)
		assert.Equal(t, "123456789", resp[0].Number)
	})

	t.Run("Invalid Company ID", func(t *testing.T) {
		_, err := service.ListRegistrations(ctx, "invalid-uuid")
		assert.Error(t, err)
	})

	t.Run("Repo Error", func(t *testing.T) {
		companyID := uuid.New()

		mockRepo.EXPECT().
			GetRegistrationsByCompanyID(ctx, companyID).
			Return(nil, errors.New("db error"))

		_, err := service.ListRegistrations(ctx, companyID.String())
		assert.Error(t, err)
	})

	t.Run("Empty Result", func(t *testing.T) {
		companyID := uuid.New()

		mockRepo.EXPECT().
			GetRegistrationsByCompanyID(ctx, companyID).
			Return([]company.CompanyRegistration{}, nil)

		resp, err := service.ListRegistrations(ctx, companyID.String())

		assert.NoError(t, err)
		assert.Len(t, resp, 0)
	})
}

func TestCompanyService_DeleteRegistration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := companyMock.NewMockRepository(ctrl)
	service := company.NewService(mockRepo)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		companyID := uuid.New()

		mockRepo.EXPECT().
			DeleteRegistration(ctx, companyID, company.RegistrationTypeNPWP).
			Return(nil)

		err := service.DeleteRegistration(
			ctx,
			companyID.String(),
			company.RegistrationTypeNPWP,
		)

		assert.NoError(t, err)
	})

	t.Run("Invalid Company ID", func(t *testing.T) {
		err := service.DeleteRegistration(
			ctx,
			"invalid-uuid",
			company.RegistrationTypeNPWP,
		)

		assert.Error(t, err)
	})

	t.Run("Invalid Registration Type", func(t *testing.T) {
		companyID := uuid.New()

		err := service.DeleteRegistration(
			ctx,
			companyID.String(),
			"",
		)

		assert.Error(t, err)
	})

	t.Run("Repo Error", func(t *testing.T) {
		companyID := uuid.New()

		mockRepo.EXPECT().
			DeleteRegistration(ctx, companyID, company.RegistrationTypeNPWP).
			Return(errors.New("db error"))

		err := service.DeleteRegistration(
			ctx,
			companyID.String(),
			company.RegistrationTypeNPWP,
		)

		assert.Error(t, err)
	})
}
