package leave

import (
	"context"
	"database/sql"
	"time"

	"gorm.io/gorm"
)

//go:generate mockgen -source=leave_repo.go -destination=mock/leave_repo_mock.go -package=mock
type Repository interface {
	WithTx(tx *sql.Tx) Repository
	Create(ctx context.Context, l *Leave) error
	FindAllByCompany(ctx context.Context, companyID string) ([]Leave, error)
	FindByIDAndCompany(ctx context.Context, companyID, id string) (*Leave, error)
	Update(ctx context.Context, l *Leave) error
	Delete(ctx context.Context, companyID, id string) error
	EmployeeBelongsToCompany(ctx context.Context, companyID, employeeID string) (bool, error)
	HasOverlappingPeriod(ctx context.Context, companyID, employeeID string, startDate, endDate time.Time, excludeID *string) (bool, error)
}

type repository struct {
	db *gorm.DB
	tx *sql.Tx
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) WithTx(tx *sql.Tx) Repository {
	return &repository{db: r.db, tx: tx}
}

func (r *repository) Create(ctx context.Context, l *Leave) error {
	return r.db.WithContext(ctx).Create(l).Error
}

func (r *repository) FindAllByCompany(ctx context.Context, companyID string) ([]Leave, error) {
	var leaves []Leave
	err := r.db.WithContext(ctx).
		Where("company_id = ?", companyID).
		Order("start_date DESC").
		Find(&leaves).Error
	return leaves, err
}

func (r *repository) FindByIDAndCompany(ctx context.Context, companyID, id string) (*Leave, error) {
	var l Leave
	err := r.db.WithContext(ctx).
		Where("company_id = ?", companyID).
		First(&l, "id = ?", id).Error
	return &l, err
}

func (r *repository) Update(ctx context.Context, l *Leave) error {
	return r.db.WithContext(ctx).Save(l).Error
}

func (r *repository) Delete(ctx context.Context, companyID, id string) error {
	return r.db.WithContext(ctx).
		Where("company_id = ?", companyID).
		Delete(&Leave{}, "id = ?", id).Error
}

func (r *repository) EmployeeBelongsToCompany(ctx context.Context, companyID, employeeID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("employees").
		Where("id = ?", employeeID).
		Where("company_id = ?", companyID).
		Where("deleted_at IS NULL").
		Count(&count).Error
	return count > 0, err
}

func (r *repository) HasOverlappingPeriod(ctx context.Context, companyID, employeeID string, startDate, endDate time.Time, excludeID *string) (bool, error) {
	db := r.db.WithContext(ctx).
		Model(&Leave{}).
		Where("company_id = ?", companyID).
		Where("employee_id = ?", employeeID).
		Where("status <> ?", "CANCELLED").
		Where("NOT (end_date < ? OR start_date > ?)", startDate, endDate)

	if excludeID != nil && *excludeID != "" {
		db = db.Where("id <> ?", *excludeID)
	}

	var count int64
	err := db.Count(&count).Error
	return count > 0, err
}
