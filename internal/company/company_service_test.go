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

func TestService_GetByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := companyMock.NewMockRepository(ctrl)
	service := company.NewService(mockRepo)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		id := uuid.New()
		mockComp := &company.Company{
			ID:                 id,
			Name:               "Test Company",
			Email:              "test@company.com",
			RegistrationNumber: "REG123",
			IsActive:           true,
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

func TestService_Update(t *testing.T) {
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
