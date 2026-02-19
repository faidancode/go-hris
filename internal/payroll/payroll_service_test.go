package payroll_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go-hris/internal/events"
	"go-hris/internal/messaging/kafka"
	"go-hris/internal/payroll"
	payrollerrors "go-hris/internal/payroll/errors"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type fakePayrollRepository struct {
	withTxFn                 func(tx *sql.Tx) payroll.Repository
	createFn                 func(ctx context.Context, p *payroll.Payroll) error
	findAllByCompanyFn       func(ctx context.Context, companyID string, filter payroll.PayrollQueryFilter) ([]payroll.Payroll, error)
	findByIDAndCompanyFn     func(ctx context.Context, companyID string, id string) (*payroll.Payroll, error)
	replaceComponentsFn      func(ctx context.Context, companyID string, payrollID string, components []payroll.PayrollComponent) error
	updateFn                 func(ctx context.Context, p *payroll.Payroll) error
	deleteFn                 func(ctx context.Context, companyID string, id string) error
	employeeBelongsToCompany func(ctx context.Context, companyID string, employeeID string) (bool, error)
	hasOverlappingPeriodFn   func(ctx context.Context, companyID string, employeeID string, periodStart time.Time, periodEnd time.Time, excludePayrollID *string) (bool, error)
}

type fakeOutboxRepository struct {
	withTxFn func(tx *sql.Tx) kafka.OutboxRepository
	createFn func(ctx context.Context, event kafka.OutboxEvent) error
}

func (f *fakeOutboxRepository) WithTx(tx *sql.Tx) kafka.OutboxRepository {
	if f.withTxFn != nil {
		return f.withTxFn(tx)
	}
	return f
}

func (f *fakeOutboxRepository) Create(ctx context.Context, event kafka.OutboxEvent) error {
	if f.createFn != nil {
		return f.createFn(ctx, event)
	}
	return nil
}

func (f *fakeOutboxRepository) ListPending(ctx context.Context, limit int) ([]kafka.OutboxEvent, error) {
	return nil, nil
}

func (f *fakeOutboxRepository) MarkSent(ctx context.Context, id string) error {
	return nil
}

func (f *fakeOutboxRepository) MarkFailed(ctx context.Context, id string, reason string) error {
	return nil
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

func (f *fakePayrollRepository) FindAllByCompany(ctx context.Context, companyID string, filter payroll.PayrollQueryFilter) ([]payroll.Payroll, error) {
	if f.findAllByCompanyFn != nil {
		return f.findAllByCompanyFn(ctx, companyID, filter)
	}
	return nil, nil
}

func (f *fakePayrollRepository) FindByIDAndCompany(ctx context.Context, companyID string, id string) (*payroll.Payroll, error) {
	if f.findByIDAndCompanyFn != nil {
		return f.findByIDAndCompanyFn(ctx, companyID, id)
	}
	return nil, nil
}

func (f *fakePayrollRepository) ReplaceComponents(ctx context.Context, companyID string, payrollID string, components []payroll.PayrollComponent) error {
	if f.replaceComponentsFn != nil {
		return f.replaceComponentsFn(ctx, companyID, payrollID, components)
	}
	return nil
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

	return &payrollServiceDeps{db: db, sqlMock: sqlMock, service: svc, repo: repo}
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

	deps := setupPayrollServiceTest(t)
	defer deps.db.Close()

	expectTx(t, deps.sqlMock, true)
	req := payroll.CreatePayrollRequest{
		EmployeeID:    employeeID,
		PeriodStart:   "2026-02-01",
		PeriodEnd:     "2026-02-28",
		BaseSalary:    10000000,
		Allowance:     250000,
		OvertimeHours: 3,
		OvertimeRate:  25000,
		Deduction:     100000,
		DeductionItems: []payroll.PayrollComponentInput{
			{ComponentName: "Late Penalty", Quantity: 2, UnitAmount: 50000},
		},
	}

	createdPayrollID := uuid.New()
	deps.repo.employeeBelongsToCompany = func(ctx context.Context, cid, eid string) (bool, error) {
		return true, nil
	}
	deps.repo.createFn = func(ctx context.Context, p *payroll.Payroll) error {
		p.ID = createdPayrollID
		assert.Equal(t, int64(3*25000), p.OvertimeAmount)
		assert.Equal(t, int64(200000), p.Deduction)
		assert.Equal(t, int64(10125000), p.NetSalary)
		return nil
	}
	deps.repo.findByIDAndCompanyFn = func(ctx context.Context, companyID string, id string) (*payroll.Payroll, error) {
		return &payroll.Payroll{ID: createdPayrollID, CompanyID: uuid.MustParse(companyID), EmployeeID: uuid.MustParse(employeeID), PeriodStart: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC), PeriodEnd: time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC), BaseSalary: 10000000, Allowance: 250000, OvertimeHours: 3, OvertimeRate: 25000, OvertimeAmount: 75000, Deduction: 200000, NetSalary: 10125000, Status: payroll.StatusDraft, CreatedBy: uuid.MustParse(actorID)}, nil
	}

	resp, err := deps.service.Create(ctx, companyID, actorID, req)

	assert.NoError(t, err)
	assert.Equal(t, int64(75000), resp.TotalOvertime)
	assert.Equal(t, int64(200000), resp.TotalDeduction)
	assert.Equal(t, payroll.StatusDraft, resp.Status)
	assert.NoError(t, deps.sqlMock.ExpectationsWereMet())
}

func TestPayrollService_Regenerate_OnlyDraft(t *testing.T) {
	ctx := context.Background()
	companyID := uuid.New().String()
	actorID := uuid.New().String()
	payrollID := uuid.New().String()

	deps := setupPayrollServiceTest(t)
	defer deps.db.Close()

	expectTx(t, deps.sqlMock, false)
	deps.repo.findByIDAndCompanyFn = func(ctx context.Context, companyID string, id string) (*payroll.Payroll, error) {
		return &payroll.Payroll{ID: uuid.MustParse(id), CompanyID: uuid.MustParse(companyID), Status: payroll.StatusApproved}, nil
	}

	_, err := deps.service.Regenerate(ctx, companyID, actorID, payrollID, payroll.RegeneratePayrollRequest{BaseSalary: 100})

	assert.ErrorIs(t, err, payrollerrors.ErrRegenerateOnlyDraft)
	assert.NoError(t, deps.sqlMock.ExpectationsWereMet())
}

func TestPayrollService_ApproveAndMarkPaid(t *testing.T) {
	ctx := context.Background()
	companyID := uuid.New().String()
	actorID := uuid.New().String()
	payrollID := uuid.New().String()

	t.Run("approve success", func(t *testing.T) {
		deps := setupPayrollServiceTest(t)
		defer deps.db.Close()

		expectTx(t, deps.sqlMock, true)
		deps.repo.findByIDAndCompanyFn = func(ctx context.Context, companyID string, id string) (*payroll.Payroll, error) {
			return &payroll.Payroll{ID: uuid.MustParse(id), CompanyID: uuid.MustParse(companyID), Status: payroll.StatusDraft}, nil
		}

		resp, err := deps.service.Approve(ctx, companyID, actorID, payrollID)

		assert.NoError(t, err)
		assert.Equal(t, payroll.StatusApproved, resp.Status)
		assert.NotNil(t, resp.ApprovedBy)
		assert.NoError(t, deps.sqlMock.ExpectationsWereMet())
	})

	t.Run("mark paid success", func(t *testing.T) {
		deps := setupPayrollServiceTest(t)
		defer deps.db.Close()

		expectTx(t, deps.sqlMock, true)
		deps.repo.findByIDAndCompanyFn = func(ctx context.Context, companyID string, id string) (*payroll.Payroll, error) {
			return &payroll.Payroll{ID: uuid.MustParse(id), CompanyID: uuid.MustParse(companyID), Status: payroll.StatusApproved}, nil
		}

		resp, err := deps.service.MarkAsPaid(ctx, companyID, actorID, payrollID)

		assert.NoError(t, err)
		assert.Equal(t, payroll.StatusPaid, resp.Status)
		assert.NotNil(t, resp.PaidAt)
		assert.NoError(t, deps.sqlMock.ExpectationsWereMet())
	})
}

func TestPayrollService_Delete_OnlyDraft(t *testing.T) {
	ctx := context.Background()
	companyID := uuid.New().String()
	payrollID := uuid.New().String()

	t.Run("negative non-draft", func(t *testing.T) {
		deps := setupPayrollServiceTest(t)
		defer deps.db.Close()

		expectTx(t, deps.sqlMock, false)
		deps.repo.findByIDAndCompanyFn = func(ctx context.Context, companyID string, id string) (*payroll.Payroll, error) {
			return &payroll.Payroll{ID: uuid.MustParse(id), CompanyID: uuid.MustParse(companyID), Status: payroll.StatusPaid}, nil
		}

		err := deps.service.Delete(ctx, companyID, payrollID)

		assert.ErrorIs(t, err, payrollerrors.ErrDeleteOnlyDraft)
		assert.NoError(t, deps.sqlMock.ExpectationsWereMet())
	})

	t.Run("success", func(t *testing.T) {
		deps := setupPayrollServiceTest(t)
		defer deps.db.Close()

		expectTx(t, deps.sqlMock, true)
		deps.repo.findByIDAndCompanyFn = func(ctx context.Context, companyID string, id string) (*payroll.Payroll, error) {
			return &payroll.Payroll{ID: uuid.MustParse(id), CompanyID: uuid.MustParse(companyID), Status: payroll.StatusDraft}, nil
		}
		deps.repo.deleteFn = func(ctx context.Context, companyID string, id string) error {
			return nil
		}

		err := deps.service.Delete(ctx, companyID, payrollID)

		assert.NoError(t, err)
		assert.NoError(t, deps.sqlMock.ExpectationsWereMet())
	})
}

func TestPayrollService_GetAll_RepoError(t *testing.T) {
	ctx := context.Background()
	companyID := uuid.New().String()

	deps := setupPayrollServiceTest(t)
	defer deps.db.Close()
	deps.repo.findAllByCompanyFn = func(ctx context.Context, companyID string, filter payroll.PayrollQueryFilter) ([]payroll.Payroll, error) {
		assert.NotNil(t, filter.Status)
		assert.Equal(t, payroll.StatusDraft, *filter.Status)
		assert.NotNil(t, filter.PeriodStart)
		assert.Equal(t, "2026-02-01", *filter.PeriodStart)
		return nil, errors.New("db error")
	}

	resp, err := deps.service.GetAll(ctx, companyID, payroll.GetPayrollsFilterRequest{
		Period: "2026-02",
		Status: "draft",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestPayrollService_GetBreakdown(t *testing.T) {
	ctx := context.Background()
	companyID := uuid.New().String()
	payrollID := uuid.New().String()

	deps := setupPayrollServiceTest(t)
	defer deps.db.Close()

	deps.repo.findByIDAndCompanyFn = func(ctx context.Context, companyID string, id string) (*payroll.Payroll, error) {
		note := "3 jam x 25.000"
		return &payroll.Payroll{
			ID:             uuid.MustParse(id),
			CompanyID:      uuid.MustParse(companyID),
			EmployeeID:     uuid.New(),
			PeriodStart:    time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			PeriodEnd:      time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC),
			BaseSalary:     10000000,
			Allowance:      500000,
			OvertimeHours:  3,
			OvertimeRate:   25000,
			OvertimeAmount: 75000,
			Deduction:      100000,
			NetSalary:      10475000,
			Status:         payroll.StatusDraft,
			Components: []payroll.PayrollComponent{
				{
					ComponentType: payroll.ComponentTypeAllowance,
					ComponentName: "Transport",
					Quantity:      1,
					UnitAmount:    500000,
					TotalAmount:   500000,
				},
				{
					ComponentType: payroll.ComponentTypeDeduction,
					ComponentName: "Late Penalty",
					Quantity:      2,
					UnitAmount:    50000,
					TotalAmount:   100000,
					Notes:         &note,
				},
			},
		}, nil
	}

	resp, err := deps.service.GetBreakdown(ctx, companyID, payrollID)

	assert.NoError(t, err)
	assert.Equal(t, payrollID, resp.PayrollID)
	assert.Equal(t, int64(500000), resp.AllowanceTotal)
	assert.Equal(t, int64(75000), resp.Overtime.Amount)
	assert.Equal(t, int64(100000), resp.DeductionTotal)
	assert.Equal(t, int64(10475000), resp.NetSalary)
	assert.Len(t, resp.Allowances, 1)
	assert.Len(t, resp.Deductions, 1)
}

func TestPayrollService_Approve_QueuesPayslipEvent(t *testing.T) {
	ctx := context.Background()
	companyID := uuid.New().String()
	actorID := uuid.New().String()
	payrollID := uuid.New().String()

	db, sqlMock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &fakePayrollRepository{
		findByIDAndCompanyFn: func(ctx context.Context, companyID string, id string) (*payroll.Payroll, error) {
			return &payroll.Payroll{ID: uuid.MustParse(id), CompanyID: uuid.MustParse(companyID), Status: payroll.StatusDraft}, nil
		},
	}
	outbox := &fakeOutboxRepository{
		createFn: func(ctx context.Context, event kafka.OutboxEvent) error {
			assert.Equal(t, events.PayrollPayslipRequestedTopic, event.Topic)
			var payload events.PayrollPayslipRequestedEvent
			err := json.Unmarshal(event.Payload, &payload)
			assert.NoError(t, err)
			assert.Equal(t, companyID, payload.CompanyID)
			assert.Equal(t, payrollID, payload.PayrollID)
			return nil
		},
	}
	svc := payroll.NewServiceWithOutbox(db, repo, outbox)

	expectTx(t, sqlMock, true)
	_, err = svc.Approve(ctx, companyID, actorID, payrollID)
	assert.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestPayrollService_GeneratePayslip(t *testing.T) {
	ctx := context.Background()
	companyID := uuid.New().String()
	payrollID := uuid.New().String()

	deps := setupPayrollServiceTest(t)
	defer deps.db.Close()

	tmpDir := t.TempDir()
	_ = os.Setenv("PAYSLIP_STORAGE_DIR", tmpDir)
	_ = os.Setenv("PAYSLIP_PUBLIC_BASE_URL", "/files/payslips")
	t.Cleanup(func() {
		_ = os.Unsetenv("PAYSLIP_STORAGE_DIR")
		_ = os.Unsetenv("PAYSLIP_PUBLIC_BASE_URL")
	})

	expectTx(t, deps.sqlMock, true)
	deps.repo.findByIDAndCompanyFn = func(ctx context.Context, companyID string, id string) (*payroll.Payroll, error) {
		return &payroll.Payroll{
			ID:             uuid.MustParse(id),
			CompanyID:      uuid.MustParse(companyID),
			EmployeeID:     uuid.New(),
			PeriodStart:    time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			PeriodEnd:      time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC),
			BaseSalary:     10000000,
			Allowance:      500000,
			OvertimeHours:  2,
			OvertimeRate:   25000,
			OvertimeAmount: 50000,
			Deduction:      100000,
			NetSalary:      10450000,
			Status:         payroll.StatusApproved,
		}, nil
	}

	resp, err := deps.service.GeneratePayslip(ctx, companyID, payrollID)
	assert.NoError(t, err)
	if assert.NotNil(t, resp.PayslipURL) {
		assert.Contains(t, *resp.PayslipURL, "/files/payslips/payslip_")
	}
	assert.NotNil(t, resp.PayslipGeneratedAt)

	filename := "payslip_" + payrollID + ".pdf"
	_, statErr := os.Stat(filepath.Join(tmpDir, filename))
	assert.NoError(t, statErr)
	assert.NoError(t, deps.sqlMock.ExpectationsWereMet())
}
