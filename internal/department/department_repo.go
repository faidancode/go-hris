package department

import (
	"context"
	"database/sql"
	"go-hris/internal/tenant"

	"gorm.io/gorm"
)

//go:generate mockgen -source=department_repo.go -destination=mock/department_repo_mock.go -package=mock
type Repository interface {
	WithTx(tx *sql.Tx) Repository
	Create(ctx context.Context, dept *Department) error
	FindAllByCompany(ctx context.Context, companyID string) ([]Department, error)
	FindByIDAndCompany(ctx context.Context, companyID string, id string) (*Department, error)
	Update(ctx context.Context, dept *Department) error
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

func (r *repository) Create(ctx context.Context, dept *Department) error {
	return r.db.WithContext(ctx).Create(dept).Error
}

func (r *repository) FindAllByCompany(ctx context.Context, companyID string) ([]Department, error) {
	var depts []Department
	err := r.db.WithContext(ctx).
		Scopes(tenant.Scope(companyID)).
		Find(&depts).Error
	return depts, err
}

func (r *repository) FindByIDAndCompany(ctx context.Context, id string, companyID string) (*Department, error) {
	var dept Department
	err := r.db.WithContext(ctx).
		Scopes(tenant.Scope(companyID)).
		First(&dept).Error
	return &dept, err
}

func (r *repository) Update(ctx context.Context, dept *Department) error {
	return r.db.WithContext(ctx).Save(dept).Error
}

func (r *repository) Delete(ctx context.Context, id string, companyID string) error {
	return r.db.WithContext(ctx).
		Scopes(tenant.Scope(companyID)).
		Delete(&Department{}).Error
}
