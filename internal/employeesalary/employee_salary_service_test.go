package employeesalary_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"go-hris/internal/employeesalary"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type fakeSalaryRepository struct {
	withTxFn             func(tx *sql.Tx) employeesalary.Repository
	createFn             func(ctx context.Context, salary *employeesalary.EmployeeSalary) error
	findAllByCompanyFn   func(ctx context.Context, companyID string) ([]employeesalary.EmployeeSalary, error)
	findByIDAndCompanyFn func(ctx context.Context, companyID string, id string) (*employeesalary.EmployeeSalary, error)
	deleteFn             func(ctx context.Context, companyID string, id string) error
}

func (f *fakeSalaryRepository) WithTx(tx *sql.Tx) employeesalary.Repository {
	if f.withTxFn != nil {
		return f.withTxFn(tx)
	}
	return f
}

func (f *fakeSalaryRepository) Create(ctx context.Context, salary *employeesalary.EmployeeSalary) error {
	if f.createFn != nil {
		return f.createFn(ctx, salary)
	}
	return nil
}

func (f *fakeSalaryRepository) FindAllByCompany(ctx context.Context, companyID string) ([]employeesalary.EmployeeSalary, error) {
	if f.findAllByCompanyFn != nil {
		return f.findAllByCompanyFn(ctx, companyID)
	}
	return nil, nil
}

func (f *fakeSalaryRepository) FindByIDAndCompany(ctx context.Context, companyID string, id string) (*employeesalary.EmployeeSalary, error) {
	if f.findByIDAndCompanyFn != nil {
		return f.findByIDAndCompanyFn(ctx, companyID, id)
	}
	return nil, nil
}

func (f *fakeSalaryRepository) Delete(ctx context.Context, companyID string, id string) error {
	if f.deleteFn != nil {
		return f.deleteFn(ctx, companyID, id)
	}
	return nil
}

type serviceDeps struct {
	db      *sql.DB
	sqlMock sqlmock.Sqlmock
	service employeesalary.Service
	repo    *fakeSalaryRepository
}

func setupServiceTest(t *testing.T) *serviceDeps {
	t.Helper()

	db, sqlMock, err := sqlmock.New()
	assert.NoError(t, err)

	repo := &fakeSalaryRepository{}
	svc := employeesalary.NewService(db, repo)

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

func TestEmployeeSalaryService_Create(t *testing.T) {
	deps := setupServiceTest(t)
	defer deps.db.Close()

	ctx := context.Background()
	companyID := uuid.New().String()
	employeeID := uuid.New()

	t.Run("success", func(t *testing.T) {
		req := employeesalary.CreateEmployeeSalaryRequest{
			EmployeeID:    employeeID.String(),
			BaseSalary:    9000000,
			EffectiveDate: "2026-02-01",
		}

		expectTx(t, deps.sqlMock, true)

		deps.repo.createFn = func(ctx context.Context, salary *employeesalary.EmployeeSalary) error {
			assert.Equal(t, employeeID, salary.EmployeeID)
			assert.Equal(t, 9000000, salary.BaseSalary)
			assert.Equal(t, "2026-02-01", salary.EffectiveDate.Format("2006-01-02"))
			return nil
		}
		deps.repo.findByIDAndCompanyFn = func(ctx context.Context, cid string, id string) (*employeesalary.EmployeeSalary, error) {
			assert.Equal(t, companyID, cid)
			return &employeesalary.EmployeeSalary{
				ID:            uuid.MustParse(id),
				EmployeeID:    employeeID,
				BaseSalary:    9000000,
				EffectiveDate: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			}, nil
		}

		resp, err := deps.service.Create(ctx, companyID, req)

		assert.NoError(t, err)
		assert.Equal(t, employeeID.String(), resp.EmployeeID)
		assert.Equal(t, 9000000, resp.BaseSalary)
		assert.Equal(t, "2026-02-01", resp.EffectiveDate)
		assert.NoError(t, deps.sqlMock.ExpectationsWereMet())
	})

	t.Run("invalid employee id", func(t *testing.T) {
		req := employeesalary.CreateEmployeeSalaryRequest{
			EmployeeID:    "invalid-uuid",
			BaseSalary:    9000000,
			EffectiveDate: "2026-02-01",
		}

		expectTx(t, deps.sqlMock, false)

		_, err := deps.service.Create(ctx, companyID, req)

		assert.Error(t, err)
		assert.NoError(t, deps.sqlMock.ExpectationsWereMet())
	})

	t.Run("repo create error", func(t *testing.T) {
		req := employeesalary.CreateEmployeeSalaryRequest{
			EmployeeID:    employeeID.String(),
			BaseSalary:    9000000,
			EffectiveDate: "2026-02-01",
		}

		expectTx(t, deps.sqlMock, false)

		deps.repo.createFn = func(ctx context.Context, salary *employeesalary.EmployeeSalary) error {
			return errors.New("db error")
		}

		_, err := deps.service.Create(ctx, companyID, req)

		assert.Error(t, err)
		assert.NoError(t, deps.sqlMock.ExpectationsWereMet())
	})
}

func TestEmployeeSalaryService_GetAll(t *testing.T) {
	deps := setupServiceTest(t)
	defer deps.db.Close()

	ctx := context.Background()
	companyID := uuid.New().String()
	employeeID := uuid.New()

	t.Run("success", func(t *testing.T) {
		deps.repo.findAllByCompanyFn = func(ctx context.Context, cid string) ([]employeesalary.EmployeeSalary, error) {
			assert.Equal(t, companyID, cid)
			return []employeesalary.EmployeeSalary{
				{
					ID:            uuid.New(),
					EmployeeID:    employeeID,
					BaseSalary:    10000000,
					EffectiveDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			}, nil
		}

		resp, err := deps.service.GetAll(ctx, companyID)

		assert.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, employeeID.String(), resp[0].EmployeeID)
		assert.Equal(t, 10000000, resp[0].BaseSalary)
	})

	t.Run("repo error", func(t *testing.T) {
		deps.repo.findAllByCompanyFn = func(ctx context.Context, cid string) ([]employeesalary.EmployeeSalary, error) {
			return nil, errors.New("db error")
		}

		resp, err := deps.service.GetAll(ctx, companyID)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestEmployeeSalaryService_Update(t *testing.T) {
	deps := setupServiceTest(t)
	defer deps.db.Close()

	ctx := context.Background()
	companyID := uuid.New().String()
	oldSalaryID := uuid.New()
	employeeID := uuid.New()

	t.Run("success insert new history", func(t *testing.T) {
		req := employeesalary.UpdateEmployeeSalaryRequest{
			EmployeeID:    employeeID.String(),
			BaseSalary:    12000000,
			EffectiveDate: "2026-03-01",
		}

		expectTx(t, deps.sqlMock, true)

		deps.repo.findByIDAndCompanyFn = func(ctx context.Context, cid string, id string) (*employeesalary.EmployeeSalary, error) {
			assert.Equal(t, companyID, cid)
			assert.Equal(t, oldSalaryID.String(), id)
			return &employeesalary.EmployeeSalary{
				ID:            oldSalaryID,
				EmployeeID:    employeeID,
				BaseSalary:    10000000,
				EffectiveDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			}, nil
		}

		deps.repo.createFn = func(ctx context.Context, salary *employeesalary.EmployeeSalary) error {
			assert.Equal(t, employeeID, salary.EmployeeID)
			assert.Equal(t, 12000000, salary.BaseSalary)
			assert.Equal(t, "2026-03-01", salary.EffectiveDate.Format("2006-01-02"))
			assert.NotEqual(t, oldSalaryID, salary.ID)
			return nil
		}

		resp, err := deps.service.Update(ctx, companyID, oldSalaryID.String(), req)

		assert.NoError(t, err)
		assert.Equal(t, employeeID.String(), resp.EmployeeID)
		assert.Equal(t, 12000000, resp.BaseSalary)
		assert.Equal(t, "2026-03-01", resp.EffectiveDate)
		assert.NoError(t, deps.sqlMock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		req := employeesalary.UpdateEmployeeSalaryRequest{
			EmployeeID:    employeeID.String(),
			BaseSalary:    12000000,
			EffectiveDate: "2026-03-01",
		}

		expectTx(t, deps.sqlMock, false)

		deps.repo.findByIDAndCompanyFn = func(ctx context.Context, cid string, id string) (*employeesalary.EmployeeSalary, error) {
			return nil, errors.New("salary not found")
		}

		_, err := deps.service.Update(ctx, companyID, oldSalaryID.String(), req)

		assert.Error(t, err)
		assert.NoError(t, deps.sqlMock.ExpectationsWereMet())
	})
}

func TestEmployeeSalaryService_Delete(t *testing.T) {
	deps := setupServiceTest(t)
	defer deps.db.Close()

	ctx := context.Background()
	companyID := uuid.New().String()
	salaryID := uuid.New().String()

	t.Run("success", func(t *testing.T) {
		expectTx(t, deps.sqlMock, true)

		deps.repo.deleteFn = func(ctx context.Context, cid string, id string) error {
			assert.Equal(t, companyID, cid)
			assert.Equal(t, salaryID, id)
			return nil
		}

		err := deps.service.Delete(ctx, companyID, salaryID)

		assert.NoError(t, err)
		assert.NoError(t, deps.sqlMock.ExpectationsWereMet())
	})

	t.Run("repo error", func(t *testing.T) {
		expectTx(t, deps.sqlMock, false)

		deps.repo.deleteFn = func(ctx context.Context, cid string, id string) error {
			return errors.New("delete failed")
		}

		err := deps.service.Delete(ctx, companyID, salaryID)

		assert.Error(t, err)
		assert.NoError(t, deps.sqlMock.ExpectationsWereMet())
	})
}
