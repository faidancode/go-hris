package employee_test

import (
	"context"
	"database/sql"
	"errors"
	"go-hris/internal/employee"
	"testing"

	employeeMock "go-hris/internal/employee/mock"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type serviceDeps struct {
	db      *sql.DB
	sqlMock sqlmock.Sqlmock
	service employee.Service
	repo    *employeeMock.MockRepository
}

func setupServiceTest(t *testing.T) *serviceDeps {
	ctrl := gomock.NewController(t)

	db, sqlMock, _ := sqlmock.New()
	repo := employeeMock.NewMockRepository(ctrl)

	svc := employee.NewService(db, repo)

	return &serviceDeps{
		db:      db,
		sqlMock: sqlMock,
		service: svc,
		repo:    repo,
	}
}
func expectTx(t *testing.T, mock sqlmock.Sqlmock, commit bool) {
	t.Helper()
	mock.ExpectBegin()
	if commit {
		mock.ExpectCommit()
	} else {
		mock.ExpectRollback()
	}
}
func TestEmployeeService_Create(t *testing.T) {
	deps := setupServiceTest(t)
	defer deps.db.Close()

	ctx := context.Background()

	req := employee.CreateEmployeeRequest{
		Name:      "John",
		Email:     "john@test.com",
		CompanyID: "company-1",
	}

	t.Run("success", func(t *testing.T) {
		expectTx(t, deps.sqlMock, true)

		deps.repo.EXPECT().
			WithTx(gomock.Any()).
			Return(deps.repo)

		deps.repo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(employee.Employee{
				ID:        uuid.MustParse("uuid-1"),
				Name:      req.Name,
				Email:     req.Email,
				CompanyID: uuid.MustParse(req.CompanyID),
			}, nil)

		res, err := deps.service.Create(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, "uuid-1", res.ID)
	})

	t.Run("create failed", func(t *testing.T) {
		expectTx(t, deps.sqlMock, false)

		deps.repo.EXPECT().
			WithTx(gomock.Any()).
			Return(deps.repo)

		deps.repo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(employee.Employee{}, errors.New("db error"))

		_, err := deps.service.Create(ctx, req)

		assert.Error(t, err)
	})
}
