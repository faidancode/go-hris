package employee

import (
	"context"
	"database/sql"
	"go-hris/internal/tenant"

	"gorm.io/gorm"
)

//go:generate mockgen -source=employee_repo.go -destination=mock/employee_repo_mock.go -package=mock
type Repository interface {
	WithTx(tx *sql.Tx) Repository
	Create(ctx context.Context, dept *Employee) error
	FindAllByCompany(ctx context.Context, companyID string) ([]Employee, error)
	FindByIDAndCompany(ctx context.Context, companyID string, id string) (*Employee, error)
	GetDepartmentIDByPosition(ctx context.Context, companyID, positionID string) (string, error)
	Update(ctx context.Context, dept *Employee) error
	Delete(ctx context.Context, companyID string, id string) error
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

func (r *repository) Create(ctx context.Context, dept *Employee) error {
	return r.db.WithContext(ctx).Create(dept).Error
}

func (r *repository) FindAllByCompany(ctx context.Context, companyID string) ([]Employee, error) {
	var depts []Employee
	err := r.db.WithContext(ctx).
		Scopes(tenant.Scope(companyID)).
		Find(&depts).Error
	return depts, err
}

func (r *repository) FindByIDAndCompany(ctx context.Context, companyID string, id string) (*Employee, error) {
	var dept Employee
	err := r.db.WithContext(ctx).
		Scopes(tenant.Scope(companyID)).
		First(&dept, "id = ?", id).Error
	return &dept, err
}

func (r *repository) GetDepartmentIDByPosition(ctx context.Context, companyID, positionID string) (string, error) {
	var departmentID string
	err := r.db.WithContext(ctx).
		Table("positions").
		Select("department_id").
		Where("id = ?", positionID).
		Where("company_id = ?", companyID).
		Where("deleted_at IS NULL").
		Scan(&departmentID).Error
	return departmentID, err
}

func (r *repository) Update(ctx context.Context, dept *Employee) error {
	return r.db.WithContext(ctx).Save(dept).Error
}

func (r *repository) Delete(ctx context.Context, companyID string, id string) error {
	return r.db.WithContext(ctx).
		Scopes(tenant.Scope(companyID)).
		Delete(&Employee{}, "id = ?", id).Error
}
