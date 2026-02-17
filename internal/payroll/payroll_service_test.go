package payroll_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"go-hris/internal/payroll"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type fakePayrollRepository struct {
	withTxFn                 func(tx *sql.Tx) payroll.Repository
	createFn                 func(ctx context.Context, p *payroll.Payroll) error
	findAllByCompanyFn       func(ctx context.Context, companyID string) ([]payroll.Payroll, error)
	findByIDAndCompanyFn     func(ctx context.Context, companyID string, id string) (*payroll.Payroll, error)
	updateFn                 func(ctx context.Context, p *payroll.Payroll) error
	deleteFn                 func(ctx context.Context, companyID string, id string) error
	employeeBelongsToCompany func(ctx context.Context, companyID string, employeeID string) (bool, error)
	hasOverlappingPeriodFn   func(ctx context.Context, companyID string, employeeID string, periodStart time.Time, periodEnd time.Time, excludePayrollID *string) (bool, error)
}

func (f *fakePayrollRepository) WithTx(tx *sql.Tx) payroll.Repository {
	if f.withTxFn != nil {
		return f.withTxFn(tx)
	}
	return f
}

func (f *fakePayrollRepository) Create(ctx context.Context, p *payroll.Payroll) error {
	if f.createFn != nil {
		return f.createFn(ctx, p)
	}
	return nil
}

func (f *fakePayrollRepository) FindAllByCompany(ctx context.Context, companyID string) ([]payroll.Payroll, error) {
	if f.findAllByCompanyFn != nil {
		return f.findAllByCompanyFn(ctx, companyID)
	}
	return nil, nil
}

func (f *fakePayrollRepository) FindByIDAndCompany(ctx context.Context, companyID string, id string) (*payroll.Payroll, error) {
	if f.findByIDAndCompanyFn != nil {
		return f.findByIDAndCompanyFn(ctx, companyID, id)
	}
	return nil, nil
}

func (f *fakePayrollRepository) Update(ctx context.Context, p *payroll.Payroll) error {
	if f.updateFn != nil {
		return f.updateFn(ctx, p)
	}
	return nil
}

func (f *fakePayrollRepository) Delete(ctx context.Context, companyID string, id string) error {
	if f.deleteFn != nil {
		return f.deleteFn(ctx, companyID, id)
	}
	return nil
}

func (f *fakePayrollRepository) EmployeeBelongsToCompany(ctx context.Context, companyID string, employeeID string) (bool, error) {
	if f.employeeBelongsToCompany != nil {
		return f.employeeBelongsToCompany(ctx, companyID, employeeID)
	}
	return true, nil
}

func (f *fakePayrollRepository) HasOverlappingPeriod(ctx context.Context, companyID string, employeeID string, periodStart time.Time, periodEnd time.Time, excludePayrollID *string) (bool, error) {
	if f.hasOverlappingPeriodFn != nil {
		return f.hasOverlappingPeriodFn(ctx, companyID, employeeID, periodStart, periodEnd, excludePayrollID)
	}
	return false, nil
}

type payrollServiceDeps struct {
	db      *sql.DB
	sqlMock sqlmock.Sqlmock
	service payroll.Service
	repo    *fakePayrollRepository
}

func setupPayrollServiceTest(t *testing.T) *payrollServiceDeps {
	t.Helper()

	db, sqlMock, err := sqlmock.New()
	assert.NoError(t, err)

	repo := &fakePayrollRepository{}
	svc := payroll.NewService(db, repo)

	return &payrollServiceDeps{
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

func TestPayrollService_Create(t *testing.T) {
	ctx := context.Background()
	companyID := uuid.New().String()
	actorID := uuid.New().String()
	employeeID := uuid.New().String()

	t.Run("success", func(t *testing.T) {
		deps := setupPayrollServiceTest(t)
		defer deps.db.Close()

		expectTx(t, deps.sqlMock, true)
		req := payroll.CreatePayrollRequest{
			EmployeeID:  employeeID,
			PeriodStart: "2026-02-01",
			PeriodEnd:   "2026-02-28",
			BaseSalary:  10000000,
			Allowance:   250000,
			Deduction:   100000,
		}

		deps.repo.employeeBelongsToCompany = func(ctx context.Context, cid, eid string) (bool, error) {
			assert.Equal(t, companyID, cid)
			assert.Equal(t, employeeID, eid)
			return true, nil
		}
		deps.repo.hasOverlappingPeriodFn = func(ctx context.Context, cid, eid string, start, end time.Time, exclude *string) (bool, error) {
			assert.Equal(t, companyID, cid)
			assert.Equal(t, employeeID, eid)
			assert.Nil(t, exclude)
			return false, nil
		}
		deps.repo.createFn = func(ctx context.Context, p *payroll.Payroll) error {
			assert.Equal(t, uuid.MustParse(companyID), p.CompanyID)
			assert.Equal(t, uuid.MustParse(employeeID), p.EmployeeID)
			assert.Equal(t, uuid.MustParse(actorID), p.CreatedBy)
			assert.Equal(t, "2026-02-01", p.PeriodStart.Format("2006-01-02"))
			assert.Equal(t, "2026-02-28", p.PeriodEnd.Format("2006-01-02"))
			assert.Equal(t, payroll.StatusDraft, p.Status)
			assert.Equal(t, int64(10150000), p.NetSalary)
			return nil
		}

		resp, err := deps.service.Create(ctx, companyID, actorID, req)

		assert.NoError(t, err)
		assert.Equal(t, req.EmployeeID, resp.EmployeeID)
		assert.Equal(t, payroll.StatusDraft, resp.Status)
		assert.Equal(t, int64(10150000), resp.NetSalary)
		assert.Equal(t, actorID, resp.CreatedBy)
		assert.NoError(t, deps.sqlMock.ExpectationsWereMet())
	})

	t.Run("negative invalid company id", func(t *testing.T) {
		deps := setupPayrollServiceTest(t)
		defer deps.db.Close()

		expectTx(t, deps.sqlMock, false)
		req := payroll.CreatePayrollRequest{
			EmployeeID:  employeeID,
			PeriodStart: "2026-02-01",
			PeriodEnd:   "2026-02-28",
			BaseSalary:  10000000,
		}

		_, err := deps.service.Create(ctx, "invalid-company-id", actorID, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid company id")
		assert.NoError(t, deps.sqlMock.ExpectationsWereMet())
	})
}

func TestPayrollService_GetAll(t *testing.T) {
	ctx := context.Background()
	companyID := uuid.New().String()

	t.Run("success", func(t *testing.T) {
		deps := setupPayrollServiceTest(t)
		defer deps.db.Close()

		employeeID := uuid.New()
		deps.repo.findAllByCompanyFn = func(ctx context.Context, cid string) ([]payroll.Payroll, error) {
			assert.Equal(t, companyID, cid)
			return []payroll.Payroll{
				{
					ID:          uuid.New(),
					CompanyID:   uuid.MustParse(companyID),
					EmployeeID:  employeeID,
					PeriodStart: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
					PeriodEnd:   time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC),
					BaseSalary:  10000000,
					Allowance:   200000,
					Deduction:   50000,
					NetSalary:   10150000,
					Status:      payroll.StatusDraft,
					CreatedBy:   uuid.New(),
				},
			}, nil
		}

		resp, err := deps.service.GetAll(ctx, companyID)

		assert.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, employeeID.String(), resp[0].EmployeeID)
	})

	t.Run("negative repo error", func(t *testing.T) {
		deps := setupPayrollServiceTest(t)
		defer deps.db.Close()

		deps.repo.findAllByCompanyFn = func(ctx context.Context, cid string) ([]payroll.Payroll, error) {
			return nil, errors.New("db error")
		}

		resp, err := deps.service.GetAll(ctx, companyID)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestPayrollService_GetByID(t *testing.T) {
	ctx := context.Background()
	companyID := uuid.New().String()
	id := uuid.New().String()

	t.Run("success", func(t *testing.T) {
		deps := setupPayrollServiceTest(t)
		defer deps.db.Close()

		employeeID := uuid.New()
		deps.repo.findByIDAndCompanyFn = func(ctx context.Context, cid, targetID string) (*payroll.Payroll, error) {
			assert.Equal(t, companyID, cid)
			assert.Equal(t, id, targetID)
			return &payroll.Payroll{
				ID:          uuid.MustParse(id),
				CompanyID:   uuid.MustParse(companyID),
				EmployeeID:  employeeID,
				PeriodStart: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
				PeriodEnd:   time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC),
				BaseSalary:  10000000,
				Allowance:   200000,
				Deduction:   50000,
				NetSalary:   10150000,
				Status:      payroll.StatusDraft,
				CreatedBy:   uuid.New(),
			}, nil
		}

		resp, err := deps.service.GetByID(ctx, companyID, id)

		assert.NoError(t, err)
		assert.Equal(t, id, resp.ID)
	})

	t.Run("negative repo error", func(t *testing.T) {
		deps := setupPayrollServiceTest(t)
		defer deps.db.Close()

		deps.repo.findByIDAndCompanyFn = func(ctx context.Context, cid, targetID string) (*payroll.Payroll, error) {
			return nil, errors.New("not found")
		}

		resp, err := deps.service.GetByID(ctx, companyID, id)

		assert.Error(t, err)
		assert.Empty(t, resp.ID)
	})
}

func TestPayrollService_Update(t *testing.T) {
	ctx := context.Background()
	companyID := uuid.New().String()
	actorID := uuid.New().String()
	id := uuid.New().String()
	employeeID := uuid.New().String()

	t.Run("success", func(t *testing.T) {
		deps := setupPayrollServiceTest(t)
		defer deps.db.Close()

		expectTx(t, deps.sqlMock, true)
		approvedBy := uuid.New().String()
		req := payroll.UpdatePayrollRequest{
			EmployeeID:  employeeID,
			PeriodStart: "2026-03-01",
			PeriodEnd:   "2026-03-31",
			BaseSalary:  12000000,
			Allowance:   300000,
			Deduction:   100000,
			Status:      payroll.StatusProcessed,
			ApprovedBy:  &approvedBy,
		}

		deps.repo.findByIDAndCompanyFn = func(ctx context.Context, cid, targetID string) (*payroll.Payroll, error) {
			return &payroll.Payroll{
				ID:          uuid.MustParse(targetID),
				CompanyID:   uuid.MustParse(cid),
				EmployeeID:  uuid.New(),
				PeriodStart: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
				PeriodEnd:   time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC),
				Status:      payroll.StatusDraft,
				CreatedBy:   uuid.MustParse(actorID),
			}, nil
		}
		deps.repo.employeeBelongsToCompany = func(ctx context.Context, cid, eid string) (bool, error) {
			return true, nil
		}
		deps.repo.hasOverlappingPeriodFn = func(ctx context.Context, cid, eid string, start, end time.Time, exclude *string) (bool, error) {
			assert.NotNil(t, exclude)
			assert.Equal(t, id, *exclude)
			return false, nil
		}
		deps.repo.updateFn = func(ctx context.Context, p *payroll.Payroll) error {
			assert.Equal(t, int64(12200000), p.NetSalary)
			assert.Equal(t, payroll.StatusProcessed, p.Status)
			assert.NotNil(t, p.ApprovedBy)
			assert.Equal(t, approvedBy, p.ApprovedBy.String())
			return nil
		}

		resp, err := deps.service.Update(ctx, companyID, actorID, id, req)

		assert.NoError(t, err)
		assert.Equal(t, payroll.StatusProcessed, resp.Status)
		assert.Equal(t, int64(12200000), resp.NetSalary)
		assert.NotNil(t, resp.ApprovedBy)
		assert.Equal(t, approvedBy, *resp.ApprovedBy)
		assert.NoError(t, deps.sqlMock.ExpectationsWereMet())
	})

	t.Run("negative invalid approved_by", func(t *testing.T) {
		deps := setupPayrollServiceTest(t)
		defer deps.db.Close()

		expectTx(t, deps.sqlMock, false)
		approvedBy := "invalid-approved-by"
		req := payroll.UpdatePayrollRequest{
			EmployeeID:  employeeID,
			PeriodStart: "2026-03-01",
			PeriodEnd:   "2026-03-31",
			BaseSalary:  12000000,
			Allowance:   300000,
			Deduction:   100000,
			Status:      payroll.StatusProcessed,
			ApprovedBy:  &approvedBy,
		}

		deps.repo.findByIDAndCompanyFn = func(ctx context.Context, cid, targetID string) (*payroll.Payroll, error) {
			return &payroll.Payroll{
				ID:        uuid.MustParse(targetID),
				CompanyID: uuid.MustParse(cid),
				CreatedBy: uuid.MustParse(actorID),
			}, nil
		}
		deps.repo.employeeBelongsToCompany = func(ctx context.Context, cid, eid string) (bool, error) {
			return true, nil
		}
		deps.repo.hasOverlappingPeriodFn = func(ctx context.Context, cid, eid string, start, end time.Time, exclude *string) (bool, error) {
			return false, nil
		}

		_, err := deps.service.Update(ctx, companyID, actorID, id, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid approved_by")
		assert.NoError(t, deps.sqlMock.ExpectationsWereMet())
	})

	t.Run("negative paid status without paid_at", func(t *testing.T) {
		deps := setupPayrollServiceTest(t)
		defer deps.db.Close()

		expectTx(t, deps.sqlMock, false)
		req := payroll.UpdatePayrollRequest{
			EmployeeID:  employeeID,
			PeriodStart: "2026-03-01",
			PeriodEnd:   "2026-03-31",
			BaseSalary:  12000000,
			Allowance:   300000,
			Deduction:   100000,
			Status:      payroll.StatusPaid,
		}

		deps.repo.findByIDAndCompanyFn = func(ctx context.Context, cid, targetID string) (*payroll.Payroll, error) {
			return &payroll.Payroll{
				ID:        uuid.MustParse(targetID),
				CompanyID: uuid.MustParse(cid),
				CreatedBy: uuid.MustParse(actorID),
			}, nil
		}
		deps.repo.employeeBelongsToCompany = func(ctx context.Context, cid, eid string) (bool, error) {
			return true, nil
		}
		deps.repo.hasOverlappingPeriodFn = func(ctx context.Context, cid, eid string, start, end time.Time, exclude *string) (bool, error) {
			return false, nil
		}

		_, err := deps.service.Update(ctx, companyID, actorID, id, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "paid_at is required")
		assert.NoError(t, deps.sqlMock.ExpectationsWereMet())
	})
}

func TestPayrollService_Delete(t *testing.T) {
	ctx := context.Background()
	companyID := uuid.New().String()
	id := uuid.New().String()

	t.Run("success", func(t *testing.T) {
		deps := setupPayrollServiceTest(t)
		defer deps.db.Close()

		expectTx(t, deps.sqlMock, true)
		deps.repo.deleteFn = func(ctx context.Context, cid, targetID string) error {
			assert.Equal(t, companyID, cid)
			assert.Equal(t, id, targetID)
			return nil
		}

		err := deps.service.Delete(ctx, companyID, id)

		assert.NoError(t, err)
		assert.NoError(t, deps.sqlMock.ExpectationsWereMet())
	})

	t.Run("negative repo error", func(t *testing.T) {
		deps := setupPayrollServiceTest(t)
		defer deps.db.Close()

		expectTx(t, deps.sqlMock, false)
		deps.repo.deleteFn = func(ctx context.Context, cid, targetID string) error {
			return errors.New("delete failed")
		}

		err := deps.service.Delete(ctx, companyID, id)

		assert.Error(t, err)
		assert.NoError(t, deps.sqlMock.ExpectationsWereMet())
	})
}
