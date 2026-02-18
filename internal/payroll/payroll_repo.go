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
	FindAllByCompany(ctx context.Context, companyID string) ([]Payroll, error)
	FindByIDAndCompany(ctx context.Context, companyID string, id string) (*Payroll, error)
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

func (r *repository) FindAllByCompany(ctx context.Context, companyID string) ([]Payroll, error) {
	var payrolls []Payroll
	err := r.db.WithContext(ctx).
		Scopes(tenant.Scope(companyID)).
		Order("period_start DESC").
		Find(&payrolls).Error
	return payrolls, err
}

func (r *repository) FindByIDAndCompany(ctx context.Context, companyID string, id string) (*Payroll, error) {
	var payroll Payroll
	err := r.db.WithContext(ctx).
		Scopes(tenant.Scope(companyID)).
		First(&payroll, "id = ?", id).Error
	return &payroll, err
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
