package payroll

import (
	"context"
	"database/sql"
	"go-hris/internal/tenant"
	"time"

	"gorm.io/gorm"
)

//go:generate mockgen -source=payroll_repo.go -destination=mock/payroll_repo_mock.go -package=mock
type Repository interface {
	WithTx(tx *sql.Tx) Repository
	Create(ctx context.Context, payroll *Payroll) error
	FindAllByCompany(ctx context.Context, companyID string, filter PayrollQueryFilter) ([]Payroll, error)
	FindByIDAndCompany(ctx context.Context, companyID string, id string) (*Payroll, error)
	ReplaceComponents(ctx context.Context, companyID string, payrollID string, components []PayrollComponent) error
	Update(ctx context.Context, payroll *Payroll) error
	Delete(ctx context.Context, companyID string, id string) error
	EmployeeBelongsToCompany(ctx context.Context, companyID string, employeeID string) (bool, error)
	HasOverlappingPeriod(ctx context.Context, companyID string, employeeID string, periodStart time.Time, periodEnd time.Time, excludePayrollID *string) (bool, error)
}

type repository struct {
	db *gorm.DB
	tx *sql.Tx
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) WithTx(tx *sql.Tx) Repository {
	return &repository{
		db: r.db,
		tx: tx,
	}
}

func (r *repository) Create(ctx context.Context, payroll *Payroll) error {
	return r.db.WithContext(ctx).Create(payroll).Error
}

func (r *repository) FindAllByCompany(ctx context.Context, companyID string, filter PayrollQueryFilter) ([]Payroll, error) {
	var payrolls []Payroll

	db := r.db.WithContext(ctx).
		Model(&Payroll{}).
		Preload("Employee").
		Joins("LEFT JOIN employees ON employees.id = payrolls.employee_id AND employees.deleted_at IS NULL").
		Where("payrolls.company_id = ?", companyID)

	if filter.Status != nil && *filter.Status != "" {
		db = db.Where("payrolls.status = ?", *filter.Status)
	}
	if filter.DepartmentID != nil && *filter.DepartmentID != "" {
		db = db.Where("employees.department_id = ?", *filter.DepartmentID)
	}
	if filter.PeriodStart != nil && *filter.PeriodStart != "" {
		db = db.Where("payrolls.period_end >= ?", *filter.PeriodStart)
	}
	if filter.PeriodEnd != nil && *filter.PeriodEnd != "" {
		db = db.Where("payrolls.period_start <= ?", *filter.PeriodEnd)
	}

	err := db.Order("payrolls.period_start DESC").
		Find(&payrolls).Error
	return payrolls, err
}

func (r *repository) FindByIDAndCompany(ctx context.Context, companyID string, id string) (*Payroll, error) {
	var payroll Payroll
	err := r.db.WithContext(ctx).
		Scopes(tenant.Scope(companyID)).
		Preload("Employee").
		Preload("Components").
		First(&payroll, "id = ?", id).Error
	return &payroll, err
}

func (r *repository) ReplaceComponents(
	ctx context.Context,
	companyID string,
	payrollID string,
	components []PayrollComponent,
) error {
	db := r.db.WithContext(ctx)
	if err := db.Scopes(tenant.Scope(companyID)).
		Where("payroll_id = ?", payrollID).
		Delete(&PayrollComponent{}).Error; err != nil {
		return err
	}

	if len(components) == 0 {
		return nil
	}

	return db.Create(&components).Error
}

func (r *repository) Update(ctx context.Context, payroll *Payroll) error {
	return r.db.WithContext(ctx).Save(payroll).Error
}

func (r *repository) Delete(ctx context.Context, companyID string, id string) error {
	return r.db.WithContext(ctx).
		Scopes(tenant.Scope(companyID)).
		Delete(&Payroll{}, "id = ?", id).Error
}

func (r *repository) EmployeeBelongsToCompany(ctx context.Context, companyID string, employeeID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("employees").
		Where("id = ?", employeeID).
		Scopes(tenant.Scope(companyID)).
		Where("deleted_at IS NULL").
		Count(&count).Error
	return count > 0, err
}

func (r *repository) HasOverlappingPeriod(
	ctx context.Context,
	companyID string,
	employeeID string,
	periodStart time.Time,
	periodEnd time.Time,
	excludePayrollID *string,
) (bool, error) {
	db := r.db.WithContext(ctx).
		Model(&Payroll{}).
		Scopes(tenant.Scope(companyID)).
		Where("employee_id = ?", employeeID).
		Where("NOT (period_end < ? OR period_start > ?)", periodStart, periodEnd)

	if excludePayrollID != nil && *excludePayrollID != "" {
		db = db.Where("id <> ?", *excludePayrollID)
	}

	var count int64
	err := db.Count(&count).Error
	return count > 0, err
}
